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
	"runtime"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/fprint"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/pipeline"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"go.opentelemetry.io/otel/attribute"
)

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

// acquireRead acquires the lock and searches for a `*record`.
// This returns the overflow attributes and fingerprint in case the
// the cardinality limit is reached.  The caller should exchange their
// fp and attrs for the ones returned by this call.
func acquireReadKV(inst *Observer, fp uint64, attrs []attribute.KeyValue) (uint64, []attribute.KeyValue, *recordKV) {
	inst.lock.RLock()
	defer inst.lock.RUnlock()

	overflow := false
	fp, attrs, rec := acquireReadLockedKV(inst, fp, attrs, &overflow)

	if rec != nil {
		return fp, attrs, rec
	}
	// The overflow signal indicates another call is needed w/ the
	// same logic but updated fp and attrs.
	if !overflow {
		// Otherwise, this is the first appearance of an overflow.
		return fp, attrs, nil
	}
	// In which case fp and attrs are now the overflow attributes.
	return acquireReadLockedKV(inst, fp, attrs, &overflow)
}

func acquireReadLockedKV(inst *Observer, fp uint64, attrs []attribute.KeyValue, overflow *bool) (uint64, []attribute.KeyValue, *recordKV) {
	rec := inst.currentFP[fp]

	// Potentially test for hash collisions.
	if !inst.performance.IgnoreCollisions {
		for rec != nil && !attributesEqual(attrs, rec.attrsList) {
			rec = rec.next
		}
	}

	// Existing record case.
	if rec != nil && rec.refMapped.ref() {
		// At this moment it is guaranteed that the
		// record is in the map and will not be removed.
		return fp, attrs, rec
	}

	// Check for overflow after checking for the original
	// attribute set.  Note this means we are performing
	// two map lookups for overflowing attributes and only
	// one lookup if the attribute set was preexisting.
	if !*overflow && uint32(len(inst.currentFP)) >= inst.performance.InstrumentCardinalityLimit-1 {
		// Use the overflow attributes, repeat.
		attrs = pipeline.OverflowAttributes
		fp = overflowAttributesFingerprint
		*overflow = true
	}

	return fp, attrs, nil
}

// acquireUninitializedKV gets or creates a `*record` corresponding to
// `attrs`, the input attributes.  The returned record is mapped but
// possibly not initialized.
func acquireUninitializedKV[N number.Any](inst *Observer, attrs []attribute.KeyValue) *recordKV {
	fp := fingerprintAttributes(attrs)

	// acquireRead may replace fp and attrs when there is overflow.
	var rec *recordKV
	fp, attrs, rec = acquireReadKV(inst, fp, attrs)
	if rec != nil {
		return rec
	}

	return acquireNotfoundKV[N](inst, fp, attrs)
}

// acquireNotfoundKV is the code path taken when acquireRead does not
// locate a record.
func acquireNotfoundKV[N number.Any](inst *Observer, fp uint64, attrs []attribute.KeyValue) *recordKV {
	newRec := &recordKV{
		record: record{
			inst:      inst,
			refMapped: newRefcountMapped(),
		},
	}
	newRec.computeAttrsUnderLock(attrs)

	for {
		acquired, loaded := acquireWriteKV(inst, fp, newRec)

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

// acquireWriteKV acquires the write lock and gets or sets a `*record`.
func acquireWriteKV(inst *Observer, fp uint64, newRec *recordKV) (*recordKV, bool) {
	inst.lock.Lock()
	defer inst.lock.Unlock()

	for oldRec := inst.currentFP[fp]; oldRec != nil; oldRec = oldRec.next {

		if inst.performance.IgnoreCollisions || attributesEqual(oldRec.attrsList, newRec.attrsList) {
			if oldRec.refMapped.ref() {
				return oldRec, true
			}
			// in which case, there's been a race
			return nil, false
		}
	}

	newRec.next = inst.currentFP[fp]
	inst.currentFP[fp] = newRec
	return newRec, true
}
