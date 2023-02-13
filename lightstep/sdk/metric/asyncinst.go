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
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/asyncstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel/metric/instrument"
)

func (m *meter) Int64ObservableCounter(name string, opts ...instrument.Int64ObserverOption) (instrument.Int64ObservableCounter, error) {
	inst, err := m.asynchronousInstrument(name, instrument.NewInt64ObserverConfig(opts...), number.Int64Kind, sdkinstrument.AsyncCounter)
	return asyncstate.NewObserver[int64, number.Int64Traits](inst), err
}

func (m *meter) Int64ObservableUpDownCounter(name string, opts ...instrument.Int64ObserverOption) (instrument.Int64ObservableUpDownCounter, error) {
	inst, err := m.asynchronousInstrument(name, instrument.NewInt64ObserverConfig(opts...), number.Int64Kind, sdkinstrument.AsyncUpDownCounter)
	return asyncstate.NewObserver[int64, number.Int64Traits](inst), err
}

func (m *meter) Int64ObservableGauge(name string, opts ...instrument.Int64ObserverOption) (instrument.Int64ObservableGauge, error) {
	inst, err := m.asynchronousInstrument(name, instrument.NewInt64ObserverConfig(opts...), number.Int64Kind, sdkinstrument.AsyncGauge)
	return asyncstate.NewObserver[int64, number.Int64Traits](inst), err
}

func (m *meter) Float64ObservableCounter(name string, opts ...instrument.Float64ObserverOption) (instrument.Float64ObservableCounter, error) {
	inst, err := m.asynchronousInstrument(name, instrument.NewFloat64ObserverConfig(opts...), number.Float64Kind, sdkinstrument.AsyncCounter)
	return asyncstate.NewObserver[float64, number.Float64Traits](inst), err
}

func (m *meter) Float64ObservableUpDownCounter(name string, opts ...instrument.Float64ObserverOption) (instrument.Float64ObservableUpDownCounter, error) {
	inst, err := m.asynchronousInstrument(name, instrument.NewFloat64ObserverConfig(opts...), number.Float64Kind, sdkinstrument.AsyncUpDownCounter)
	return asyncstate.NewObserver[float64, number.Float64Traits](inst), err
}

func (m *meter) Float64ObservableGauge(name string, opts ...instrument.Float64ObserverOption) (instrument.Float64ObservableGauge, error) {
	inst, err := m.asynchronousInstrument(name, instrument.NewFloat64ObserverConfig(opts...), number.Float64Kind, sdkinstrument.AsyncGauge)
	return asyncstate.NewObserver[float64, number.Float64Traits](inst), err
}
