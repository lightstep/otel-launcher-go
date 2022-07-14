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
		structure.Histogram[N]
	}

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

func (h Histogram[N, Traits]) Kind() aggregation.Kind {
	return aggregation.HistogramKind
}

func (h Histogram[N, Traits]) Max() number.Number {
	var traits Traits
	return traits.ToNumber(h.Histogram.Max())
}

func (h Histogram[N, Traits]) Min() number.Number {
	var traits Traits
	return traits.ToNumber(h.Histogram.Min())
}

func (h Histogram[N, Traits]) Sum() number.Number {
	var traits Traits
	return traits.ToNumber(h.Histogram.Sum())
}

func (h Histogram[N, Traits]) Negative() aggregation.Buckets {
	return h.Histogram.Negative()
}

func (h Histogram[N, Traits]) Positive() aggregation.Buckets {
	return h.Histogram.Positive()
}

func (Methods[N, Traits]) Kind() aggregation.Kind {
	return aggregation.HistogramKind
}

func (Methods[N, Traits]) Init(state *Histogram[N, Traits], cfg aggregator.Config) {
	state.Init(cfg.Histogram)
}

func (Methods[N, Traits]) HasChange(ptr *Histogram[N, Traits]) bool {
	return ptr.Count() != 0
}

func (Methods[N, Traits]) Move(from, to *Histogram[N, Traits]) {
	from.Move(&to.Histogram)
}

func (Methods[N, Traits]) Copy(from, to *Histogram[N, Traits]) {
	from.Copy(&to.Histogram)
}

func (Methods[N, Traits]) Update(state *Histogram[N, Traits], number N) {
	state.Update(number)
}

func (Methods[N, Traits]) Merge(from, to *Histogram[N, Traits]) {
	to.Merge(&from.Histogram)
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
