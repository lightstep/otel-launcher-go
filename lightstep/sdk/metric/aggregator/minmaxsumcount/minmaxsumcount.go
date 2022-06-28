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

package minmaxsumcount // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/minmaxsumcount"

import (
	"sync"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
)

// Note: there is no checking for negative inputs here.  We assume
// that the Histogram API contract is being met and that no negative
// values arrive, even though the data model allows for Sum to be
// absent when negative values are used.  IMO the Histogram data point
// should carry information about the nature of the underlying
// observation: a histogram of counter values has a monotonic sum; a
// histogram of UpDownCounter and Gauge values has a non-monotonic
// sum.

type (
	Methods[N number.Any, Traits number.Traits[N], Storage State[N, Traits]] struct{}

	fields[N number.Any, Traits number.Traits[N]] struct {
		min   N
		max   N
		sum   N
		count uint64
	}

	State[N number.Any, Traits number.Traits[N]] struct {
		lock sync.Mutex
		fields[N, Traits]
	}

	Int64   = State[int64, number.Int64Traits]
	Float64 = State[float64, number.Float64Traits]

	Int64Methods   = Methods[int64, number.Int64Traits, Int64]
	Float64Methods = Methods[float64, number.Float64Traits, Float64]
)

var (
	_ aggregator.Methods[int64, Int64]     = Int64Methods{}
	_ aggregator.Methods[float64, Float64] = Float64Methods{}

	_ aggregation.MinMaxSumCount = &Int64{}
	_ aggregation.MinMaxSumCount = &Float64{}
)

func NewInt64(vals ...int64) *Int64 {
	a := &Int64{}
	for _, val := range vals {
		Int64Methods{}.Update(a, val)
	}
	return a
}

func NewFloat64(vals ...float64) *Float64 {
	a := &Float64{}
	for _, val := range vals {
		Float64Methods{}.Update(a, val)
	}
	return a
}

func (g *State[N, Traits]) Sum() number.Number {
	var t Traits
	return t.ToNumber(g.sum)
}

func (g *State[N, Traits]) Count() uint64 {
	return g.count
}

func (g *State[N, Traits]) Min() number.Number {
	var t Traits
	return t.ToNumber(g.min)
}

func (g *State[N, Traits]) Max() number.Number {
	var t Traits
	return t.ToNumber(g.max)
}

func (g *State[N, Traits]) Kind() aggregation.Kind {
	return aggregation.MinMaxSumCountKind
}

func (Methods[N, Traits, Storage]) Kind() aggregation.Kind {
	return aggregation.MinMaxSumCountKind
}

func (Methods[N, Traits, Storage]) Init(state *State[N, Traits], _ aggregator.Config) {
}

func (Methods[N, Traits, Storage]) HasChange(ptr *State[N, Traits]) bool {
	return ptr.count != 0
}

func (Methods[N, Traits, Storage]) Move(from, to *State[N, Traits]) {
	from.lock.Lock()
	defer from.lock.Unlock()

	to.fields, from.fields = from.fields, fields[N, Traits]{}
}

func (Methods[N, Traits, Storage]) Copy(from, to *State[N, Traits]) {
	from.lock.Lock()
	defer from.lock.Unlock()

	to.fields = from.fields
}

func (Methods[N, Traits, Storage]) Update(state *State[N, Traits], number N) {
	state.lock.Lock()
	defer state.lock.Unlock()

	if state.count == 0 {
		state.min = number
		state.max = number
	} else {
		if number < state.min {
			state.min = number
		}
		if number > state.max {
			state.max = number
		}
	}

	state.sum += number
	state.count++
}

func (Methods[N, Traits, Storage]) Merge(from, to *State[N, Traits]) {
	to.lock.Lock()
	defer to.lock.Unlock()

	if from.fields.count != 0 {
		if to.fields.count == 0 {
			to.fields.min = from.fields.min
			to.fields.max = from.fields.max
		} else {
			if from.fields.min < to.fields.min {
				to.fields.min = from.fields.min
			}
			if from.fields.max > to.fields.max {
				to.fields.max = from.fields.max
			}
		}
	}

	to.fields.sum += from.fields.sum
	to.fields.count += from.fields.count
}

func (Methods[N, Traits, Storage]) ToAggregation(state *State[N, Traits]) aggregation.Aggregation {
	return state
}

func (Methods[N, Traits, Storage]) ToStorage(aggr aggregation.Aggregation) (*State[N, Traits], bool) {
	r, ok := aggr.(*State[N, Traits])
	return r, ok
}

func (Methods[N, Traits, Storage]) SubtractSwap(operand, argument *State[N, Traits]) {
	// This can't be called b/c histogram's are only used with synchronous instruments,
	// which start as delta temporality and thus never subtract.
	panic("impossible call")
}
