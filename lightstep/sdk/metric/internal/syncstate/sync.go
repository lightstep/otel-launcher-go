// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syncstate

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/pipeline"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel/attribute"
)

// Instrument maintains a mapping from attribute.Set to an internal
// record type for a single API-level instrument.  This type is
// organized so that a single attribute.Set lookup is performed
// regardless of the number of reader and instrument-view behaviors.
// Entries in the map have their accumulator's SnapshotAndProcess()
// method called whenever they are removed from the map, which can
// happen when any reader collects the instrument.
type Instrument struct {
	// descriptor is the API-provided descriptor for the
	// instrument, unmodified by views.
	descriptor sdkinstrument.Descriptor

	// compiled will be a single compiled instrument or a
	// multi-instrument in case of multiple view behaviors
	// and/or readers; these distinctions do not matter
	// for synchronous aggregation.
	compiled viewstate.Instrument

	// current is a synchronous form of map[attribute.Set]*record.
	current sync.Map
}

// NewInstruments builds a new synchronous instrument given the
// per-pipeline instrument-views compiled.  Note that the unused
// second parameter is an opaque value used in the asyncstate package,
// passed here to make these two packages generalize.
func NewInstrument(desc sdkinstrument.Descriptor, _ interface{}, compiled pipeline.Register[viewstate.Instrument]) *Instrument {
	var nonnil []viewstate.Instrument
	for _, comp := range compiled {
		if comp != nil {
			nonnil = append(nonnil, comp)
		}
	}
	if nonnil == nil {
		// When no readers enable the instrument, no need for an instrument.
		return nil
	}
	return &Instrument{
		descriptor: desc,

		// Note that viewstate.Combine is used to eliminate
		// the per-pipeline distinction that is useful in the
		// asyncstate package.  Here, in the common case there
		// will be one pipeline and one view, such that
		// viewstate.Combine produces a single concrete
		// viewstate.Instrument.  Only when there are multiple
		// views or multiple pipelines will the combination
		// produce a viewstate.multiInstrument here.
		compiled: viewstate.Combine(desc, nonnil...),
	}
}

// SnapshotAndProcess calls SnapshotAndProcess() for all live
// accumulators of this instrument.  Inactive accumulators will be
// subsequently removed from the map.
func (inst *Instrument) SnapshotAndProcess() {
	inst.current.Range(func(key interface{}, value interface{}) bool {
		rec := value.(*record)
		if rec.conditionalSnapshotAndProcess(false) {
			return true
		}
		// Having no updates since last collection, try to unmap:
		unmapped := rec.refMapped.tryUnmap()

		// The first rec.conditionalSnapshotAndProcess
		// returned false indicating no change, except:
		// (a) it's now possible there was a race, the collector needs to
		// see it.
		// (b) if this is indeed the last reference, the collector needs the
		// release signal.
		_ = rec.conditionalSnapshotAndProcess(unmapped)

		if unmapped {
			// If any other goroutines are now trying to re-insert this
			// entry in the map, they are busy calling Gosched() awaiting
			// this deletion:
			inst.current.Delete(key)
		}

		return true
	})
}

// record consists of an accumulator, a reference count, the number of
// updates, and the number of collected updates.
type record struct {
	refMapped refcountMapped

	// updateCount is incremented on every Update.
	updateCount int64

	// collectedCount is set to updateCount on collection,
	// supports checking for no updates during a round.
	collectedCount int64

	// accumulator can be a multi-accumulator if there
	// are multiple behaviors or multiple readers, but
	// these distinctions are not relevant for synchronous
	// instruments.
	accumulator viewstate.Accumulator
}

// conditionalSnapshotAndProcess checks whether the accumulator has been
// modified since the last collection (by any reader), returns a
// boolean indicating whether the record is active.  If active, calls
// SnapshotAndProcess on the associated accumulator and returns true.
// If updates happened since the last collection (by any reader),
// returns false.
func (rec *record) conditionalSnapshotAndProcess(release bool) bool {
	mods := atomic.LoadInt64(&rec.updateCount)

	if !release {
		if mods == atomic.LoadInt64(&rec.collectedCount) {
			return false
		}
	}

	rec.accumulator.SnapshotAndProcess(release)

	// Updates happened in this interval, collect and continue.
	atomic.StoreInt64(&rec.collectedCount, mods)
	return true
}

// capture performs a single update for any synchronous instrument.
func capture[N number.Any, Traits number.Traits[N]](_ context.Context, inst *Instrument, num N, attrs []attribute.KeyValue) {
	if inst == nil {
		// Instrument was completely disabled by the view.
		return
	}

	// Note: Here, this is the place to use context, e.g., extract baggage.

	if !aggregator.RangeTest[N, Traits](num, inst.descriptor) {
		return
	}

	rec, updater := acquireRecord[N](inst, attrs)
	defer rec.refMapped.unref()

	updater.Update(num)

	// Record was modified.
	atomic.AddInt64(&rec.updateCount, 1)
}

// acquireRecord gets or creates a `*record` corresponding to `attrs`,
// the input attributes.
func acquireRecord[N number.Any](inst *Instrument, attrs []attribute.KeyValue) (*record, viewstate.Updater[N]) {
	aset := attribute.NewSet(attrs...)

	if lookup, ok := inst.current.Load(aset); ok {
		// Existing record case.
		rec := lookup.(*record)

		if rec.refMapped.ref() {
			// At this moment it is guaranteed that the
			// record is in the map and will not be removed.
			return rec, rec.accumulator.(viewstate.Updater[N])
		}
		// This record is no longer mapped, try to add a new
		// record below.
	}

	// Note: the accumulator set below is created speculatively;
	// if it is never returned, it will not be updated and can be
	// safely discarded.
	newRec := &record{
		refMapped:   newRefcountMapped(),
		accumulator: inst.compiled.NewAccumulator(aset),
	}

	for {
		if found, loaded := inst.current.LoadOrStore(aset, newRec); loaded {
			oldRec := found.(*record)
			if oldRec.refMapped.ref() {
				return oldRec, oldRec.accumulator.(viewstate.Updater[N])
			}
			// When this happens, we are waiting for the call to Delete()
			// inside SnapshotAndProcess() to complete before inserting
			// a new record.  This avoids busy-waiting.
			runtime.Gosched()
			continue
		}
		break
	}

	return newRec, newRec.accumulator.(viewstate.Updater[N])
}
