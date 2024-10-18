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

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"go.opentelemetry.io/otel/metric"
)

// Callback is the implementation object associated with one
// asynchronous callback.
type Callback struct {
	// function is the user-provided callback function.
	function metric.Callback

	// instruments are the set of instruments permitted to be used
	// inside this callback.
	instruments map[*Observer]struct{}
}

// NewCallback returns a new Callback; this checks that each of the
// provided instruments belongs to the same meter provider.
func NewCallback(instruments []metric.Observable, opaque interface{}, function metric.Callback) (*Callback, error) {
	if len(instruments) == 0 {
		return nil, fmt.Errorf("asynchronous callback without instruments")
	}
	if function == nil {
		return nil, fmt.Errorf("asynchronous callback with nil function")
	}

	cb := &Callback{
		function:    function,
		instruments: map[*Observer]struct{}{},
	}

	for _, inst := range instruments {
		if unwrapped, ok := inst.(interface {
			Unwrap() metric.Observable
		}); ok {
			inst = unwrapped.Unwrap()
		}
		thisInstImpl, ok := inst.(implementation)
		if !ok {
			return nil, fmt.Errorf("asynchronous instrument does not belong to this SDK: %T", inst)
		}
		thisInst := thisInstImpl.get()
		if thisInst.opaque != opaque {
			return nil, fmt.Errorf("asynchronous instrument belongs to a different meter")
		}

		cb.instruments[thisInst] = struct{}{}
	}

	return cb, nil
}

// Run executes the callback after setting up the appropriate context
// for a specific reader.
func (c *Callback) Run(ctx context.Context, state *State) {
	cp := &callbackState{
		callback: c,
		state:    state,
	}
	c.function(ctx, cp)
	cp.invalidate()
}

// callbackState is used to lookup the current callback and
// pipeline from within an executing callback function.
type callbackState struct {
	metric.Observer

	// lock protects callback, see invalidate() and getCallback()
	lock sync.Mutex

	// callback is the currently running callback; this is set to nil
	// after the associated callback function returns.
	callback *Callback

	// state is a single collection of data.
	state *State
}

func (cs *callbackState) invalidate() {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	cs.callback = nil
}

func (cs *callbackState) getCallback() *Callback {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	return cs.callback
}

func (cs *callbackState) ObserveFloat64(obsrv metric.Float64Observable, value float64, options ...metric.ObserveOption) {
	Observe[float64, number.Float64Traits](obsrv, cs, value, options)
}

func (cs *callbackState) ObserveInt64(obsrv metric.Int64Observable, value int64, options ...metric.ObserveOption) {
	Observe[int64, number.Int64Traits](obsrv, cs, value, options)
}
