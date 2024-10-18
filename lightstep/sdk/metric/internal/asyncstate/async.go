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

package asyncstate // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/asyncstate"

import (
	"fmt"
	"sync"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/pipeline"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type (
	// State is the object used to maintain independent collection
	// state for each asynchronous meter.
	State struct {
		// pipe is the pipeline.Register number of this state.
		pipe int

		// lock protects against errant use of the instrument
		// w/ copied context after the callback returns.
		lock sync.Mutex

		// store is a map from instrument to set of values
		// observed during one collection.
		store map[*Observer]map[attribute.Set]viewstate.Accumulator
	}

	// Observer is the implementation object associated with one
	// asynchronous instrument.
	Observer struct {
		metric.Observable

		// opaque is used to ensure that callbacks are
		// registered with instruments from the same provider.
		opaque interface{}

		// compiled is the per-pipeline compiled instrument.
		compiled pipeline.Register[viewstate.Instrument]

		// descriptor describes the API-level instrument.
		//
		// Note: Not clear why we need this. It's used for a
		// range test, but shouldn't the range test be
		// performed by the aggregator?  If a View is allowed
		// to reconfigure the aggregation in ways that change
		// semantics, should the range test be based on the
		// aggregation, not the original instrument?
		descriptor sdkinstrument.Descriptor

		// performance has performance settings for the
		// instrument.
		performance sdkinstrument.Performance
	}

	implementation interface {
		get() *Observer
	}
)

func NewState(pipe int) *State {
	return &State{
		pipe:  pipe,
		store: map[*Observer]map[attribute.Set]viewstate.Accumulator{},
	}
}

// New returns a new Observer; this compiles individual
// instruments for each reader.
func New(desc sdkinstrument.Descriptor, perf sdkinstrument.Performance, opaque interface{}, compiled pipeline.Register[viewstate.Instrument]) *Observer {
	// Note: we return a non-nil instrument even when all readers
	// disabled the instrument. This ensures that certain error
	// checks still work (wrong meter, wrong callback, etc).
	//
	// Note: performance settings are not used.
	// 1. There is no fingerprinting, so IgnoreCollisions is meaningless.
	// 2. InstrumentCardinalityLimit is not enforceable, because of duplicate
	//    suppression -- better left to the aggregator.
	return &Observer{
		opaque:      opaque,
		descriptor:  desc,
		compiled:    compiled,
		performance: perf,
	}
}

// get returns this instance, used for unwrapping the instrument.
func (obs *Observer) get() *Observer {
	return obs
}

// SnapshotAndProcess calls SnapshotAndProcess() on each of the pending
// aggregations for a given reader.
func (obs *Observer) SnapshotAndProcess(state *State) {
	state.lock.Lock()
	defer state.lock.Unlock()

	for _, acc := range state.store[obs] {
		// SnapshotAndProcess is always final for asynchronous state, since
		// the map is built anew for each collection.
		acc.SnapshotAndProcess(true)
	}
}

func (obs *Observer) getOrCreate(cs *callbackState, options []metric.ObserveOption) viewstate.Accumulator {
	comp := obs.compiled[cs.state.pipe]

	if comp == nil {
		// The view disabled the instrument.
		return nil
	}

	cs.state.lock.Lock()
	defer cs.state.lock.Unlock()

	imap, has := cs.state.store[obs]

	if !has {
		imap = map[attribute.Set]viewstate.Accumulator{}
		cs.state.store[obs] = imap
	}

	ocfg := metric.NewObserveConfig(options)
	aset := ocfg.Attributes()
	aset = attribute.NewSet(obs.performance.TruncateAttributes(aset.ToSlice())...)

	se, has := imap[aset]
	if !has {
		se = comp.NewAccumulator(aset)
		imap[aset] = se
	}
	return se
}

func Observe[N number.Any, Traits number.Traits[N]](inst metric.Observable, cs *callbackState, value N, options []metric.ObserveOption) {
	cb := cs.getCallback()
	if cb == nil {
		otel.Handle(fmt.Errorf("async instrument used after callback return"))
		return
	}
	if unwrapped, ok := inst.(interface {
		Unwrap() metric.Observable
	}); ok {
		inst = unwrapped.Unwrap()
	}
	obsImpl, ok := inst.(implementation)
	if !ok {
		otel.Handle(fmt.Errorf("asynchronous instrument does not belong to this SDK: %T", inst))
		return
	}
	obs := obsImpl.get()
	if _, ok := cb.instruments[obs]; !ok {
		otel.Handle(fmt.Errorf("async instrument not declared for use in callback"))
		return
	}

	if !aggregator.RangeTest[N, Traits](value, obs.descriptor) {
		return
	}

	if acc := obs.getOrCreate(cs, options); acc != nil {
		acc.(viewstate.Updater[N]).Update(value, aggregator.ExemplarBits{})
	}
}
