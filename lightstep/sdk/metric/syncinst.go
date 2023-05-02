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

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/syncstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/embedded"
)

type (
	int64Counter struct {
		embedded.Int64Counter
		observer *syncstate.Observer
	}
	int64UpDownCounter struct {
		embedded.Int64UpDownCounter
		observer *syncstate.Observer
	}
	int64Histogram struct {
		embedded.Int64Histogram
		observer *syncstate.Observer
	}
	float64Counter struct {
		embedded.Float64Counter
		observer *syncstate.Observer
	}
	float64UpDownCounter struct {
		embedded.Float64UpDownCounter
		observer *syncstate.Observer
	}
	float64Histogram struct {
		embedded.Float64Histogram
		observer *syncstate.Observer
	}
)

func (i int64Counter) Add(ctx context.Context, value int64, options ...metric.AddOption) {
	i.observer.ObserveInt64(ctx, value, options...)
}

func (i int64UpDownCounter) Add(ctx context.Context, value int64, options ...metric.AddOption) {
	i.observer.ObserveInt64(ctx, value, options...)
}

func (i int64Histogram) Record(ctx context.Context, value int64, options ...metric.AddOption) {
	i.observer.ObserveInt64(ctx, value, options...)
}

func (i float64Counter) Add(ctx context.Context, value float64, options ...metric.AddOption) {
	i.observer.ObserveFloat64(ctx, value, options...)
}

func (i float64UpDownCounter) Add(ctx context.Context, value float64, options ...metric.AddOption) {
	i.observer.ObserveFloat64(ctx, value, options...)
}

func (i float64Histogram) Record(ctx context.Context, value float64, options ...metric.AddOption) {
	i.observer.ObserveFloat64(ctx, value, options...)
}

func (m *meter) Int64Counter(name string, opts ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	inst, err := m.synchronousInstrument(name, metric.NewInt64CounterConfig(opts...), number.Int64Kind, sdkinstrument.SyncCounter)
	return int64Counter{observer: inst}, err
}

func (m *meter) Int64UpDownCounter(name string, opts ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error) {
	inst, err := m.synchronousInstrument(name, metric.NewInt64UpDownCounterConfig(opts...), number.Int64Kind, sdkinstrument.SyncUpDownCounter)
	return int64UpDownCounter{observer: inst}, err
}

func (m *meter) Int64Histogram(name string, opts ...metric.Int64HistogramOption) (metric.Int64Histogram, error) {
	inst, err := m.synchronousInstrument(name, metric.NewInt64Config(opts...), number.Int64Kind, sdkinstrument.SyncHistogram)
	return int64Histogram{observer: inst}, err
}

func (m *meter) Float64Counter(name string, opts ...metric.Float64CounterOption) (metric.Float64Counter, error) {
	inst, err := m.synchronousInstrument(name, metric.NewFloat64Config(opts...), number.Float64Kind, sdkinstrument.SyncCounter)
	return float64Counter{observer: inst}, err
}

func (m *meter) Float64UpDownCounter(name string, opts ...metric.Float64UpDownCounterOption) (metric.Float64UpDownCounter, error) {
	inst, err := m.synchronousInstrument(name, metric.NewFloat64Config(opts...), number.Float64Kind, sdkinstrument.SyncUpDownCounter)
	return float64UpDownCounter{observer: inst}, err
}

func (m *meter) Float64Histogram(name string, opts ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	inst, err := m.synchronousInstrument(name, metric.NewFloat64Config(opts...), number.Float64Kind, sdkinstrument.SyncHistogram)
	return float64Histogram{observer: inst}, err
}
