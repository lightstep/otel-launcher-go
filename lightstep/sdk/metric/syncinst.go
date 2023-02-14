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
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/syncstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel/metric/instrument"
)

type (
	int64Counter struct {
		instrument.Int64Counter
		syncstate.Instrument
	}
	int64UpDownCounter struct {
		instrument.Int64UpDownCounter
		syncstate.Instrument
	}
	int64Histogram struct {
		instrument.Int64Histogram
		syncstate.Instrument
	}
	float64Counter struct {
		instrument.Float64Counter
		syncstate.Instrument
	}
	float64UpDownCounter struct {
		instrument.Float64UpDownCounter
		syncstate.Instrument
	}
	float64Histogram struct {
		instrument.Float64Histogram
		syncstate.Instrument
	}
)

func (m *meter) Int64Counter(name string, opts ...instrument.Int64Option) (instrument.Int64Counter, error) {
	inst, err := m.synchronousInstrument(name, instrument.NewInt64Config(opts...), number.Int64Kind, sdkinstrument.SyncCounter)
	return int64Counter{Instrument: inst}, err
}

func (m *meter) Int64UpDownCounter(name string, opts ...instrument.Int64Option) (instrument.Int64UpDownCounter, error) {
	inst, err := m.synchronousInstrument(name, instrument.NewInt64Config(opts...), number.Int64Kind, sdkinstrument.SyncUpDownCounter)
	return int64UpDownCounter{Instrument: inst}, err
}

func (m *meter) Int64Histogram(name string, opts ...instrument.Int64Option) (instrument.Int64Histogram, error) {
	inst, err := m.synchronousInstrument(name, instrument.NewInt64Config(opts...), number.Int64Kind, sdkinstrument.SyncHistogram)
	return int64Histogram{Instrument: inst}, err
}

func (m *meter) Float64Counter(name string, opts ...instrument.Float64Option) (instrument.Float64Counter, error) {
	inst, err := m.synchronousInstrument(name, instrument.NewFloat64Config(opts...), number.Float64Kind, sdkinstrument.SyncCounter)
	return float64Counter{Instrument: inst}, err
}

func (m *meter) Float64UpDownCounter(name string, opts ...instrument.Float64Option) (instrument.Float64UpDownCounter, error) {
	inst, err := m.synchronousInstrument(name, instrument.NewFloat64Config(opts...), number.Float64Kind, sdkinstrument.SyncUpDownCounter)
	return float64UpDownCounter{Instrument: inst}, err
}

func (m *meter) Float64Histogram(name string, opts ...instrument.Float64Option) (instrument.Float64Histogram, error) {
	inst, err := m.synchronousInstrument(name, instrument.NewFloat64Config(opts...), number.Float64Kind, sdkinstrument.SyncHistogram)
	return float64Histogram{Instrument: inst}, err
}
