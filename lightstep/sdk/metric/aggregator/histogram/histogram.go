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

package histogram // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram"

import (
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/sdk/metric/aggregator/exponential/mapping"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/exponential/mapping/exponent"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/exponential/mapping/logarithm"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
)

var ErrNoSubtract = fmt.Errorf("histogram subtract not implemented")

// Note: This code uses a Mutex to govern access to the exclusive
// aggregator state.  This is in contrast to a lock-free approach
// (as in the Go prometheus client) that was reverted here:
// https://github.com/open-telemetry/opentelemetry-go/pull/669

type (
	// State observes counts observations in exponentially-spaced
	// buckets.  It is configured with a maximum scale factor
	// which determines resolution.  Scale is automatically
	// adjusted to accommodate the range of input data.
	State[N int64 | float64] struct {
		lock    sync.Mutex
		maxSize int32
		data[N]
	}

	data[N int64 | float64] struct {
		// sum is the sum of all Updates reflected in the
		// aggregator.  It has the same type number as the
		// corresponding sdkinstrument.Descriptor.
		sum N
		// count is incremented by 1 per Update.
		count uint64
		// zeroCount is incremented by 1 when the measured
		// value is exactly 0.
		zeroCount uint64
		// min is set when count > 0
		min N
		// max is set when count > 0
		max N
		// positive holds the positive values
		positive buckets
		// negative holds the negative values in these buckets
		// by their absolute value.
		negative buckets
		// mapping corresponds to the current scale, is shared
		// by both positive and negative ranges.
		mapping mapping.Mapping
	}

	Methods[N number.Any, Storage State[N]] struct{}

	Int64   = State[int64]
	Float64 = State[float64]

	Int64Methods   = Methods[int64, Int64]
	Float64Methods = Methods[float64, Float64]
)

var (
	_ aggregator.Methods[int64, Int64]     = Int64Methods{}
	_ aggregator.Methods[float64, Float64] = Float64Methods{}

	_ aggregation.Histogram = asLightstepHistogram[int64, number.Int64Traits](&Int64{})
	_ aggregation.Histogram = asLightstepHistogram[float64, number.Float64Traits](&Float64{})
)

func asLightstepHistogram[N valueType, Traits number.Traits[N]](h *State[N]) aggregation.Histogram {
	// @@@
	return nil
}

func NewFloat64(cfg Config, values ...float64) *Float64 {
	return newHist[float64](cfg, values)
}

func NewInt64(cfg Config, values ...int64) *Int64 {
	return newHist[int64](cfg, values)
}

func newHist[N number.Any](cfg Config, values []N) *State[N] {
	var methods Methods[N, State[N]]

	state := &State[N]{}

	state.Init(cfg)

	for _, val := range values {
		methods.Update(state, val)
	}
	return state
}

func (h *State[N]) Init(cfg Config) {
	h.maxSize = cfg.maxSize

	mapping, _ := newMapping(logarithm.MaxScale)
	h.mapping = mapping
}

// Sum implements aggregation.Histogram.
func (h *State[N]) Sum() N {
	return h.sum
}

// Min implements aggregation.Histogram.
func (h *State[N]) Min() N {
	return h.min
}

// Max implements aggregation.Histogram.
func (h *State[N]) Max() N {
	return h.max
}

// Count implements aggregation.Histogram.
func (h *State[N]) Count() uint64 {
	return h.count
}

// Scale implements aggregation.Histogram.
func (h *State[N]) Scale() int32 {
	if h.data.count == h.data.zeroCount {
		// all zeros! scale doesn't matter, use zero.
		return 0
	}
	return h.data.mapping.Scale()
}

// ZeroCount implements aggregation.Histogram.
func (h *State[N]) ZeroCount() uint64 {
	return h.data.zeroCount
}

// Positive implements aggregation.Histogram.
func (h *State[N]) Positive() *buckets {
	return &h.data.positive
}

// Negative implements aggregation.Histogram.
func (h *State[N]) Negative() *buckets {
	return &h.data.negative
}

// Offset implements aggregation.Bucket.
func (b *buckets) Offset() int32 {
	return b.indexStart
}

// Len implements aggregation.Bucket.
func (b *buckets) Len() uint32 {
	if b.backing == nil {
		return 0
	}
	if b.indexEnd == b.indexStart && b.At(0) == 0 {
		return 0
	}
	return uint32(b.indexEnd - b.indexStart + 1)
}

// At returns the count of the bucket at a position in the logical
// array of counts.
func (b *buckets) At(pos0 uint32) uint64 {
	pos := pos0
	bias := uint32(b.indexBase - b.indexStart)

	if pos < bias {
		pos += uint32(b.backing.size())
	}
	pos -= bias

	return b.backing.countAt(pos)
}

// clearState resets a histogram to the empty state without changing
// backing array.  Scale is reset if there are no range limits.
func (h *State[N]) clearState() {
	h.positive.clearState()
	h.negative.clearState()
	h.sum = 0
	h.count = 0
	h.zeroCount = 0
	h.min = 0
	h.max = 0
	h.mapping, _ = newMapping(logarithm.MaxScale)
}

// clearState zeros the backing array.
func (b *buckets) clearState() {
	b.indexStart = 0
	b.indexEnd = 0
	b.indexBase = 0
	if b.backing != nil {
		b.backing.reset()
	}
}

func newMapping(scale int32) (mapping.Mapping, error) {
	if scale <= 0 {
		return exponent.NewMapping(scale)
	}
	return logarithm.NewMapping(scale)
}

func (h *State[N]) Kind() aggregation.Kind {
	return aggregation.HistogramKind
}

func (Methods[N, Storage]) Kind() aggregation.Kind {
	return aggregation.HistogramKind
}

func (Methods[N, Storage]) Init(state *State[N], cfg aggregator.Config) {
	state.Init(Config{
		maxSize: cfg.Histogram.MaxSize,
	})
}

func (Methods[N, Storage]) HasChange(ptr *State[N]) bool {
	return ptr.count != 0
}

func (Methods[N, Storage]) Move(from, to *State[N]) {
	from.Move(to)
}

func (Methods[N, Storage]) Copy(from, to *State[N]) {
	from.lock.Lock()
	defer from.lock.Unlock()

	to.clearState()
	to.Merge(from)
}

// Update adds the recorded measurement to the current data set.
func (Methods[N, Storage]) Update(state *State[N], number N) {
	state.Update(number)
}

// Merge combines two histograms that have the same buckets into a single one.
func (Methods[N, Storage]) Merge(from, to *State[N]) {
	to.Merge(from)
}

func (Methods[N, Storage]) ToAggregation(state *State[N]) aggregation.Aggregation {
	return state
}

func (Methods[N, Storage]) ToStorage(aggr aggregation.Aggregation) (*State[N], bool) {
	r, ok := aggr.(*State[N])
	return r, ok
}

func (Methods[N, Storage]) SubtractSwap(operand, argument *State[N]) {
	// This can't be called b/c histogram's are only used with synchronous instruments,
	// which start as delta temporality and thus never subtract.
	panic("impossible call")
}
