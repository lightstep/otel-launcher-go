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

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram/structure"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
)

var ErrNoSubtract = fmt.Errorf("histogram subtract not implemented")

type (
	Methods[N number.Any, Storage structure.State[N]] struct{}

	Int64Methods   = Methods[int64, structure.Int64]
	Float64Methods = Methods[float64, structure.Float64]
)

var (
	_ aggregator.Methods[int64, structure.Int64]     = Int64Methods{}
	_ aggregator.Methods[float64, structure.Float64] = Float64Methods{}

	_ aggregation.Histogram = asLightstepHistogram[int64, number.Int64Traits](&structure.Int64{})
	_ aggregation.Histogram = asLightstepHistogram[float64, number.Float64Traits](&structure.Float64{})
)

func asLightstepHistogram[N structure.ValueType, Traits number.Traits[N]](h *structure.State[N]) aggregation.Histogram {
	// @@@
	return nil
}

func (Methods[N, Storage]) Kind() aggregation.Kind {
	return aggregation.HistogramKind
}

func (Methods[N, Storage]) Init(state *structure.State[N], cfg aggregator.Config) {
	state.Init(cfg.Histogram)
}

func (Methods[N, Storage]) HasChange(ptr *structure.State[N]) bool {
	return ptr.Count() != 0
}

func (Methods[N, Storage]) Move(from, to *structure.State[N]) {
	from.Move(to)
}

func (Methods[N, Storage]) Copy(from, to *structure.State[N]) {
	from.Copy(to)
}

// Update adds the recorded measurement to the current data set.
func (Methods[N, Storage]) Update(state *structure.State[N], number N) {
	state.Update(number)
}

// Merge combines two histograms that have the same buckets into a single one.
func (Methods[N, Storage]) Merge(from, to *structure.State[N]) {
	to.Merge(from)
}

func (Methods[N, Storage]) ToAggregation(state *structure.State[N]) aggregation.Aggregation {
	return state
}

func (Methods[N, Storage]) ToStorage(aggr aggregation.Aggregation) (*structure.State[N], bool) {
	r, ok := aggr.(*structure.State[N])
	return r, ok
}

func (Methods[N, Storage]) SubtractSwap(operand, argument *structure.State[N]) {
	// This can't be called b/c histogram's are only used with synchronous instruments,
	// which start as delta temporality and thus never subtract.
	panic("impossible call")
}
