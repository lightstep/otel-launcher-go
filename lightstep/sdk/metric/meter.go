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

package metric // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"

import (
	"fmt"
	"sync"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/asyncstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/pipeline"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/syncstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type (
	// meter handles the creation and coordination of all metric instruments. A
	// meter represents a single instrumentation scope; all metric telemetry
	// produced by an instrumentation scope will use metric instruments from a
	// single meter.
	meter struct {
		library   instrumentation.Library
		provider  *MeterProvider
		compilers pipeline.Register[*viewstate.Compiler]

		lock       sync.Mutex
		byDesc     map[sdkinstrument.Descriptor]interface{}
		syncInsts  []*syncstate.Observer
		asyncInsts []*asyncstate.Observer
		callbacks  []*asyncstate.Callback
	}

	// metricRegistration implements metric.Registration
	metricRegistration struct {
		lock     sync.Mutex
		meter    *meter
		callback *asyncstate.Callback
	}

	// instConfig generalizes the float/int-specific
	// instrument.Config interfaces.
	instConfig interface {
		Description() string
		Unit() unit.Unit
	}
)

// Compile-time check meter implements metric.Meter.
var _ metric.Meter = (*meter)(nil)

var ErrAlreadyUnregistered = fmt.Errorf("callback was already unregistered")

// RegisterCallback registers the function f to be called when any of the
// insts Collect method is called.
func (m *meter) RegisterCallback(function metric.Callback, instruments ...instrument.Asynchronous) (metric.Registration, error) {
	cb, err := asyncstate.NewCallback(instruments, m, function)

	var reg metric.Registration
	if err == nil {
		m.lock.Lock()
		defer m.lock.Unlock()
		m.callbacks = append(m.callbacks, cb)
		reg = &metricRegistration{
			meter:    m,
			callback: cb,
		}
	}
	return reg, err
}

// Unregister prevents the callback from being invoked in the future.
func (mr *metricRegistration) Unregister() error {
	mr.lock.Lock()
	defer mr.lock.Unlock()

	if mr.callback == nil {
		return ErrAlreadyUnregistered
	}

	mycb := mr.callback
	cbs := mr.meter.callbacks

	for i := 0; i < len(cbs); i++ {
		if mycb == cbs[i] {
			cbs[i] = cbs[len(cbs)-1]
			mr.meter.callbacks = cbs[:len(cbs)-1]
			break
		}
	}

	mr.callback = nil

	return nil
}

// instrumentConstructor refers to either the syncstate or asyncstate
// NewInstrument method.  Although both receive an opaque interface{}
// to distinguish providers, only the asyncstate package needs to know
// this information.  The unused parameter is passed to the syncstate
// package for the generalization used here to work.
type instrumentConstructor[T any] func(
	instrument sdkinstrument.Descriptor,
	opaque interface{},
	compiled pipeline.Register[viewstate.Instrument],
) *T

// configureInstrument applies the instrument configuration, checks
// for an existing definition for the same descriptor, and compiles
// and constructs the instrument if necessary.
func configureInstrument[T any](
	m *meter,
	name string,
	cfg instConfig,
	nk number.Kind,
	ik sdkinstrument.Kind,
	listPtr *[]*T,
	ctor instrumentConstructor[T],
) (*T, error) {
	// Compute the instrument descriptor
	desc := sdkinstrument.NewDescriptor(name, ik, nk, cfg.Description(), cfg.Unit())

	m.lock.Lock()
	defer m.lock.Unlock()

	// Lookup a pre-existing instrument by descriptor.
	if lookup, has := m.byDesc[desc]; has {
		// Recompute conflicts since they may have changed.
		var conflicts viewstate.ViewConflictsBuilder

		for _, compiler := range m.compilers {
			_, err := compiler.Compile(desc)
			conflicts.Combine(err)
		}

		return lookup.(*T), conflicts.AsError()
	}

	// Compile the instrument for each pipeline. the first time.
	var conflicts viewstate.ViewConflictsBuilder
	compiled := pipeline.NewRegister[viewstate.Instrument](len(m.compilers))

	for pipe, compiler := range m.compilers {
		comp, err := compiler.Compile(desc)
		compiled[pipe] = comp
		conflicts.Combine(err)
	}

	// Build the new instrument, cache it, append to the list.
	inst := ctor(desc, m, compiled)
	err := conflicts.AsError()

	if inst != nil {
		m.byDesc[desc] = inst
	}
	*listPtr = append(*listPtr, inst)
	if err != nil {
		// Handle instrument creation errors when they're new,
		// not for repeat entries above.
		otel.Handle(err)
	}
	return inst, err
}

// synchronousInstrument configures a synchronous instrument.
func (m *meter) synchronousInstrument(name string, cfg instConfig, nk number.Kind, ik sdkinstrument.Kind) (*syncstate.Observer, error) {
	return configureInstrument(m, name, cfg, nk, ik, &m.syncInsts, syncstate.New)
}

// synchronousInstrument configures an asynchronous instrument.
func (m *meter) asynchronousInstrument(name string, cfg instConfig, nk number.Kind, ik sdkinstrument.Kind) (*asyncstate.Observer, error) {
	return configureInstrument(m, name, cfg, nk, ik, &m.asyncInsts, asyncstate.New)
}
