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
	"context"
	"fmt"
	"sync"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/pipeline"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
		store map[*Instrument]map[attribute.Set]viewstate.Accumulator
	}

	// Instrument is the implementation object associated with one
	// asynchronous instrument.
	Instrument struct {
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
	}

	// contextKey is used with context.WithValue() to lookup
	// per-reader state from within an executing callback
	// function.
	contextKey struct{}
)

func NewState(pipe int) *State {
	return &State{
		pipe:  pipe,
		store: map[*Instrument]map[attribute.Set]viewstate.Accumulator{},
	}
}

// NewInstrument returns a new Instrument; this compiles individual
// instruments for each reader.
func NewInstrument(desc sdkinstrument.Descriptor, opaque interface{}, compiled pipeline.Register[viewstate.Instrument]) *Instrument {
	// Note: we return a non-nil instrument even when all readers
	// disabled the instrument. This ensures that certain error
	// checks still work (wrong meter, wrong callback, etc).
	return &Instrument{
		opaque:     opaque,
		descriptor: desc,
		compiled:   compiled,
	}
}

// SnapshotAndProcess calls SnapshotAndProcess() on each of the pending
// aggregations for a given reader.
func (inst *Instrument) SnapshotAndProcess(state *State) {
	state.lock.Lock()
	defer state.lock.Unlock()

	for _, acc := range state.store[inst] {
		// SnapshotAndProcess is always final for asynchronous state, since
		// the map is built anew for each collection.
		acc.SnapshotAndProcess(true)
	}
}

func (inst *Instrument) getOrCreate(cs *callbackState, attrs []attribute.KeyValue) viewstate.Accumulator {
	comp := inst.compiled[cs.state.pipe]

	if comp == nil {
		// The view disabled the instrument.
		return nil
	}

	cs.state.lock.Lock()
	defer cs.state.lock.Unlock()

	imap, has := cs.state.store[inst]

	if !has {
		imap = map[attribute.Set]viewstate.Accumulator{}
		cs.state.store[inst] = imap
	}

	aset := attribute.NewSet(attrs...)
	se, has := imap[aset]
	if !has {
		se = comp.NewAccumulator(aset)
		imap[aset] = se
	}
	return se
}

func capture[N number.Any, Traits number.Traits[N]](ctx context.Context, inst *Instrument, value N, attrs []attribute.KeyValue) {
	lookup := ctx.Value(contextKey{})
	if lookup == nil {
		otel.Handle(fmt.Errorf("async instrument used outside of callback"))
		return
	}

	cs := lookup.(*callbackState)
	cb := cs.getCallback()
	if cb == nil {
		otel.Handle(fmt.Errorf("async instrument used after callback return"))
		return
	}
	if _, ok := cb.instruments[inst]; !ok {
		otel.Handle(fmt.Errorf("async instrument not declared for use in callback"))
		return
	}

	if !aggregator.RangeTest[N, Traits](value, inst.descriptor.Kind) {
		return
	}

	if acc := inst.getOrCreate(cs, attrs); acc != nil {
		acc.(viewstate.Updater[N]).Update(value)
	}
}
