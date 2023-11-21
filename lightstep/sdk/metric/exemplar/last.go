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

package exemplar

import (
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
)

type LastStorage[N number.Any, Traits number.Traits[N], Storage any, Methods aggregator.Methods[N, Storage]] struct {
	Aggregate Storage
}

type LastMethods[N number.Any, Traits number.Traits[N], Storage any, Methods aggregator.Methods[N, Storage]] struct{}

func (s LastStorage[N, Traits, Storage, Methods]) Kind() aggregation.Kind {
	var am Methods
	return am.Kind()
}

func (m LastMethods[N, Traits, Storage, Methods]) Init(ptr *LastStorage[N, Traits, Storage, Methods], cfg aggregator.Config) {
	var am Methods
	am.Init(&ptr.Aggregate, cfg)
}

func (m LastMethods[N, Traits, Storage, Methods]) Update(ptr *LastStorage[N, Traits, Storage, Methods], number N, ex aggregator.ExemplarBits) {
	var am Methods
	// TODO: ex
	am.Update(&ptr.Aggregate, number, ex)
}

func (m LastMethods[N, Traits, Storage, Methods]) Move(input, output *LastStorage[N, Traits, Storage, Methods]) {
	var am Methods
	am.Move(&input.Aggregate, &output.Aggregate)
}

func (m LastMethods[N, Traits, Storage, Methods]) Merge(input, output *LastStorage[N, Traits, Storage, Methods]) {
	var am Methods
	am.Merge(&input.Aggregate, &output.Aggregate)
}

func (m LastMethods[N, Traits, Storage, Methods]) Copy(input, output *LastStorage[N, Traits, Storage, Methods]) {
	var am Methods
	am.Copy(&input.Aggregate, &output.Aggregate)
}

func (m LastMethods[N, Traits, Storage, Methods]) SubtractSwap(operand, argument *LastStorage[N, Traits, Storage, Methods]) {
	panic("impossible use")
}

func (m LastMethods[N, Traits, Storage, Methods]) ToAggregation(ptr *LastStorage[N, Traits, Storage, Methods]) aggregation.Aggregation {
	return ptr
}

func (m LastMethods[N, Traits, Storage, Methods]) ToStorage(agg aggregation.Aggregation) (*LastStorage[N, Traits, Storage, Methods], bool) {
	r, ok := agg.(*LastStorage[N, Traits, Storage, Methods])
	return r, ok
}

func (m LastMethods[N, Traits, Storage, Methods]) Kind() aggregation.Kind {
	var am Methods
	return am.Kind()
}

func (m LastMethods[N, Traits, Storage, Methods]) HasChange(ptr *LastStorage[N, Traits, Storage, Methods]) bool {
	var am Methods
	return am.HasChange(&ptr.Aggregate)
}
