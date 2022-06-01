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

package viewstate // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"

import (
	"sync"
	"time"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel/attribute"
)

// instrumentBase is the common type embedded in any of the compiled instrument views.
type instrumentBase[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]] struct {
	lock     sync.Mutex
	fromName string
	desc     sdkinstrument.Descriptor
	acfg     aggregator.Config
	data     map[attribute.Set]*Storage

	keysSet    *attribute.Set
	keysFilter *attribute.Filter
}

func (metric *instrumentBase[N, Storage, Methods]) Aggregation() aggregation.Kind {
	var methods Methods
	return methods.Kind()
}

func (metric *instrumentBase[N, Storage, Methods]) OriginalName() string {
	return metric.fromName
}

func (metric *instrumentBase[N, Storage, Methods]) Descriptor() sdkinstrument.Descriptor {
	return metric.desc
}

func (metric *instrumentBase[N, Storage, Methods]) Keys() *attribute.Set {
	return metric.keysSet
}

func (metric *instrumentBase[N, Storage, Methods]) Config() aggregator.Config {
	return metric.acfg
}

func (metric *instrumentBase[N, Storage, Methods]) initStorage(s *Storage) {
	var methods Methods
	methods.Init(s, metric.acfg)
}

func (metric *instrumentBase[N, Storage, Methods]) mergeDescription(d string) {
	metric.lock.Lock()
	defer metric.lock.Unlock()
	if len(d) > len(metric.desc.Description) {
		metric.desc.Description = d
	}
}

// storageFinder searches for and possibly allocates an output Storage
// for this metric.  Filtered keys, if a filter is provided, will be
// computed once.
func (metric *instrumentBase[N, Storage, Methods]) storageFinder(
	kvs attribute.Set,
) func() *Storage {
	if metric.keysFilter != nil {
		kvs, _ = attribute.NewSetWithFiltered(kvs.ToSlice(), *metric.keysFilter)
	}

	return func() *Storage {
		metric.lock.Lock()
		defer metric.lock.Unlock()

		storage, has := metric.data[kvs]
		if has {
			return storage
		}

		ns := metric.newStorage()
		metric.data[kvs] = ns
		return ns
	}
}

// newStorage allocates and initializes a new Storage.
func (metric *instrumentBase[N, Storage, Methods]) newStorage() *Storage {
	ns := new(Storage)
	metric.initStorage(ns)
	return ns
}

// appendInstrument adds a new instrument to the output.  Note that
// this is expected to be called unconditionally (even when there are
// no points); it means that the same list of instruments will always
// be produced (in the same order); consumers of delta temporality
// should expect to see empty instruments in the output for metric
// data that is unchanged.
func (metric *instrumentBase[N, Storage, Methods]) appendInstrument(output *[]data.Instrument) *data.Instrument {
	inst := data.ReallocateFrom(output)
	inst.Descriptor = metric.desc
	return inst
}

// appendPoint is used in cases where the output Aggregation is the
// stored object; use appendOrReusePoint in the case where the output
// Aggregation is a copy of the stored object (in case the stored
// object will be reset on collection, as opposed to a second pass to
// reset delta temporality outputs before the next accumulation.
func (metric *instrumentBase[N, Storage, Methods]) appendPoint(inst *data.Instrument, set attribute.Set, agg aggregation.Aggregation, tempo aggregation.Temporality, start, end time.Time) {
	point := data.ReallocateFrom(&inst.Points)

	point.Attributes = set
	point.Aggregation = agg
	point.Temporality = tempo
	point.Start = start
	point.End = end
}

// appendOrReusePoint is an alternate to appendPoint; this form is used when
// the storage will be reset on collection.
func (metric *instrumentBase[N, Storage, Methods]) appendOrReusePoint(inst *data.Instrument) (*data.Point, *Storage) {
	point := data.ReallocateFrom(&inst.Points)

	var methods Methods
	if s, ok := methods.ToStorage(point.Aggregation); ok {
		return point, s
	}
	return point, nil
}
