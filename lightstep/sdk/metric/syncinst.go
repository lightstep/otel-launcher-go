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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
)

type (
	int64AnyCounter struct {
		*syncstate.Observer
	}
	int64Histogram struct {
		*syncstate.Observer
	}
	float64AnyCounter struct {
		*syncstate.Observer
	}
	float64Histogram struct {
		*syncstate.Observer
	}
)

func (i int64AnyCounter) Add(ctx context.Context, value int64, attrs ...attribute.KeyValue) {
	i.ObserveInt64(ctx, value, attrs...)
}

func (i int64Histogram) Record(ctx context.Context, value int64, attrs ...attribute.KeyValue) {
	i.ObserveInt64(ctx, value, attrs...)
}

func (i float64AnyCounter) Add(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	i.ObserveFloat64(ctx, value, attrs...)
}

func (i float64Histogram) Record(ctx context.Context, value float64, attrs ...attribute.KeyValue) {
	i.ObserveFloat64(ctx, value, attrs...)
}

func (m *meter) Int64Counter(name string, opts ...instrument.Int64Option) (instrument.Int64Counter, error) {
	inst, err := m.synchronousInstrument(name, instrument.NewInt64Config(opts...), number.Int64Kind, sdkinstrument.SyncCounter)
	return int64AnyCounter{Observer: inst}, err
}

func (m *meter) Int64UpDownCounter(name string, opts ...instrument.Int64Option) (instrument.Int64UpDownCounter, error) {
	inst, err := m.synchronousInstrument(name, instrument.NewInt64Config(opts...), number.Int64Kind, sdkinstrument.SyncUpDownCounter)
	return int64AnyCounter{Observer: inst}, err
}

func (m *meter) Int64Histogram(name string, opts ...instrument.Int64Option) (instrument.Int64Histogram, error) {
	inst, err := m.synchronousInstrument(name, instrument.NewInt64Config(opts...), number.Int64Kind, sdkinstrument.SyncHistogram)
	return int64Histogram{Observer: inst}, err
}

func (m *meter) Float64Counter(name string, opts ...instrument.Float64Option) (instrument.Float64Counter, error) {
	inst, err := m.synchronousInstrument(name, instrument.NewFloat64Config(opts...), number.Float64Kind, sdkinstrument.SyncCounter)
	return float64AnyCounter{Observer: inst}, err
}

func (m *meter) Float64UpDownCounter(name string, opts ...instrument.Float64Option) (instrument.Float64UpDownCounter, error) {
	inst, err := m.synchronousInstrument(name, instrument.NewFloat64Config(opts...), number.Float64Kind, sdkinstrument.SyncUpDownCounter)
	return float64AnyCounter{Observer: inst}, err
}

func (m *meter) Float64Histogram(name string, opts ...instrument.Float64Option) (instrument.Float64Histogram, error) {
	inst, err := m.synchronousInstrument(name, instrument.NewFloat64Config(opts...), number.Float64Kind, sdkinstrument.SyncHistogram)
	return float64Histogram{Observer: inst}, err
}
