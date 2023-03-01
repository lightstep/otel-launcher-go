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
	"go.opentelemetry.io/otel/metric/instrument"
)

// Instrument maintains a mapping from attribute.Set to an internal
// record type for a single API-level instrument.  This type is
// organized so that a single attribute.Set lookup is performed
// regardless of the number of reader and instrument-view behaviors.
// Entries in the map have their accumulator's SnapshotAndProcess()
// method called whenever they are removed from the map, which can
// happen when any reader collects the instrument.
type Observer struct {
	instrument.Synchronous

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

	// current is protected by lock.
	current map[uint64]*record
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
		current:     map[uint64]*record{},
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

// collect collects the record.  When the record has been inactive for
// the configured number of periods, it is removed from memory.
func (inst *Observer) collect(fp uint64, rec *record) bool {
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

	// next is protected by the instrument's RWLock.
	//
	// this field is unused when Performance.IgnoreCollisions is true.
	next *record

	// once governs access to `accumulatorsUnsafe.  The caller
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

	// attrsListInputUnsafe is a temporary copy of the caller's
	// attribute list in its original order, possibly having duplicates.
	// this is passed to attribute.NewSetWithSortable, which performs a
	// stable-sort of the slice and resets the pointer to nil inside
	// once.Do(initialize).  This field is used consistently regardless
	// of IgnoreCollisions.
	attrsListInputUnsafe attribute.Sortable

	// attrsListCopy is a copy of the original attribute list, only
	// used when checking collisions (i.e., IgnoreCollisions is false).
	// set in computeAttrsUnderLock.
	attrsListCopy []attribute.KeyValue
}

// normalCollect equals conditionalCollect(false), is named
// differently from scavengeCollect for profiling.
func (rec *record) normalCollect() bool {
	return rec.conditionalCollect(false)
}

// scavengeCollect equals conditionalCollect(false), is named
// differently from normalCollect for profiling.
func (rec *record) scavengeCollect(release bool) bool {
	return rec.conditionalCollect(release)
}

// conditionalCollect checks whether the accumulator has been modified
// since the last collection (by any reader), returns a boolean
// indicating whether the record is active.  If modified, calls
// SnapshotAndProcess on the associated accumulator and returns true.
// If updates happened since the last collection (by any reader),
// returns false.
func (rec *record) conditionalCollect(release bool) bool {
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

// readAttributes gets the accumulator for this record after once.Do(initialize).
func (rec *record) readAccumulator() viewstate.Accumulator {
	rec.once.Do(rec.initialize)
	return rec.accumulatorUnsafe
}

// initialize ensures that accumulatorUnsafe and attrsUnsafe are correctly initialized.
//
// readAccumulator() calls this inside a sync.Once.Do().
func (rec *record) initialize() {
	// Note that rec.attrsListInputUnsafe is set to nil in NewSetWithSortable().
	aset := attribute.NewSetWithSortable(rec.attrsListInputUnsafe, &rec.attrsListInputUnsafe)

	rec.accumulatorUnsafe = rec.inst.compiled.NewAccumulator(aset)
}

// computeAttrsUnderLock sets the attribute.Set that will be used to
// construct the accumulator.
func (rec *record) computeAttrsUnderLock(attrs []attribute.KeyValue) {
	// The work of NewSetWithSortable and NewAccumulator is
	// deferred until once.Do(initialize) outside of the lock.
	rec.attrsListInputUnsafe = attrs

	if rec.inst.performance.IgnoreCollisions {
		return
	}

	acpy := make([]attribute.KeyValue, len(attrs))
	copy(acpy, attrs)
	rec.attrsListCopy = acpy
}

func (inst *Observer) ObserveInt64(ctx context.Context, num int64, attrs ...attribute.KeyValue) {
	Observe[int64, number.Int64Traits](ctx, inst, num, attrs...)
}

func (inst *Observer) ObserveFloat64(ctx context.Context, num float64, attrs ...attribute.KeyValue) {
	Observe[float64, number.Float64Traits](ctx, inst, num, attrs...)
}

// Observe performs a generic update for any synchronous instrument.
func Observe[N number.Any, Traits number.Traits[N]](_ context.Context, inst *Observer, num N, attrs ...attribute.KeyValue) {
	if inst == nil {
		// Instrument was completely disabled by the view.
		return
	}

	// Note: Here, this is the place to use context, e.g., extract baggage.

	if !aggregator.RangeTest[N, Traits](num, inst.descriptor) {
		return
	}

	rec := acquireUninitialized[N](inst, attrs)
	defer rec.refMapped.unref()

	rec.readAccumulator().(viewstate.Updater[N]).Update(num)

	// Record was modified.
	atomic.AddUint32(&rec.updateCount, 1)
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
		if a[i].Value.Type() != b[i].Value.Type() {
			return false
		}
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
func acquireRead(inst *Observer, fp uint64, attrs []attribute.KeyValue) *record {
	inst.lock.RLock()
	defer inst.lock.RUnlock()

	rec := inst.current[fp]

	// Potentially test for hash collisions.
	if !inst.performance.IgnoreCollisions {
		for rec != nil && !attributesEqual(attrs, rec.attrsListCopy) {
			rec = rec.next
		}
	}

	// Existing record case.
	if rec != nil && rec.refMapped.ref() {
		// At this moment it is guaranteed that the
		// record is in the map and will not be removed.
		return rec
	}

	return nil
}

// acquireUninitialized gets or creates a `*record` corresponding to
// `attrs`, the input attributes.  The returned record is mapped but
// possibly not initialized.
func acquireUninitialized[N number.Any](inst *Observer, attrs []attribute.KeyValue) *record {
	fp := fingerprintAttributes(attrs)

	if rec := acquireRead(inst, fp, attrs); rec != nil {
		return rec
	}

	return acquireNotfound[N](inst, fp, attrs)
}

// acquireNotfound is the code path taken when acquireRead does not
// locate a record.
func acquireNotfound[N number.Any](inst *Observer, fp uint64, attrs []attribute.KeyValue) *record {
	newRec := &record{
		inst:      inst,
		refMapped: newRefcountMapped(),
	}
	newRec.computeAttrsUnderLock(attrs)

	for {
		acquired, loaded := acquireWrite(inst, fp, newRec)

		if !loaded {
			// When this happens, we are waiting for the call to delete()
			// inside SnapshotAndProcess() to complete before inserting
			// a new record.  This avoids busy-waiting.
			runtime.Gosched()
			continue
		}

		return acquired
	}
}

// acquireWrite acquires the write lock and gets or sets a `*record`.
func acquireWrite(inst *Observer, fp uint64, newRec *record) (*record, bool) {
	inst.lock.Lock()
	defer inst.lock.Unlock()

	for oldRec := inst.current[fp]; oldRec != nil; oldRec = oldRec.next {

		if inst.performance.IgnoreCollisions || attributesEqual(oldRec.attrsListCopy, newRec.attrsListCopy) {
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
