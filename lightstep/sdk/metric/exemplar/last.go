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
	"sync"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
)

type LastStorage[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]] struct {
	aggregate Storage

	lock     sync.Mutex
	exemplar aggregator.ExemplarBits
}

type LastMethods[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]] struct{}

func (s *LastStorage[N, Storage, Methods]) Kind() aggregation.Kind {
	var am Methods
	return am.Kind()
}

func (s *LastStorage[N, Storage, Methods]) Unwrap() aggregation.Aggregation {
	var am Methods
	return am.ToAggregation(&s.aggregate)
}

func (m LastMethods[N, Storage, Methods]) Init(ptr *LastStorage[N, Storage, Methods], cfg aggregator.Config) {
	var am Methods
	am.Init(&ptr.aggregate, cfg)
}

func (m LastMethods[N, Storage, Methods]) Update(ptr *LastStorage[N, Storage, Methods], number N, ex aggregator.ExemplarBits) {
	var am Methods
	if ex.Attributes == nil {
		am.Update(&ptr.aggregate, number, ex)
		return
	}

	ptr.lock.Lock()
	defer ptr.lock.Unlock()
	ptr.exemplar = ex
	am.Update(&ptr.aggregate, number, ex)
}

func (m LastMethods[N, Storage, Methods]) Move(input, output *LastStorage[N, Storage, Methods]) {
	var am Methods
	input.lock.Lock()
	defer input.lock.Unlock()
	output.exemplar, input.exemplar = input.exemplar, output.exemplar
	am.Move(&input.aggregate, &output.aggregate)
}

func (m LastMethods[N, Storage, Methods]) Copy(input, output *LastStorage[N, Storage, Methods]) {
	var am Methods
	input.lock.Lock()
	defer input.lock.Unlock()
	output.exemplar = input.exemplar
	am.Copy(&input.aggregate, &output.aggregate)
}

func (m LastMethods[N, Storage, Methods]) Merge(input, output *LastStorage[N, Storage, Methods]) {
	var am Methods
	output.lock.Lock()
	defer output.lock.Unlock()
	if input.exemplar.Attributes != nil {
		output.exemplar = input.exemplar
	}
	am.Merge(&input.aggregate, &output.aggregate)
}

func (m LastMethods[N, Storage, Methods]) SubtractSwap(operand, argument *LastStorage[N, Storage, Methods]) {
	panic("impossible use")
}

func (m LastMethods[N, Storage, Methods]) ToAggregation(ptr *LastStorage[N, Storage, Methods]) aggregation.Aggregation {
	return ptr
}

func (m LastMethods[N, Storage, Methods]) ToStorage(agg aggregation.Aggregation) (*LastStorage[N, Storage, Methods], bool) {
	r, ok := agg.(*LastStorage[N, Storage, Methods])
	return r, ok
}

func (m LastMethods[N, Storage, Methods]) Kind() aggregation.Kind {
	var am Methods
	return am.Kind()
}

func (m LastMethods[N, Storage, Methods]) HasChange(ptr *LastStorage[N, Storage, Methods]) bool {
	var am Methods
	return am.HasChange(&ptr.aggregate)
}

func (m LastMethods[N, Storage, Methods]) Exemplars(ptr *LastStorage[N, Storage, Methods], in []aggregator.WeightedExemplarBits) []aggregator.WeightedExemplarBits {
	// By the time exemplars are read, the object does not require locking.
	return append(in, aggregator.WeightedExemplarBits{
		ExemplarBits: ptr.exemplar,
	})
}

func (m LastMethods[N, Storage, Methods]) Weight(n N) float64 {
	var am Methods
	return am.Weight(n)
}
