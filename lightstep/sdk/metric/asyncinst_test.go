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

package metric

import (
	"context"
	"testing"
	"time"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/gauge"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/sum"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/test"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/resource"
)

func TestAsyncInsts(t *testing.T) {
	rdr := NewManualReader("test")
	res := resource.Empty()
	provider := NewMeterProvider(WithReader(rdr), WithResource(res))

	ci := must(provider.Meter("test").AsyncInt64().Counter("icount"))
	cf := must(provider.Meter("test").AsyncFloat64().Counter("fcount"))
	ui := must(provider.Meter("test").AsyncInt64().UpDownCounter("iupcount"))
	uf := must(provider.Meter("test").AsyncFloat64().UpDownCounter("fupcount"))
	gi := must(provider.Meter("test").AsyncInt64().Gauge("igauge"))
	gf := must(provider.Meter("test").AsyncFloat64().Gauge("fgauge"))

	attr := attribute.String("a", "B")

	_ = provider.Meter("test").RegisterCallback([]instrument.Asynchronous{
		ci, cf, ui, uf, gi, gf,
	}, func(ctx context.Context) {
		ci.Observe(ctx, 2, attr)
		cf.Observe(ctx, 3, attr)
		ui.Observe(ctx, 4, attr)
		uf.Observe(ctx, 5, attr)
		gi.Observe(ctx, 6, attr)
		gf.Observe(ctx, 7, attr)
	})

	data := rdr.Produce(nil)
	notime := time.Time{}
	cumulative := aggregation.CumulativeTemporality

	test.RequireEqualResourceMetrics(
		t, data, res,
		test.Scope(
			test.Library("test"),
			test.Instrument(
				test.Descriptor("icount", sdkinstrument.AsyncCounter, number.Int64Kind),
				test.Point(notime, notime, sum.NewMonotonicInt64(2), cumulative, attr),
			),
			test.Instrument(
				test.Descriptor("fcount", sdkinstrument.AsyncCounter, number.Float64Kind),
				test.Point(notime, notime, sum.NewMonotonicFloat64(3), cumulative, attr),
			),
			test.Instrument(
				test.Descriptor("iupcount", sdkinstrument.AsyncUpDownCounter, number.Int64Kind),
				test.Point(notime, notime, sum.NewNonMonotonicInt64(4), cumulative, attr),
			),
			test.Instrument(
				test.Descriptor("fupcount", sdkinstrument.AsyncUpDownCounter, number.Float64Kind),
				test.Point(notime, notime, sum.NewNonMonotonicFloat64(5), cumulative, attr),
			),
			test.Instrument(
				test.Descriptor("igauge", sdkinstrument.AsyncGauge, number.Int64Kind),
				test.Point(notime, notime, gauge.NewInt64(6), cumulative, attr),
			),
			test.Instrument(
				test.Descriptor("fgauge", sdkinstrument.AsyncGauge, number.Float64Kind),
				test.Point(notime, notime, gauge.NewFloat64(7), cumulative, attr),
			),
		),
	)
}
