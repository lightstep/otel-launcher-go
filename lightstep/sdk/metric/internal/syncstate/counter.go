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

// Counter is a synchronous instrument having an Add() method.
type Counter[N number.Any, Traits number.Traits[N]] struct {
	instrument.Synchronous // Note: wasted space

	inst *Instrument
}

// Counter satisfies 4 instrument APIs.
var (
	_ syncint64.Counter         = Counter[int64, number.Int64Traits]{}
	_ syncint64.UpDownCounter   = Counter[int64, number.Int64Traits]{}
	_ syncfloat64.Counter       = Counter[float64, number.Float64Traits]{}
	_ syncfloat64.UpDownCounter = Counter[float64, number.Float64Traits]{}
)

// NewCounter returns a value that implements the Counter and UpDownCounter APIs.
func NewCounter[N number.Any, Traits number.Traits[N]](inst *Instrument) Counter[N, Traits] {
	return Counter[N, Traits]{inst: inst}
}

// Add increments a Counter or UpDownCounter.
func (c Counter[N, Traits]) Add(ctx context.Context, incr N, attrs ...attribute.KeyValue) {
	capture[N, Traits](ctx, c.inst, incr, attrs)
}
