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

package gauge // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/gauge"

import (
	"sync"
	"sync/atomic"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
)

// Note: this Gauge aggregator is designed to be used as a synchronous
// aggregator.  For this, it captures monotonic sequence counter with
// each Update, allowing the collection path to detect an unused
// instrument as distinct from an unchanged value.  The atomic sequence
// is unnecessary overhead for the async path.

type (
	Methods[N number.Any, Traits number.Traits[N], Storage State[N, Traits]] struct{}

	State[N number.Any, Traits number.Traits[N]] struct {
		lock  sync.Mutex
		value N
		seq   uint64
	}

	Int64   = State[int64, number.Int64Traits]
	Float64 = State[float64, number.Float64Traits]

	Int64Methods   = Methods[int64, number.Int64Traits, Int64]
	Float64Methods = Methods[float64, number.Float64Traits, Float64]
)

const initialSequence uint64 = 1

var (
	// The zero sequence value indicates an unset gauge
	sequenceVar uint64 = initialSequence

	_ aggregator.Methods[int64, Int64]     = Int64Methods{}
	_ aggregator.Methods[float64, Float64] = Float64Methods{}

	_ aggregation.Gauge = &Int64{}
	_ aggregation.Gauge = &Float64{}
)

func NewInt64(x int64) *Int64 {
	return &Int64{
		value: x,
		seq:   initialSequence,
	}
}

func NewFloat64(x float64) *Float64 {
	return &Float64{
		value: x,
		seq:   initialSequence,
	}
}

func (g *State[N, Traits]) Gauge() (number.Number, bool) {
	var t Traits
	return t.ToNumber(g.value), g.seq != 0
}

func (g *State[N, Traits]) Kind() aggregation.Kind {
	return aggregation.GaugeKind
}

// SetSequenceForTesting sets the Gauge to match one of the test
// gauges so far as its sequence number, allowing it to match exactly
// in tests.
func (g *State[N, Traits]) SetSequenceForTesting() {
	g.seq = initialSequence
}

func (Methods[N, Traits, Storage]) Kind() aggregation.Kind {
	return aggregation.GaugeKind
}

func (Methods[N, Traits, Storage]) Init(state *State[N, Traits], _ aggregator.Config) {
	// Note: storage is zero to start
}

func (Methods[N, Traits, Storage]) HasChange(ptr *State[N, Traits]) bool {
	return ptr.seq != 0
}

func (Methods[N, Traits, Storage]) Move(from, to *State[N, Traits]) {
	from.lock.Lock()
	defer from.lock.Unlock()

	to.value = from.value
	to.seq = from.seq

	from.seq = 0
}

func (Methods[N, Traits, Storage]) Copy(from, to *State[N, Traits]) {
	from.lock.Lock()
	defer from.lock.Unlock()
	to.value = from.value
	to.seq = from.seq
}

func (Methods[N, Traits, Storage]) Update(state *State[N, Traits], number N) {
	newSeq := atomic.AddUint64(&sequenceVar, 1)

	state.lock.Lock()
	defer state.lock.Unlock()

	state.value = number
	state.seq = newSeq
}

func (Methods[N, Traits, Storage]) Merge(from, to *State[N, Traits]) {
	to.lock.Lock()
	defer to.lock.Unlock()

	if from.seq != 0 && from.seq > to.seq {
		to.value = from.value
		to.seq = from.seq
	}
}

func (Methods[N, Traits, Storage]) ToAggregation(state *State[N, Traits]) aggregation.Aggregation {
	return state
}

func (Methods[N, Traits, Storage]) ToStorage(aggr aggregation.Aggregation) (*State[N, Traits], bool) {
	r, ok := aggr.(*State[N, Traits])
	return r, ok
}

func (Methods[N, Traits, Storage]) SubtractSwap(operand, argument *State[N, Traits]) {
	panic("not used for non-temporal metrics")
}
