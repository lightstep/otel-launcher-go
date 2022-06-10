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
	State[N number.Any, Traits number.Traits[N]] struct {
		lock    sync.Mutex
		maxSize int32
		data[N, Traits]
	}

	data[N number.Any, Traits number.Traits[N]] struct {
		// sum is the sum of all Updates reflected in the
		// aggregator.  It has the same type number as the
		// corresponding sdkinstrument.Descriptor.
		sum N
		// count is incremented by 1 per Update.
		count uint64
		// zeroCount is incremented by 1 when the measured
		// value is exactly 0.
		zeroCount uint64
		// positive holds the positive values
		positive buckets
		// negative holds the negative values in these buckets
		// by their absolute value.
		negative buckets
		// mapping corresponds to the current scale, is shared
		// by both positive and negative ranges.
		mapping mapping.Mapping
	}

	Methods[N number.Any, Traits number.Traits[N], Storage State[N, Traits]] struct{}

	Int64   = State[int64, number.Int64Traits]
	Float64 = State[float64, number.Float64Traits]

	Int64Methods   = Methods[int64, number.Int64Traits, Int64]
	Float64Methods = Methods[float64, number.Float64Traits, Float64]
)

// DefaultMaxSize is the default number of buckets.
//
// 256 is a good choice
// 320 is a historical choice
//
// The OpenHistogram representation of the Prometheus default explicit
// histogram boundaries (spanning 0.005 to 10) yields 320 base-10
// 90-per-decade log-linear buckets.   NrSketch used 320.
//
// OTel settled on 160, which yields a maximum relative error of less
// than 5% for data with contrast 10^5 (e.g., latencies in the range
// 1ms to 100s).  See the derivation here: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/metrics/sdk.md#exponential-histogram-aggregation
const DefaultMaxSize int32 = 160

// MinSize is the smallest reasonable configuration, which is small
// enough to contain the entire normal floating point range at
// MinScale.
const MinSize = 2

var (
	_ aggregator.Methods[int64, Int64]     = Int64Methods{}
	_ aggregator.Methods[float64, Float64] = Float64Methods{}

	_ aggregation.Histogram = &Int64{}
	_ aggregation.Histogram = &Float64{}
)

func NewFloat64(cfg aggregator.HistogramConfig, values ...float64) *Float64 {
	return newHist[float64, number.Float64Traits](cfg, values...)
}

func NewInt64(cfg aggregator.HistogramConfig, values ...int64) *Int64 {
	return newHist[int64, number.Int64Traits](cfg, values...)
}

type Option func(aggregator.HistogramConfig) aggregator.HistogramConfig

func WithMaxSize(maxSize int32) Option {
	return func(cfg aggregator.HistogramConfig) aggregator.HistogramConfig {
		cfg.MaxSize = maxSize
		return cfg
	}
}

func NewConfig(opts ...Option) aggregator.HistogramConfig {
	var cfg aggregator.HistogramConfig
	for _, opt := range opts {
		cfg = opt(cfg)
	}
	return cfg
}

func newHist[N number.Any, Traits number.Traits[N]](cfg aggregator.HistogramConfig, values ...N) *State[N, Traits] {
	var methods Methods[N, Traits, State[N, Traits]]

	state := &State[N, Traits]{}

	methods.Init(state, aggregator.Config{
		Histogram: cfg,
	})

	for _, val := range values {
		methods.Update(state, val)
	}
	return state
}

// Sum implements aggregation.Histogram.
func (h *State[N, Traits]) Sum() number.Number {
	var t Traits
	return t.ToNumber(h.sum)
}

// Count implements aggregation.Histogram.
func (h *State[N, Traits]) Count() uint64 {
	return h.count
}

// Scale implements aggregation.Histogram.
func (s *State[N, Traits]) Scale() int32 {
	if s.data.count == s.data.zeroCount {
		// all zeros! scale doesn't matter, use zero.
		return 0
	}
	return s.data.mapping.Scale()
}

// ZeroCount implements aggregation.Histogram.
func (s *State[N, Traits]) ZeroCount() uint64 {
	return s.data.zeroCount
}

// Positive implements aggregation.Histogram.
func (s *State[N, Traits]) Positive() aggregation.Buckets {
	return &s.data.positive
}

// Negative implements aggregation.Histogram.
func (s *State[N, Traits]) Negative() aggregation.Buckets {
	return &s.data.negative
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
func (s *State[N, Traits]) clearState() {
	s.positive.clearState()
	s.negative.clearState()
	s.sum = 0
	s.count = 0
	s.zeroCount = 0
	s.mapping, _ = newMapping(logarithm.MaxScale)
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

func (h *State[N, Traits]) Kind() aggregation.Kind {
	return aggregation.HistogramKind
}

func (Methods[N, Traits, Storage]) Kind() aggregation.Kind {
	return aggregation.HistogramKind
}

func (Methods[N, Traits, Storage]) Init(state *State[N, Traits], cfg aggregator.Config) {
	state.maxSize = cfg.Histogram.MaxSize

	if state.maxSize == 0 {
		state.maxSize = DefaultMaxSize
	}
	if state.maxSize < MinSize {
		state.maxSize = MinSize
	}

	mapping, _ := newMapping(logarithm.MaxScale)
	state.mapping = mapping
}

func (Methods[N, Traits, Storage]) HasChange(ptr *State[N, Traits]) bool {
	return ptr.count != 0
}

func (Methods[N, Traits, Storage]) Move(from, to *State[N, Traits]) {
	from.Move(to)
}

func (Methods[N, Traits, Storage]) Copy(from, to *State[N, Traits]) {
	from.lock.Lock()
	defer from.lock.Unlock()

	to.clearState()
	to.Merge(from)
}

// Update adds the recorded measurement to the current data set.
func (Methods[N, Traits, Storage]) Update(state *State[N, Traits], number N) {
	state.Update(number)
}

// Merge combines two histograms that have the same buckets into a single one.
func (Methods[N, Traits, Storage]) Merge(from, to *State[N, Traits]) {
	to.Merge(from)
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
