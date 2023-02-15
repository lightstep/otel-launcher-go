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
	"context"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/asyncstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
)

type (
	int64ObservableCounter struct {
		instrument.Int64ObservableCounter
		*asyncstate.Observer
	}
	int64ObservableUpDownCounter struct {
		instrument.Int64ObservableUpDownCounter
		*asyncstate.Observer
	}
	int64ObservableGauge struct {
		instrument.Int64ObservableGauge
		*asyncstate.Observer
	}
	float64ObservableCounter struct {
		instrument.Float64ObservableCounter
		*asyncstate.Observer
	}
	float64ObservableUpDownCounter struct {
		instrument.Float64ObservableUpDownCounter
		*asyncstate.Observer
	}
	float64ObservableGauge struct {
		instrument.Float64ObservableGauge
		*asyncstate.Observer
	}

	intObserver struct {
		observer   metric.Observer
		observable instrument.Int64Observable
	}
	floatObserver struct {
		observer   metric.Observer
		observable instrument.Float64Observable
	}
)

func (io intObserver) Observe(value int64, attrs ...attribute.KeyValue) {
	io.observer.ObserveInt64(io.observable, value, attrs...)
}

func (fo floatObserver) Observe(value float64, attrs ...attribute.KeyValue) {
	fo.observer.ObserveFloat64(fo.observable, value, attrs...)
}

func registerIntCallbacks[T instrument.Int64Observable](m *meter, inst T, cbs []instrument.Int64Callback) {
	for _, cb := range cbs {
		_, _ = m.RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
			return cb(ctx, intObserver{
				observer:   obs,
				observable: inst,
			})
		}, inst)
	}
}

func registerFloatCallbacks[T instrument.Float64Observable](m *meter, inst T, cbs []instrument.Float64Callback) {
	for _, cb := range cbs {
		_, _ = m.RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
			return cb(ctx, floatObserver{
				observer:   obs,
				observable: inst,
			})
		}, inst)
	}
}

func (m *meter) Int64ObservableCounter(name string, opts ...instrument.Int64ObserverOption) (instrument.Int64ObservableCounter, error) {
	cfg := instrument.NewInt64ObserverConfig(opts...)
	impl, err := m.asynchronousInstrument(name, cfg, number.Int64Kind, sdkinstrument.AsyncCounter)
	inst := int64ObservableCounter{
		Observer: impl,
	}
	registerIntCallbacks(m, inst, cfg.Callbacks())
	return inst, err
}

func (m *meter) Int64ObservableUpDownCounter(name string, opts ...instrument.Int64ObserverOption) (instrument.Int64ObservableUpDownCounter, error) {
	cfg := instrument.NewInt64ObserverConfig(opts...)
	impl, err := m.asynchronousInstrument(name, cfg, number.Int64Kind, sdkinstrument.AsyncUpDownCounter)
	inst := int64ObservableUpDownCounter{
		Observer: impl,
	}
	registerIntCallbacks(m, inst, cfg.Callbacks())
	return inst, err
}

func (m *meter) Int64ObservableGauge(name string, opts ...instrument.Int64ObserverOption) (instrument.Int64ObservableGauge, error) {
	cfg := instrument.NewInt64ObserverConfig(opts...)
	impl, err := m.asynchronousInstrument(name, cfg, number.Int64Kind, sdkinstrument.AsyncGauge)
	inst := int64ObservableGauge{
		Observer: impl,
	}
	registerIntCallbacks(m, inst, cfg.Callbacks())
	return inst, err
}

func (m *meter) Float64ObservableCounter(name string, opts ...instrument.Float64ObserverOption) (instrument.Float64ObservableCounter, error) {
	cfg := instrument.NewFloat64ObserverConfig(opts...)
	impl, err := m.asynchronousInstrument(name, cfg, number.Float64Kind, sdkinstrument.AsyncCounter)
	inst := float64ObservableCounter{
		Observer: impl,
	}
	registerFloatCallbacks(m, inst, cfg.Callbacks())
	return inst, err
}

func (m *meter) Float64ObservableUpDownCounter(name string, opts ...instrument.Float64ObserverOption) (instrument.Float64ObservableUpDownCounter, error) {
	cfg := instrument.NewFloat64ObserverConfig(opts...)
	impl, err := m.asynchronousInstrument(name, cfg, number.Float64Kind, sdkinstrument.AsyncUpDownCounter)
	inst := float64ObservableUpDownCounter{
		Observer: impl,
	}
	registerFloatCallbacks(m, inst, cfg.Callbacks())
	return inst, err
}

func (m *meter) Float64ObservableGauge(name string, opts ...instrument.Float64ObserverOption) (instrument.Float64ObservableGauge, error) {
	cfg := instrument.NewFloat64ObserverConfig(opts...)
	impl, err := m.asynchronousInstrument(name, cfg, number.Float64Kind, sdkinstrument.AsyncGauge)
	inst := float64ObservableGauge{
		Observer: impl,
	}
	registerFloatCallbacks(m, inst, cfg.Callbacks())
	return inst, err
}
