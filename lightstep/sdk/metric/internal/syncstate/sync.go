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
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/fprint"
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

	// lock protects current.
	lock sync.RWMutex

	// current is protected by lock.
	current map[uint64]*record
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
		current:    map[uint64]*record{},

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
	inst.lock.Lock()
	defer inst.lock.Unlock()

	for key, reclist := range inst.current {
		// reclist is a list of records for this fingerprint.
		var head *record
		var tail *record

		// Scan reclist and modify the list. We're holding the
		// lock giving exclusive access to the head-of-list
		// and each next field, so the process here builds a new
		// linked list after filtering records that are no longer
		// in use.
		for rec := reclist; rec != nil; rec = rec.next {
			if inst.singleSnapshotAndProcess(key, rec) {
				if head == nil {
					// The first time a record will be kept,
					// it becomes the head and tail.
					head = rec
					tail = rec
				} else {
					// Subsequently, update the
					// tail of the list.
					tail.next = rec
					tail = rec
				}
			}
		}

		// When records are kept, delete the map entry.
		if head == nil {
			delete(inst.current, key)
			continue
		}

		// Otherwise, terminate the list that was built.
		tail.next = nil

		if head != reclist {
			// If the head changes, update the map.
			inst.current[key] = head
		}
	}
}

// singleSnapshotAndProcess
func (inst *Instrument) singleSnapshotAndProcess(fp uint64, rec *record) bool {
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

	// When `unmapped` is true, any other goroutines are now
	// trying to re-insert this entry in the map, they are busy
	// calling Gosched() waiting for this record to disappear.
	return !unmapped
}

// record consists of an accumulator, a reference count, the number of
// updates, and the number of collected updates.
type record struct {
	// refMapped tracks concurrent references to the record in
	// order to keep the record mapped as long as it is active or
	// uncollected.
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

	// attributeSet is ordered and deduplicated
	attributeSet attribute.Set

	// attributeList is in user-specified order, may contain duplicates.
	attributeList []attribute.KeyValue

	// next is protected by the instrument's RWLock.
	next *record
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

	rec := acquireRecord[N](inst, attrs)
	defer rec.refMapped.unref()

	rec.accumulator.(viewstate.Updater[N]).Update(num)

	// Record was modified.
	atomic.AddInt64(&rec.updateCount, 1)
}

func fingerprintAttributes(attrs []attribute.KeyValue) uint64 {
	var fp uint64
	for _, attr := range attrs {
		fp += fprint.Mix(
			fprint.FingerprintString(string(attr.Key)),
			fingerprintValue(attr.Value),
		)
	}

	return fp
}

func fingerprintSlice[T any](slice []T, f func(T) uint64) uint64 {
	var fp uint64
	for _, item := range slice {
		fp += f(item)
	}
	return fp
}

func fingerprintValue(value attribute.Value) uint64 {
	switch value.Type() {
	case attribute.BOOL:
		return fprint.FingerprintBool(value.AsBool())
	case attribute.INT64:
		return fprint.FingerprintInt64(value.AsInt64())
	case attribute.FLOAT64:
		return fprint.FingerprintFloat64(value.AsFloat64())
	case attribute.STRING:
		return fprint.FingerprintString(value.AsString())
	case attribute.BOOLSLICE:
		return fingerprintSlice(value.AsBoolSlice(), fprint.FingerprintBool)
	case attribute.INT64SLICE:
		return fingerprintSlice(value.AsInt64Slice(), fprint.FingerprintInt64)
	case attribute.FLOAT64SLICE:
		return fingerprintSlice(value.AsFloat64Slice(), fprint.FingerprintFloat64)
	case attribute.STRINGSLICE:
		return fingerprintSlice(value.AsStringSlice(), fprint.FingerprintString)
	}

	return 0
}

func sliceEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// attributesEqual returns true if two slices are exactly equal.
func attributesEqual(a, b []attribute.KeyValue) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Key != b[i].Key {
			return false
		}
		switch a[i].Value.Type() {
		case attribute.INVALID, attribute.BOOL, attribute.INT64, attribute.FLOAT64, attribute.STRING:
			if a[i].Value != b[i].Value {
				return false
			}
		case attribute.BOOLSLICE:
			if !sliceEqual(a[i].Value.AsBoolSlice(), b[i].Value.AsBoolSlice()) {
				return false
			}
		case attribute.INT64SLICE:
			if !sliceEqual(a[i].Value.AsInt64Slice(), b[i].Value.AsInt64Slice()) {
				return false
			}
		case attribute.FLOAT64SLICE:
			if !sliceEqual(a[i].Value.AsFloat64Slice(), b[i].Value.AsFloat64Slice()) {
				return false
			}
		case attribute.STRINGSLICE:
			if !sliceEqual(a[i].Value.AsStringSlice(), b[i].Value.AsStringSlice()) {
				return false
			}
		}

	}
	return true
}

// acquireRead acquires the read lock and searches for a `*record`.
func acquireRead(inst *Instrument, fp uint64, attrs []attribute.KeyValue) *record {
	inst.lock.RLock()
	defer inst.lock.RUnlock()

	rec := inst.current[fp]

	// Note: we could (optionally) allow collisions and not scan this list.
	for rec != nil && !attributesEqual(attrs, rec.attributeList) {
		rec = rec.next
	}

	// Existing record case.
	if rec != nil && rec.refMapped.ref() {
		// At this moment it is guaranteed that the
		// record is in the map and will not be removed.
		return rec
	}

	return nil
}

// acquireRecord gets or creates a `*record` corresponding to `attrs`,
// the input attributes.
func acquireRecord[N number.Any](inst *Instrument, attrs []attribute.KeyValue) *record {
	fp := fingerprintAttributes(attrs)

	rec := acquireRead(inst, fp, attrs)
	if rec != nil {
		return rec
	}

	// Build the attribute set.
	acpy := make([]attribute.KeyValue, len(attrs))
	copy(acpy, attrs)
	aset := attribute.NewSet(acpy...)

	// Note: the accumulator set below is created speculatively;
	// it will be released if it is never returned.
	newRec := &record{
		refMapped:     newRefcountMapped(),
		accumulator:   inst.compiled.NewAccumulator(aset),
		attributeList: attrs,
		attributeSet:  aset,
	}

	for {
		acquired, loaded := acquireWrite(inst, fp, newRec)

		if !loaded {
			// When this happens, we are waiting for the call to delete()
			// inside SnapshotAndProcess() to complete before inserting
			// a new record.  This avoids busy-waiting.
			runtime.Gosched()
			continue
		}

		if acquired != newRec {
			// Release the speculative accumulator, since it was not used.
			newRec.accumulator.SnapshotAndProcess(true)
		}
		return acquired
	}
}

// acquireWrite acquires the write lock and gets or sets a `*record`.
func acquireWrite(inst *Instrument, fp uint64, newRec *record) (*record, bool) {
	inst.lock.Lock()
	defer inst.lock.Unlock()

	for oldRec := inst.current[fp]; oldRec != nil; oldRec = oldRec.next {

		if attributesEqual(oldRec.attributeList, newRec.attributeList) {
			if oldRec.refMapped.ref() {
				return oldRec, true
			}
			// in which case, there's been a race
			return nil, false
		}
	}

	newRec.next = inst.current[fp]
	inst.current[fp] = newRec
	return newRec, true
}
