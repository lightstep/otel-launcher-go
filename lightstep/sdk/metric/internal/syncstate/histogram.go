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

// Histogram is a synchronous instrument having a Record() method.
type Histogram[N number.Any, Traits number.Traits[N]] struct {
	instrument.Synchronous // Note: wasted space

	inst *Instrument
}

// Histogram satisfies 2 instrument APIs.
var (
	_ syncint64.Histogram   = Histogram[int64, number.Int64Traits]{}
	_ syncfloat64.Histogram = Histogram[float64, number.Float64Traits]{}
)

// NewCounter returns a value that implements the Histogram API.
func NewHistogram[N number.Any, Traits number.Traits[N]](inst *Instrument) Histogram[N, Traits] {
	return Histogram[N, Traits]{inst: inst}
}

// Record records a Histogram observation.
func (h Histogram[N, Traits]) Record(ctx context.Context, incr N, attrs ...attribute.KeyValue) {
	capture[N, Traits](ctx, h.inst, incr, attrs)
}
