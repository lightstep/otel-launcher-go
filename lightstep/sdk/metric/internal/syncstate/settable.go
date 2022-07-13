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

package syncstate // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/syncstate"

import (
	"context"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
)

// Settable is a synchronous instrument having a Record() method.
type Settable[N number.Any, Traits number.Traits[N]] struct {
	instrument.Synchronous // Note: wasted space

	inst *Instrument
}

// Settable satisfies 2 instrument APIs.
var (
	_ syncint64.SettableGauge           = Settable[int64, number.Int64Traits]{}
	_ syncfloat64.SettableGauge         = Settable[float64, number.Float64Traits]{}
	_ syncint64.SettableUpDownCounter   = Settable[int64, number.Int64Traits]{}
	_ syncfloat64.SettableUpDownCounter = Settable[float64, number.Float64Traits]{}
)

// NewCounter returns a value that implements the Settable API.
func NewSettable[N number.Any, Traits number.Traits[N]](inst *Instrument) Settable[N, Traits] {
	return Settable[N, Traits]{inst: inst}
}

// Set records a Settable observation.
func (h Settable[N, Traits]) Set(ctx context.Context, incr N, attrs ...attribute.KeyValue) {
	capture[N, Traits](ctx, h.inst, incr, attrs)
}
