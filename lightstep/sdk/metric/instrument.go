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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/instrument"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/asyncstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/pipeline"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/syncstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
)

// instrumentConstructor refers to either the syncstate or asyncstate
// NewInstrument method.  Although both receive an opaque interface{}
// to distinguish providers, only the asyncstate package needs to know
// this information.  The unused parameter is passed to the syncstate
// package for the generalization used here to work.
type instrumentConstructor[T any] func(
	instrument sdkinstrument.Descriptor,
	opaque interface{},
	compiled pipeline.Register[viewstate.Instrument],
) T

// configureInstrument applies the instrument configuration, checks
// for an existing definition for the same descriptor, and compiles
// and constructs the instrument if necessary.
func configureInstrument[T any](
	m *meter,
	name string,
	opts []instrument.Option,
	nk number.Kind,
	ik sdkinstrument.Kind,
	listPtr *[]T,
	ctor instrumentConstructor[T],
) (T, error) {
	// Compute the instrument descriptor
	cfg := instrument.NewConfig(opts...)
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

		return lookup.(T), conflicts.AsError()
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

	m.byDesc[desc] = inst
	*listPtr = append(*listPtr, inst)
	if err != nil {
		// Handle instrument creation errors when they're new,
		// not for repeat entries above.
		otel.Handle(err)
	}
	return inst, err
}

// synchronousInstrument configures a synchronous instrument.
func (m *meter) synchronousInstrument(name string, opts []instrument.Option, nk number.Kind, ik sdkinstrument.Kind) (*syncstate.Instrument, error) {
	return configureInstrument(m, name, opts, nk, ik, &m.syncInsts, syncstate.NewInstrument)
}

// synchronousInstrument configures an asynchronous instrument.
func (m *meter) asynchronousInstrument(name string, opts []instrument.Option, nk number.Kind, ik sdkinstrument.Kind) (*asyncstate.Instrument, error) {
	return configureInstrument(m, name, opts, nk, ik, &m.asyncInsts, asyncstate.NewInstrument)
}
