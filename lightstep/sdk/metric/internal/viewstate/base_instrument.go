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
	"fmt"
	"sync"
	"time"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/doevery"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// storageHolder is a generic struct for holding one storage and one
// auxiliary field.  Storage will be one of the aggregators.  The
// auxiliary type depends on whether synchronous or asynchronous.
//
// Auxiliary is an int64 reference count for synchronous instruments
// and notUsed for asynchronous instruments.
type storageHolder[Storage, Auxiliary any] struct {
	auxiliary Auxiliary
	storage   Storage
}

// notUsed is the Auxiliary type for asynchronous instruments.
type notUsed struct{}

// instrumentBase is the common type embedded in any of the compiled instrument views.
type instrumentBase[N number.Any, Storage, Auxiliary any, Methods aggregator.Methods[N, Storage]] struct {
	instLock sync.Mutex
	fromName string
	desc     sdkinstrument.Descriptor
	acfg     aggregator.Config
	data     map[attribute.Set]*storageHolder[Storage, Auxiliary]

	keysSet    *attribute.Set
	keysFilter *attribute.Filter
}

// Size reports the size of the data map.
func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) Size() int {
	metric.instLock.Lock()
	defer metric.instLock.Unlock()
	return len(metric.data)
}

func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) Aggregation() aggregation.Kind {
	var methods Methods
	return methods.Kind()
}

func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) OriginalName() string {
	return metric.fromName
}

func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) Descriptor() sdkinstrument.Descriptor {
	return metric.desc
}

func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) Keys() *attribute.Set {
	return metric.keysSet
}

func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) Config() aggregator.Config {
	return metric.acfg
}

func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) initStorage(s *Storage) {
	var methods Methods
	methods.Init(s, metric.acfg)
}

func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) mergeDescription(d string) {
	metric.instLock.Lock()
	defer metric.instLock.Unlock()
	if len(d) > len(metric.desc.Description) {
		metric.desc.Description = d
	}
}

// isValidAttribute supports filtering invalid attributes.  Note, this
// should be fast, trye not to allocate!  Note: the specification is
// somewhat ambiguous about empty strings, see
// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/common/attribute-naming.md
// which just says "valid Unicode sequence".
func isValidAttribute(kv attribute.KeyValue) bool {
	return kv.Key != ""
}

func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) invalidAttributeFilter(kv attribute.KeyValue) bool {
	return isValidAttribute(kv) && (metric.keysFilter == nil || (*metric.keysFilter)(kv))
}

func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) applyKeysFilter(kvs attribute.Set) attribute.Set {
	invalidFilter := false
	for iter := kvs.Iter(); iter.Next(); {
		kv := iter.Attribute()
		if !isValidAttribute(kv) {
			doevery.TimePeriod(time.Minute, func() {
				otel.Handle(fmt.Errorf("use of empty attribute key, e.g., with value %q", kv.Value.Emit()))
			})
			invalidFilter = true
			break
		}
	}

	if !invalidFilter && metric.keysFilter == nil {
		return kvs
	}
	var res attribute.Set
	if !invalidFilter {
		res, _ = kvs.Filter(*metric.keysFilter)
	} else {
		res, _ = kvs.Filter(metric.invalidAttributeFilter)
	}
	return res
}

func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) getOrCreateEntry(kvs attribute.Set) *storageHolder[Storage, Auxiliary] {
	entry, has := metric.data[kvs]
	if has {
		return entry
	}

	var methods Methods
	entry = &storageHolder[Storage, Auxiliary]{}
	methods.Init(&entry.storage, metric.acfg)
	metric.data[kvs] = entry
	return entry
}

// newStorage allocates and initializes a new Storage.
func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) newStorage() *Storage {
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
func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) appendInstrument(output *[]data.Instrument) *data.Instrument {
	inst := data.ReallocateFrom(output)
	inst.Descriptor = metric.desc
	return inst
}

// appendPoint adds a new point to the output.  Note that the existing
// slice will be extended, if possible, and the existing Aggregation
// is potentially re-used.  The variable `reset` determines whether
// Move() or Copy() is used.  Note that both Move and Copy are
// synchronized with respect to Update() and Merge(), necessary for the
// synchronous code path which may see concurrent collection.
func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) appendPoint(inst *data.Instrument, set attribute.Set, storage *Storage, tempo aggregation.Temporality, start, end time.Time, reset bool) {
	var methods Methods

	// Possibly re-use the underlying storage.
	point, out := metric.appendOrReusePoint(inst)
	if out == nil {
		out = metric.newStorage()
	}

	if reset {
		// Note: synchronized move uses swap for expensive
		// copies, like histogram.
		methods.Move(storage, out)
	} else {
		methods.Copy(storage, out)
	}

	point.Attributes = set
	point.Aggregation = methods.ToAggregation(out)
	point.Temporality = tempo
	point.Start = start
	point.End = end
}

// appendOrReusePoint is an alternate to appendPoint; this form is used when
// the storage will be reset on collection.
func (metric *instrumentBase[N, Storage, Auxiliary, Methods]) appendOrReusePoint(inst *data.Instrument) (*data.Point, *Storage) {
	point := data.ReallocateFrom(&inst.Points)

	var methods Methods
	if s, ok := methods.ToStorage(point.Aggregation); ok {
		return point, s
	}
	return point, nil
}
