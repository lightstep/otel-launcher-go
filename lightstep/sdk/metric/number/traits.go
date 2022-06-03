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

package number // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"

import (
	"math"
	"sync/atomic"
	"unsafe"
)

// Traits is the generic traits interface for numbers used in the SDK.
type Traits[N Any] interface {
	// FromNumber turns a generic 64bits into the correct machine type.
	FromNumber(Number) N

	// ToNumber turns this type into a generic 64bit value.
	ToNumber(value N) Number

	// SetAtomic sets `ptr` to `value`.
	SetAtomic(ptr *N, value N)

	// GetAtomic reads `ptr`.
	GetAtomic(ptr *N) N

	// AddAtomic sets `ptr` to `value+*ptr`.
	AddAtomic(ptr *N, value N)

	// AddAtomic sets `ptr` to `value` and returns the former value.
	SwapAtomic(ptr *N, value N) N

	// IsNaN indicates whether `math.IsNaN()` is true (impossible for int64).
	IsNaN(value N) bool

	// IsInf indicates whether `math.IsInf()` is true (impossible for int64).
	IsInf(value N) bool

	// Kind of
	Kind() Kind
}

// Int64Traits implements Traits[int64].
type Int64Traits struct{}

var _ Traits[int64] = Int64Traits{}

func (Int64Traits) ToNumber(x int64) Number {
	return Number(x)
}

func (Int64Traits) FromNumber(n Number) int64 {
	return int64(n)
}

func (Int64Traits) GetAtomic(ptr *int64) int64 {
	return atomic.LoadInt64(ptr)
}

func (Int64Traits) SetAtomic(ptr *int64, value int64) {
	atomic.StoreInt64(ptr, value)
}

func (Int64Traits) SwapAtomic(ptr *int64, value int64) int64 {
	return atomic.SwapInt64(ptr, value)
}

func (Int64Traits) AddAtomic(ptr *int64, value int64) {
	atomic.AddInt64(ptr, value)
}

func (Int64Traits) IsNaN(_ int64) bool {
	return false
}

func (Int64Traits) IsInf(_ int64) bool {
	return false
}

func (Int64Traits) Kind() Kind {
	return Int64Kind
}

// Float64Traits implements Traits[float64].
type Float64Traits struct{}

var _ Traits[float64] = Float64Traits{}

func (Float64Traits) ToNumber(x float64) Number {
	return Number(math.Float64bits(x))
}

func (Float64Traits) FromNumber(n Number) float64 {
	return math.Float64frombits(uint64(n))
}

func (Float64Traits) GetAtomic(ptr *float64) float64 {
	return math.Float64frombits(atomic.LoadUint64((*uint64)(unsafe.Pointer(ptr))))
}

func (Float64Traits) SetAtomic(ptr *float64, value float64) {
	atomic.StoreUint64((*uint64)(unsafe.Pointer(ptr)), math.Float64bits(value))
}

func (Float64Traits) SwapAtomic(ptr *float64, value float64) float64 {
	return math.Float64frombits(atomic.SwapUint64((*uint64)(unsafe.Pointer(ptr)), math.Float64bits(value)))
}

func (Float64Traits) AddAtomic(ptr *float64, value float64) {
	for {
		oldBits := atomic.LoadUint64((*uint64)(unsafe.Pointer(ptr)))
		sum := math.Float64frombits(oldBits) + value
		newBits := math.Float64bits(sum)

		if atomic.CompareAndSwapUint64((*uint64)(unsafe.Pointer(ptr)), oldBits, newBits) {
			return
		}
	}
}

func (Float64Traits) IsNaN(value float64) bool {
	return math.IsNaN(value)
}

func (Float64Traits) IsInf(value float64) bool {
	return math.IsInf(value, 0)
}

func (Float64Traits) Kind() Kind {
	return Float64Kind
}
