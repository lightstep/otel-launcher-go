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
	"sync"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram/structure"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
)

// The methods in this file adapt the basic structure in ./structure
// to the Methods interface pattern used in this SDK.

type (
	Methods[N number.Any, Traits number.Traits[N]] struct{}

	Histogram[N number.Any, Traits number.Traits[N]] struct {
		lock      sync.Mutex
		Histogram structure.Histogram[N]
	}

	Config = structure.Config
	Option = structure.Option

	Int64Methods   = Methods[int64, number.Int64Traits]
	Float64Methods = Methods[float64, number.Float64Traits]

	Int64   = Histogram[int64, number.Int64Traits]
	Float64 = Histogram[float64, number.Float64Traits]
)

var (
	_ aggregator.Methods[int64, Int64]     = Int64Methods{}
	_ aggregator.Methods[float64, Float64] = Float64Methods{}

	_ aggregation.Histogram = &Histogram[int64, number.Int64Traits]{}
	_ aggregation.Histogram = &Histogram[float64, number.Float64Traits]{}
)

const (
	MinSize        = structure.MinSize
	DefaultMaxSize = structure.DefaultMaxSize
	MaximumMaxSize = structure.MaximumMaxSize
)

func NewFloat64(cfg Config, fs ...float64) *Float64 {
	return &Float64{
		Histogram: *structure.NewFloat64(cfg, fs...),
	}
}

func NewInt64(cfg Config, is ...int64) *Int64 {
	return &Int64{
		Histogram: *structure.NewInt64(cfg, is...),
	}
}

func NewConfig(opts ...Option) Config {
	return structure.NewConfig(opts...)
}

func WithMaxSize(sz int32) Option {
	return structure.WithMaxSize(sz)
}

func (h *Histogram[N, Traits]) Kind() aggregation.Kind {
	return aggregation.HistogramKind
}

func (h *Histogram[N, Traits]) Max() number.Number {
	var traits Traits
	return traits.ToNumber(h.Histogram.Max())
}

func (h *Histogram[N, Traits]) Min() number.Number {
	var traits Traits
	return traits.ToNumber(h.Histogram.Min())
}

func (h *Histogram[N, Traits]) Sum() number.Number {
	var traits Traits
	return traits.ToNumber(h.Histogram.Sum())
}

func (h *Histogram[N, Traits]) Count() uint64 {
	return h.Histogram.Count()
}

func (h *Histogram[N, Traits]) ZeroCount() uint64 {
	return h.Histogram.ZeroCount()
}

func (h *Histogram[N, Traits]) Negative() aggregation.Buckets {
	return h.Histogram.Negative()
}

func (h *Histogram[N, Traits]) Positive() aggregation.Buckets {
	return h.Histogram.Positive()
}

func (h *Histogram[N, Traits]) Scale() int32 {
	return h.Histogram.Scale()
}

func (Methods[N, Traits]) Kind() aggregation.Kind {
	return aggregation.HistogramKind
}

func (Methods[N, Traits]) Init(agg *Histogram[N, Traits], cfg aggregator.Config) {
	agg.Histogram.Init(cfg.Histogram)
}

func (Methods[N, Traits]) HasChange(ptr *Histogram[N, Traits]) bool {
	return ptr.Count() != 0
}

func (Methods[N, Traits]) Update(agg *Histogram[N, Traits], number N) {
	agg.lock.Lock()
	defer agg.lock.Unlock()
	agg.Histogram.Update(number)
}

func (Methods[N, Traits]) Move(from, to *Histogram[N, Traits]) {
	to.Histogram.Clear()

	from.lock.Lock()
	defer from.lock.Unlock()
	from.Histogram.Swap(&to.Histogram)
}

func (Methods[N, Traits]) Copy(from, to *Histogram[N, Traits]) {
	from.lock.Lock()
	defer from.lock.Unlock()
	from.Histogram.CopyInto(&to.Histogram)
}

func (Methods[N, Traits]) Merge(from, to *Histogram[N, Traits]) {
	to.lock.Lock()
	defer to.lock.Unlock()
	to.Histogram.MergeFrom(&from.Histogram)
}

func (Methods[N, Traits]) ToAggregation(histo *Histogram[N, Traits]) aggregation.Aggregation {
	return histo
}

func (Methods[N, Traits]) ToStorage(aggr aggregation.Aggregation) (*Histogram[N, Traits], bool) {
	r, ok := aggr.(*Histogram[N, Traits])
	return r, ok
}

func (Methods[N, Traits]) SubtractSwap(operand, argument *Histogram[N, Traits]) {
	// This can't be called b/c histogram's are only used with synchronous instruments,
	// which start as delta temporality and thus never subtract.
	panic("impossible call")
}
