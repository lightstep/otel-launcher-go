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

package sum // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/sum"

import (
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
)

type (
	Properties interface {
		kind() aggregation.Kind
		temporal() bool
	}

	MonotonicTemporal    struct{}
	NonMonotonicTemporal struct{}
	NonMonotonicSpatial  struct{}

	Methods[N number.Any, Traits number.Traits[N], Props Properties, Storage State[N, Traits, Props]] struct{}

	State[N number.Any, Traits number.Traits[N], Props Properties] struct {
		value N
	}

	MonotonicInt64           = State[int64, number.Int64Traits, MonotonicTemporal]
	NonMonotonicInt64        = State[int64, number.Int64Traits, NonMonotonicTemporal]
	NonMonotonicSpatialInt64 = State[int64, number.Int64Traits, NonMonotonicSpatial]

	MonotonicFloat64           = State[float64, number.Float64Traits, MonotonicTemporal]
	NonMonotonicFloat64        = State[float64, number.Float64Traits, NonMonotonicTemporal]
	NonMonotonicSpatialFloat64 = State[float64, number.Float64Traits, NonMonotonicSpatial]

	MonotonicInt64Methods           = Methods[int64, number.Int64Traits, MonotonicTemporal, MonotonicInt64]
	NonMonotonicInt64Methods        = Methods[int64, number.Int64Traits, NonMonotonicTemporal, NonMonotonicInt64]
	NonMonotonicSpatialInt64Methods = Methods[int64, number.Int64Traits, NonMonotonicSpatial, NonMonotonicSpatialInt64]

	MonotonicFloat64Methods           = Methods[float64, number.Float64Traits, MonotonicTemporal, MonotonicFloat64]
	NonMonotonicFloat64Methods        = Methods[float64, number.Float64Traits, NonMonotonicTemporal, NonMonotonicFloat64]
	NonMonotonicSpatialFloat64Methods = Methods[float64, number.Float64Traits, NonMonotonicSpatial, NonMonotonicSpatialFloat64]
)

func NewMonotonicInt64(x int64) *MonotonicInt64 {
	return &MonotonicInt64{value: x}
}

func NewNonMonotonicInt64(x int64) *NonMonotonicInt64 {
	return &NonMonotonicInt64{value: x}
}

func NewMonotonicFloat64(x float64) *MonotonicFloat64 {
	return &MonotonicFloat64{value: x}
}

func NewNonMonotonicFloat64(x float64) *NonMonotonicFloat64 {
	return &NonMonotonicFloat64{value: x}
}

func (MonotonicTemporal) kind() aggregation.Kind {
	return aggregation.MonotonicSumKind
}

func (NonMonotonicTemporal) kind() aggregation.Kind {
	return aggregation.NonMonotonicSumKind
}

func (NonMonotonicSpatial) kind() aggregation.Kind {
	return aggregation.LatestSumKind
}

func (MonotonicTemporal) temporal() bool {
	return true
}

func (NonMonotonicTemporal) temporal() bool {
	return true
}

func (NonMonotonicSpatial) temporal() bool {
	return false
}

var (
	_ aggregator.Methods[int64, MonotonicInt64]               = Methods[int64, number.Int64Traits, MonotonicTemporal, MonotonicInt64]{}
	_ aggregator.Methods[int64, NonMonotonicInt64]            = Methods[int64, number.Int64Traits, NonMonotonicTemporal, NonMonotonicInt64]{}
	_ aggregator.Methods[int64, NonMonotonicSpatialInt64]     = Methods[int64, number.Int64Traits, NonMonotonicSpatial, NonMonotonicSpatialInt64]{}
	_ aggregator.Methods[float64, MonotonicFloat64]           = Methods[float64, number.Float64Traits, MonotonicTemporal, MonotonicFloat64]{}
	_ aggregator.Methods[float64, NonMonotonicFloat64]        = Methods[float64, number.Float64Traits, NonMonotonicTemporal, NonMonotonicFloat64]{}
	_ aggregator.Methods[float64, NonMonotonicSpatialFloat64] = Methods[float64, number.Float64Traits, NonMonotonicSpatial, NonMonotonicSpatialFloat64]{}

	_ aggregation.Sum = &MonotonicInt64{}
	_ aggregation.Sum = &NonMonotonicInt64{}
	_ aggregation.Sum = &NonMonotonicSpatialInt64{}

	_ aggregation.Sum = &MonotonicFloat64{}
	_ aggregation.Sum = &NonMonotonicFloat64{}
	_ aggregation.Sum = &NonMonotonicSpatialFloat64{}
)

func (s *State[N, Traits, Props]) Sum() number.Number {
	var t Traits
	return t.ToNumber(s.value)
}

func (s *State[N, Traits, Props]) Kind() aggregation.Kind {
	var props Props
	return props.kind()
}

func (s *State[N, Traits, Props]) IsMonotonic() bool {
	var props Props
	return props.kind() == aggregation.MonotonicSumKind
}

func (Methods[N, Traits, Props, Storage]) Kind() aggregation.Kind {
	var props Props
	return props.kind()
}

func (Methods[N, Traits, Props, Storage]) Init(state *State[N, Traits, Props], _ aggregator.Config) {
	// Note: storage is zero to start
}

func (Methods[N, Traits, Props, Storage]) Move(from, to *State[N, Traits, Props]) {
	var t Traits
	to.value = t.SwapAtomic(&from.value, 0)
}

func (Methods[N, Traits, Props, Storage]) HasChange(ptr *State[N, Traits, Props]) bool {
	return ptr.value != 0
}

func (Methods[N, Traits, Props, Storage]) Update(state *State[N, Traits, Props], value N) {
	var props Props
	var t Traits
	if props.temporal() {
		t.AddAtomic(&state.value, value)
	} else {
		t.SetAtomic(&state.value, value)
	}
}

func (Methods[N, Traits, Props, Storage]) Copy(from, to *State[N, Traits, Props]) {
	var t Traits
	to.value = t.GetAtomic(&from.value)
}

func (Methods[N, Traits, Props, Storage]) Merge(from, to *State[N, Traits, Props]) {
	var t Traits
	t.AddAtomic(&to.value, from.value)
}

func (Methods[N, Traits, Props, Storage]) ToAggregation(state *State[N, Traits, Props]) aggregation.Aggregation {
	return state
}

func (Methods[N, Traits, Props, Storage]) ToStorage(aggr aggregation.Aggregation) (*State[N, Traits, Props], bool) {
	r, ok := aggr.(*State[N, Traits, Props])
	return r, ok
}

func (Methods[N, Traits, Props, Storage]) SubtractSwap(operand, argument *State[N, Traits, Props]) {
	operand.value = argument.value - operand.value
}
