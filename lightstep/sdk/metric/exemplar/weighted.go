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
	"github.com/lightstep/varopt"
)

type WeightedStorage[N number.Any, Traits number.Traits[N], Storage any, Methods aggregator.Methods[N, Storage]] struct {
	aggregate Storage

	lock    sync.Mutex
	samples varopt.Varopt
}

type WeightedMethods[N number.Any, Traits number.Traits[N], Storage any, Methods aggregator.Methods[N, Storage]] struct{}

func (s WeightedStorage[N, Traits, Storage, Methods]) Kind() aggregation.Kind {
	var am Methods
	return am.Kind()
}

func (m WeightedMethods[N, Traits, Storage, Methods]) Init(ptr *WeightedStorage[N, Traits, Storage, Methods], cfg aggregator.Config) {
	var am Methods
	am.Init(&ptr.aggregate, cfg)
	ptr.samples.Init(int(cfg.Exemplar.Size), cfg.Exemplar.Rnd)
}

func (m WeightedMethods[N, Traits, Storage, Methods]) Update(ptr *WeightedStorage[N, Traits, Storage, Methods], value N) {
	// Note: should the lock protect the Update() call as well to
	// ensure the aggregate and samples are consistent?
	ptr.lock.Lock()
	defer ptr.lock.Unlock()

	var am Methods
	var tr Traits
	am.Update(&ptr.aggregate, value)

	var samp varopt.Sample // @@@ update library to be generic.
	ptr.samples.Add(samp, number.ToFloat64(tr.ToNumber(value)))
}

func (m WeightedMethods[N, Traits, Storage, Methods]) Move(input, output *WeightedStorage[N, Traits, Storage, Methods]) {
	// @@@ see histogram, output lock is correct?
	output.lock.Lock()
	defer output.lock.Unlock()

	var am Methods
	am.Move(&input.aggregate, &output.aggregate)

	output.samples, input.samples = input.samples, output.samples
	input.samples.Reset()
}

func (m WeightedMethods[N, Traits, Storage, Methods]) Merge(input, output *WeightedStorage[N, Traits, Storage, Methods]) {
	output.lock.Lock()
	defer output.lock.Unlock()

	var am Methods
	am.Merge(&input.aggregate, &output.aggregate)

	for input.Size() {
		output.Add(input.Value())
	}
}

func (m WeightedMethods[N, Traits, Storage, Methods]) Copy(input, output *WeightedStorage[N, Traits, Storage, Methods]) {
	// @@@ see histogram, output lock is correct?
	output.lock.Lock()
	defer output.lock.Unlock()

	var am Methods
	am.Copy(&input.aggregate, &output.aggregate)

	for {
		// @@@ Copy each item?
	}
}

func (m WeightedMethods[N, Traits, Storage, Methods]) SubtractSwap(operand, argument *WeightedStorage[N, Traits, Storage, Methods]) {
	// impossible because exemplars are for synchronous
	// instruments and subtract is only used with async
	// instruments.
	panic("impossible")
}

func (m WeightedMethods[N, Traits, Storage, Methods]) ToAggregation(ptr *WeightedStorage[N, Traits, Storage, Methods]) aggregation.Aggregation {
	return ptr
}

func (m WeightedMethods[N, Traits, Storage, Methods]) ToStorage(agg aggregation.Aggregation) (*WeightedStorage[N, Traits, Storage, Methods], bool) {
	r, ok := agg.(*WeightedStorage[N, Traits, Storage, Methods])
	return r, ok
}

func (m WeightedMethods[N, Traits, Storage, Methods]) Kind() aggregation.Kind {
	var am Methods
	return am.Kind()
}

func (m WeightedMethods[N, Traits, Storage, Methods]) HasChange(ptr *WeightedStorage[N, Traits, Storage, Methods]) bool {
	ptr.lock.Lock()
	defer ptr.lock.Unlock()

	var am Methods
	return am.HasChange(&ptr.aggregate)
}
