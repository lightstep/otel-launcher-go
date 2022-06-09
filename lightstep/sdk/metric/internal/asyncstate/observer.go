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

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
)

// Observer is a generic (int64 or float64) instrument which
// satisfies any of the asynchronous instrument API interfaces.
type Observer[N number.Any, Traits number.Traits[N]] struct {
	instrument.Asynchronous // Note: wasted space

	inst *Instrument
}

// Observer implements 6 instruments and memberInstrument.
var (
	_ asyncint64.Counter       = Observer[int64, number.Int64Traits]{}
	_ asyncint64.UpDownCounter = Observer[int64, number.Int64Traits]{}
	_ asyncint64.Gauge         = Observer[int64, number.Int64Traits]{}
	_ memberInstrument         = Observer[int64, number.Int64Traits]{}

	_ asyncfloat64.Counter       = Observer[float64, number.Float64Traits]{}
	_ asyncfloat64.UpDownCounter = Observer[float64, number.Float64Traits]{}
	_ asyncfloat64.Gauge         = Observer[float64, number.Float64Traits]{}
	_ memberInstrument           = Observer[float64, number.Float64Traits]{}
)

// memberInstrument indicates whether a user-provided
// instrument was returned by this SDK.
type memberInstrument interface {
	instrument() *Instrument
}

// NewObserver returns an generic value suitable for use as any of the
// asynchronous instrument APIs.
func NewObserver[N number.Any, Traits number.Traits[N]](inst *Instrument) Observer[N, Traits] {
	return Observer[N, Traits]{inst: inst}
}

func (o Observer[N, Traits]) instrument() *Instrument {
	return o.inst
}

func (o Observer[N, Traits]) Observe(ctx context.Context, value N, attrs ...attribute.KeyValue) {
	capture[N, Traits](ctx, o.inst, value, attrs)
}
