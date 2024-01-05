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
	"sync"
	"sync/atomic"
	"time"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/pipeline"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var overflowAttributesFingerprint = fingerprintAttributes(pipeline.OverflowAttributes)

// Instrument maintains a mapping from attribute.Set to an internal
// record type for a single API-level instrument.  This type is
// organized so that a single attribute.Set lookup is performed
// regardless of the number of reader and instrument-view behaviors.
// Entries in the map have their accumulator's SnapshotAndProcess()
// method called whenever they are removed from the map, which can
// happen when any reader collects the instrument.
type Observer struct {
	// descriptor is the API-provided descriptor for the
	// instrument, unmodified by views.
	descriptor sdkinstrument.Descriptor

	// performance settings for the instrument.
	performance sdkinstrument.Performance

	// compiled will be a single compiled instrument or a
	// multi-instrument in case of multiple view behaviors
	// and/or readers; these distinctions do not matter
	// for synchronous aggregation.
	compiled viewstate.Instrument

	// lock protects current.
	lock sync.RWMutex

	// currentFP is protected by lock.
	currentFP map[uint64]*recordKV
}

// New builds a new synchronous instrument *Observer given the
// per-pipeline instrument-views compiled.  Note that the unused
// third parameter is an opaque value used in the asyncstate package,
// passed here to make these two packages generalize.
func New(desc sdkinstrument.Descriptor, performance sdkinstrument.Performance, _ interface{}, compiled pipeline.Register[viewstate.Instrument]) *Observer {
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
	// validate so that 0 is replaced w/ a better default.
	performance = performance.Validate()
	return &Observer{
		descriptor:  desc,
		currentFP:   map[uint64]*recordKV{},
		performance: performance,

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
func (inst *Observer) SnapshotAndProcess() {
	inst.lock.Lock()
	defer inst.lock.Unlock()

	for key, reclist := range inst.currentFP {
		// reclist is a list of records for this fingerprint.
		var head *recordKV
		var tail *recordKV

		// Scan reclist and modify the list. We're holding the
		// lock giving exclusive access to the head-of-list
		// and each next field, so the process here builds a new
		// linked list after filtering records that are no longer
		// in use.
		for rec := reclist; rec != nil; rec = rec.next {
			if inst.collect(key, rec) {
				if head == nil {
					// The first time a record will be kept,
					// it becomes the head and tail.
					head = rec
					tail = rec
				} else {
					// Subsequently, update the tail of the
					// list.  Note that this creates a
					// temporarily invalid list will be
					// repaired outside the loop, below.
					tail.next = rec
					tail = rec
				}
			}
		}

		// When no records are kept, delete the map entry.
		if head == nil {
			delete(inst.currentFP, key)
			continue
		}

		// Otherwise, terminate the list that was built.
		tail.next = nil

		if head != reclist {
			// If the head changes, update the map.
			inst.currentFP[key] = head
		}
	}
}

// collect collects the record.  When the record has been inactive for
// the configured number of periods, it is removed from memory.
func (inst *Observer) collect(fp uint64, rec *recordKV) bool {
	if rec.normalCollect() {
		rec.inactiveCount = 0
		return true
	}

	// Allow the record to remain in an inactive state for a number
	// of collections.
	rec.inactiveCount++

	if rec.inactiveCount < inst.performance.InactiveCollectionPeriods {
		return true
	}

	// Having no updates since last collection, try to unmap:
	unmapped := rec.refMapped.tryUnmap()

	// normalCollect returned false indicating no change, except:
	// (a) it's now possible there was a race, the collector needs to see it.
	// (b) if this is indeed the last reference, the collector needs the release signal.
	_ = rec.scavengeCollect(unmapped)

	// reset inactivity in case of !unmapped.
	rec.inactiveCount = 0

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

	// updateCount is incremented on every Update.  This field
	// uses atomic load/store operations.
	updateCount uint32

	// collectedCount is set to updateCount on collection,
	// supports checking for no updates during a round.  This
	// field is read/written under the Observer lock.
	collectedCount uint32

	// inactiveCount is incremented when the record is collected
	// but has not been updated.  This is used to compare against
	// the Performance InactiveCollectionPeriods setting.  This
	// field is read/written under the Observer lock.
	inactiveCount uint32

	// inst allows referring to performance settings.
	inst *Observer
}

type recordKV struct {
	// record includes fields in common with recordAS
	record

	// next is protected by the instrument's RWLock.
	//
	// this field is unused when Performance.IgnoreCollisions is true.
	next *recordKV

	// once governs access to `accumulatorsUnsafe`.  The caller
	// that created the `record` must call once.Do(initialize) on
	// its own code path, although another goroutine might
	// actually perform the initialization.  This is arranged with
	// the use of readAccumulator().
	once sync.Once

	// accumulatorUnsafe can be a multi-accumulator if there
	// are multiple behaviors or multiple readers, but
	// these distinctions are not relevant for synchronous
	// instruments.
	//
	// Note: use record.readAccumulator() to access this value,
	// to ensure that once.Do(initialize) is called.
	accumulatorUnsafe viewstate.Accumulator

	// attrsList is the caller's copy of the attribute list.  when
	// IgnoreCollisions is false, this is copied immediately as we
	// need it for reference after the call returns.  In both cases,
	// when the entry is not found and a new record is initialized,
	// a new copy of the attribute list will be created for the
	// call to NewSet.
	attrsList []attribute.KeyValue
}

// normalCollect equals conditionalCollect(false), is named
// differently from scavengeCollect for profiling.
func (rec *recordKV) normalCollect() bool {
	return rec.conditionalCollect(false)
}

// scavengeCollect equals conditionalCollect(false), is named
// differently from normalCollect for profiling.
func (rec *recordKV) scavengeCollect(release bool) bool {
	return rec.conditionalCollect(release)
}

// conditionalCollect checks whether the accumulator has been modified
// since the last collection (by any reader), returns a boolean
// indicating whether the record is active.  If modified, calls
// SnapshotAndProcess on the associated accumulator and returns true.
// If updates happened since the last collection (by any reader),
// returns false.
func (rec *recordKV) conditionalCollect(release bool) bool {
	mods := atomic.LoadUint32(&rec.updateCount)

	if !release {
		if mods == rec.collectedCount {
			return false
		}
	}

	rec.readAccumulator().SnapshotAndProcess(release)

	// Updates happened in this interval, collect and continue.
	rec.collectedCount = mods
	return true
}

// readAccumulator gets the accumulator for this record after once.Do(initialize).
func (rec *recordKV) readAccumulator() viewstate.Accumulator {
	rec.once.Do(rec.initialize)
	return rec.accumulatorUnsafe
}

// initialize ensures that accumulatorUnsafe and attrsUnsafe are correctly initialized.
//
// readAccumulator() calls this inside a sync.Once.Do().
func (rec *recordKV) initialize() {
	// We need another copy of the attribute list because NewSet()
	// will sort it in place.
	acpy := make([]attribute.KeyValue, len(rec.attrsList))
	copy(acpy, rec.attrsList)

	// When ignoring collisions, the list is no longer used.
	if rec.inst.performance.IgnoreCollisions {
		rec.attrsList = nil
	}

	aset := attribute.NewSet(acpy...)
	rec.accumulatorUnsafe = rec.inst.compiled.NewAccumulator(aset)
}

// computeAttrsUnderLock sets the attribute.Set that will be used to
// construct the accumulator.
func (rec *recordKV) computeAttrsUnderLock(attrs []attribute.KeyValue) {
	// The work of NewSet and NewAccumulator is deferred until
	// once.Do(initialize) outside of the lock but while the call
	// is still in flight.
	if rec.inst.performance.IgnoreCollisions {
		rec.attrsList = attrs
		return
	}

	// We have to copy the attributes here because the caller may
	// modify their copy of the list after the call returns.
	acpy := make([]attribute.KeyValue, len(attrs))
	copy(acpy, attrs)
	rec.attrsList = acpy
}

// interface conversion.
type OpConfig struct {
	Attributes attribute.Set
	KeyValues  []attribute.KeyValue
}

func (inst *Observer) ObserveInt64(ctx context.Context, num int64, cfg OpConfig) {
	Observe[int64, number.Int64Traits](ctx, inst, num, cfg)
}

func (inst *Observer) ObserveFloat64(ctx context.Context, num float64, cfg OpConfig) {
	Observe[float64, number.Float64Traits](ctx, inst, num, cfg)
}

// Observe performs a generic update for any synchronous instrument.
func Observe[N number.Any, Traits number.Traits[N]](ctx context.Context, inst *Observer, num N, cfg OpConfig) {
	if inst == nil {
		// Instrument was completely disabled by the view.
		return
	}

	if !aggregator.RangeTest[N, Traits](num, inst.descriptor) {
		return
	}

	var keyValues []attribute.KeyValue
	if cfg.KeyValues != nil {
		keyValues = cfg.KeyValues
	} else {
		// TODO: This is a new code path for optimization,
		// for now fall back to the slow path.
		keyValues = cfg.Attributes.ToSlice()
	}

	keyValues = inst.performance.TruncateAttributes(keyValues)

	if inst.performance.MeasurementProcessor != nil {
		// This is the last time context can be used.
		keyValues = inst.performance.MeasurementProcessor.Process(ctx, keyValues)
	}
	rec := acquireUninitializedKV[N](inst, keyValues)

	defer rec.refMapped.unref()

	var tr Traits
	var exBits aggregator.ExemplarBits
	updater := rec.readAccumulator().(viewstate.Updater[N])

	// TODO: Note the isTraced() calculation here is difficult to
	// place.  It can be deferred until the filter is known, but
	// there could be more than one filter, in which case it will

	// MeasurementProcessor is non-nil, the context has already
	// been probed.  Assuming the context has already been probed
	// once, we should know by now whether the context is sampled.
	span := trace.SpanFromContext(ctx)
	isTraced := span.SpanContext().IsSampled()

	if updater.MaySample(isTraced) {
		exBits.Time = time.Now()
		exBits.Attributes = keyValues
		exBits.Span = span
		exBits.Number = tr.ToNumber(num)
	}

	updater.Update(num, exBits)

	// Record was modified.
	atomic.AddUint32(&rec.updateCount, 1)
}
