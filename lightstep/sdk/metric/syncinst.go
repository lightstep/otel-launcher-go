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
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
)

type (
	syncint64Instruments   struct{ *meter }
	syncfloat64Instruments struct{ *meter }
)

func (i syncint64Instruments) Counter(name string, opts ...instrument.Option) (syncint64.Counter, error) {
	inst, err := i.synchronousInstrument(name, opts, number.Int64Kind, sdkinstrument.SyncCounter)
	return syncstate.NewCounter[int64, number.Int64Traits](inst), err
}

func (i syncint64Instruments) UpDownCounter(name string, opts ...instrument.Option) (syncint64.UpDownCounter, error) {
	inst, err := i.synchronousInstrument(name, opts, number.Int64Kind, sdkinstrument.SyncUpDownCounter)
	return syncstate.NewCounter[int64, number.Int64Traits](inst), err
}

func (i syncint64Instruments) Histogram(name string, opts ...instrument.Option) (syncint64.Histogram, error) {
	inst, err := i.synchronousInstrument(name, opts, number.Int64Kind, sdkinstrument.SyncHistogram)
	return syncstate.NewHistogram[int64, number.Int64Traits](inst), err
}

func (f syncfloat64Instruments) Counter(name string, opts ...instrument.Option) (syncfloat64.Counter, error) {
	inst, err := f.synchronousInstrument(name, opts, number.Float64Kind, sdkinstrument.SyncCounter)
	return syncstate.NewCounter[float64, number.Float64Traits](inst), err
}

func (f syncfloat64Instruments) UpDownCounter(name string, opts ...instrument.Option) (syncfloat64.UpDownCounter, error) {
	inst, err := f.synchronousInstrument(name, opts, number.Float64Kind, sdkinstrument.SyncUpDownCounter)
	return syncstate.NewCounter[float64, number.Float64Traits](inst), err
}

func (f syncfloat64Instruments) Histogram(name string, opts ...instrument.Option) (syncfloat64.Histogram, error) {
	inst, err := f.synchronousInstrument(name, opts, number.Float64Kind, sdkinstrument.SyncHistogram)
	return syncstate.NewHistogram[float64, number.Float64Traits](inst), err
}
