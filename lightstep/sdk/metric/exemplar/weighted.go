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
	"math"
	"math/rand"
	"sync"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/varopt"
)

type WeightedStorage[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]] struct {
	aggregate Storage

	lock    sync.Mutex
	samples varopt.Varopt[*aggregator.ExemplarBits]
}

type WeightedMethods[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]] struct{}

func (s *WeightedStorage[N, Storage, Methods]) Kind() aggregation.Kind {
	var am Methods
	return am.Kind()
}

func (s *WeightedStorage[N, Storage, Methods]) Unwrap() aggregation.Aggregation {
	var am Methods
	return am.ToAggregation(&s.aggregate)
}

func (m WeightedMethods[N, Storage, Methods]) Init(ptr *WeightedStorage[N, Storage, Methods], cfg aggregator.Config) {
	var am Methods
	am.Init(&ptr.aggregate, cfg)
	sz := int(cfg.Exemplar.Size)
	if sz == 0 {
		sz = aggregator.DefaultExemplarReservoirSize
	}
	ptr.samples.Init(sz, rand.New(rand.NewSource(rand.Int63())))
}

func (m WeightedMethods[N, Storage, Methods]) Update(ptr *WeightedStorage[N, Storage, Methods], value N, ex aggregator.ExemplarBits) {
	var am Methods

	if ex.Span == nil {
		// Avoid locking when the filter rejects sampling.
		am.Update(&ptr.aggregate, value, ex)
		return
	}

	// Note: The lock protects the Update() call to ensure the
	// aggregate and samples are consistent.
	ptr.lock.Lock()
	defer ptr.lock.Unlock()

	am.Update(&ptr.aggregate, value, ex)

	// am.Weight() is 1 for Histograms & (synchronous) Gauges,
	// value for (synchronous) Counters.
	//
	// The math.Abs() is a safety mechanism.  UpDownCounters would
	// otherwise cause a panic in varopt.  This is a documented
	// caveat-- for the weighted sampling logic to apply (and be
	// unbiased) for UpDownCounter measurements, we would have to
	// treat it as two monotonic instruments, one for positive and
	// one for negative measurements, and collect two samples, for
	// this process to work.  Use of absolute value can introduce
	// bias if the aim is to estimate the original data, but it
	// still yields useful exemplars.
	ptr.samples.Add(&ex, math.Abs(am.Weight(value)))
}

func (m WeightedMethods[N, Storage, Methods]) Move(input, output *WeightedStorage[N, Storage, Methods]) {
	input.lock.Lock()
	defer input.lock.Unlock()

	var am Methods
	am.Move(&input.aggregate, &output.aggregate)

	output.samples, input.samples = input.samples, output.samples
	input.samples.Reset()
}

func (m WeightedMethods[N, Storage, Methods]) Copy(input, output *WeightedStorage[N, Storage, Methods]) {
	input.lock.Lock()
	defer input.lock.Unlock()

	var am Methods
	am.Copy(&input.aggregate, &output.aggregate)

	output.samples.CopyFrom(&input.samples)
}

func (m WeightedMethods[N, Storage, Methods]) Merge(input, output *WeightedStorage[N, Storage, Methods]) {
	output.lock.Lock()
	defer output.lock.Unlock()

	var am Methods
	am.Merge(&input.aggregate, &output.aggregate)

	for i := 0; i < input.samples.Size(); i++ {
		samp, weight := input.samples.Get(i)
		output.samples.Add(samp, weight)
	}
}

func (m WeightedMethods[N, Storage, Methods]) SubtractSwap(operand, argument *WeightedStorage[N, Storage, Methods]) {
	// impossible because exemplars are for synchronous
	// instruments and subtract is only used with async
	// instruments.
	panic("impossible")
}

func (m WeightedMethods[N, Storage, Methods]) ToAggregation(ptr *WeightedStorage[N, Storage, Methods]) aggregation.Aggregation {
	return ptr
}

func (m WeightedMethods[N, Storage, Methods]) ToStorage(agg aggregation.Aggregation) (*WeightedStorage[N, Storage, Methods], bool) {
	r, ok := agg.(*WeightedStorage[N, Storage, Methods])
	return r, ok
}

func (m WeightedMethods[N, Storage, Methods]) Kind() aggregation.Kind {
	var am Methods
	return am.Kind()
}

func (m WeightedMethods[N, Storage, Methods]) HasChange(ptr *WeightedStorage[N, Storage, Methods]) bool {
	ptr.lock.Lock()
	defer ptr.lock.Unlock()

	var am Methods
	return am.HasChange(&ptr.aggregate)
}

func (m WeightedMethods[N, Storage, Methods]) Exemplars(ptr *WeightedStorage[N, Storage, Methods], in []aggregator.WeightedExemplarBits) []aggregator.WeightedExemplarBits {
	// By the time exemplars are read, the object does not require locking.
	for i := 0; i < ptr.samples.Size(); i++ {
		ex, weight := ptr.samples.Get(i)
		in = append(in, aggregator.WeightedExemplarBits{
			ExemplarBits: *ex,
			Weight:       weight,
		})
	}

	return in
}

func (m WeightedMethods[N, Storage, Methods]) Weight(n N) float64 {
	var am Methods
	return am.Weight(n)
}
