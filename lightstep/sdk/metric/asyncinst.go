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
	"go.opentelemetry.io/otel/metric"
)

type (
	int64ObservableCounter struct {
		metric.Int64ObservableCounter
		*asyncstate.Observer
	}
	int64ObservableUpDownCounter struct {
		metric.Int64ObservableUpDownCounter
		*asyncstate.Observer
	}
	int64ObservableGauge struct {
		metric.Int64ObservableGauge
		*asyncstate.Observer
	}
	float64ObservableCounter struct {
		metric.Float64ObservableCounter
		*asyncstate.Observer
	}
	float64ObservableUpDownCounter struct {
		metric.Float64ObservableUpDownCounter
		*asyncstate.Observer
	}
	float64ObservableGauge struct {
		metric.Float64ObservableGauge
		*asyncstate.Observer
	}

	intObserver struct {
		metric.Int64Observer

		observer   metric.Observer
		observable metric.Int64Observable
	}
	floatObserver struct {
		metric.Float64Observer

		observer   metric.Observer
		observable metric.Float64Observable
	}
)

func (io intObserver) Observe(value int64, options ...metric.ObserveOption) {
	io.observer.ObserveInt64(io.observable, value, options...)
}

func (fo floatObserver) Observe(value float64, options ...metric.ObserveOption) {
	fo.observer.ObserveFloat64(fo.observable, value, options...)
}

func registerIntCallbacks[T metric.Int64Observable](m *meter, inst T, cbs []metric.Int64Callback) {
	for _, cb := range cbs {
		_, _ = m.RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
			return cb(ctx, intObserver{
				observer:   obs,
				observable: inst,
			})
		}, inst)
	}
}

func registerFloatCallbacks[T metric.Float64Observable](m *meter, inst T, cbs []metric.Float64Callback) {
	for _, cb := range cbs {
		_, _ = m.RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
			return cb(ctx, floatObserver{
				observer:   obs,
				observable: inst,
			})
		}, inst)
	}
}

func (m *meter) Int64ObservableCounter(name string, opts ...metric.Int64ObservableCounterOption) (metric.Int64ObservableCounter, error) {
	cfg := metric.NewInt64ObservableCounterConfig(opts...)
	impl, err := m.asynchronousInstrument(name, cfg, number.Int64Kind, sdkinstrument.AsyncCounter)
	inst := int64ObservableCounter{
		Observer: impl,
	}
	registerIntCallbacks(m, inst, cfg.Callbacks())
	return inst, err
}

func (m *meter) Int64ObservableUpDownCounter(name string, opts ...metric.Int64ObservableUpDownCounterOption) (metric.Int64ObservableUpDownCounter, error) {
	cfg := metric.NewInt64ObservableUpDownCounterConfig(opts...)
	impl, err := m.asynchronousInstrument(name, cfg, number.Int64Kind, sdkinstrument.AsyncUpDownCounter)
	inst := int64ObservableUpDownCounter{
		Observer: impl,
	}
	registerIntCallbacks(m, inst, cfg.Callbacks())
	return inst, err
}

func (m *meter) Int64ObservableGauge(name string, opts ...metric.Int64ObservableGaugeOption) (metric.Int64ObservableGauge, error) {
	cfg := metric.NewInt64ObservableGaugeConfig(opts...)
	impl, err := m.asynchronousInstrument(name, cfg, number.Int64Kind, sdkinstrument.AsyncGauge)
	inst := int64ObservableGauge{
		Observer: impl,
	}
	registerIntCallbacks(m, inst, cfg.Callbacks())
	return inst, err
}

func (m *meter) Float64ObservableCounter(name string, opts ...metric.Float64ObservableCounterOption) (metric.Float64ObservableCounter, error) {
	cfg := metric.NewFloat64ObservableCounterConfig(opts...)
	impl, err := m.asynchronousInstrument(name, cfg, number.Float64Kind, sdkinstrument.AsyncCounter)
	inst := float64ObservableCounter{
		Observer: impl,
	}
	registerFloatCallbacks(m, inst, cfg.Callbacks())
	return inst, err
}

func (m *meter) Float64ObservableUpDownCounter(name string, opts ...metric.Float64ObservableUpDownCounterOption) (metric.Float64ObservableUpDownCounter, error) {
	cfg := metric.NewFloat64ObservableUpDownCounterConfig(opts...)
	impl, err := m.asynchronousInstrument(name, cfg, number.Float64Kind, sdkinstrument.AsyncUpDownCounter)
	inst := float64ObservableUpDownCounter{
		Observer: impl,
	}
	registerFloatCallbacks(m, inst, cfg.Callbacks())
	return inst, err
}

func (m *meter) Float64ObservableGauge(name string, opts ...metric.Float64ObservableGaugeOption) (metric.Float64ObservableGauge, error) {
	cfg := metric.NewFloat64ObservableGaugeConfig(opts...)
	impl, err := m.asynchronousInstrument(name, cfg, number.Float64Kind, sdkinstrument.AsyncGauge)
	inst := float64ObservableGauge{
		Observer: impl,
	}
	registerFloatCallbacks(m, inst, cfg.Callbacks())
	return inst, err
}
