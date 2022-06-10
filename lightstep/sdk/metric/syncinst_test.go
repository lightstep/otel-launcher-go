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

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/sum"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/test"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

func TestSyncInsts(t *testing.T) {
	cfg := aggregator.Config{}
	// A histogram with size 4 ensures the test data has 3 buckets @ scale 0
	cfg.Histogram.MaxSize = 4

	ctx := context.Background()
	rdr := NewManualReader("test")
	res := resource.Empty()
	provider := NewMeterProvider(
		WithResource(res),
		WithReader(
			rdr,
			view.WithDefaultAggregationConfigSelector(
				func(sdkinstrument.Kind) (int64Config, float64Config aggregator.Config) {
					return cfg, cfg
				},
			),
		),
	)

	ci := must(provider.Meter("test").SyncInt64().Counter("icount"))
	cf := must(provider.Meter("test").SyncFloat64().Counter("fcount"))
	ui := must(provider.Meter("test").SyncInt64().UpDownCounter("iupcount"))
	uf := must(provider.Meter("test").SyncFloat64().UpDownCounter("fupcount"))
	hi := must(provider.Meter("test").SyncInt64().Histogram("ihistogram"))
	hf := must(provider.Meter("test").SyncFloat64().Histogram("fhistogram"))

	attr := attribute.String("a", "B")

	ci.Add(ctx, 2, attr)
	cf.Add(ctx, 3, attr)
	ui.Add(ctx, 4, attr)
	uf.Add(ctx, 5, attr)

	hi.Record(ctx, 2, attr)
	hi.Record(ctx, 4, attr)
	hi.Record(ctx, 8, attr)

	hf.Record(ctx, 8, attr)
	hf.Record(ctx, 16, attr)
	hf.Record(ctx, 32, attr)

	data := rdr.Produce(nil)
	notime := time.Time{}
	cumulative := aggregation.CumulativeTemporality

	test.RequireEqualResourceMetrics(
		t, data, res,
		test.Scope(
			test.Library("test"),
			test.Instrument(
				test.Descriptor("icount", sdkinstrument.SyncCounter, number.Int64Kind),
				test.Point(notime, notime, sum.NewMonotonicInt64(2), cumulative, attr),
			),
			test.Instrument(
				test.Descriptor("fcount", sdkinstrument.SyncCounter, number.Float64Kind),
				test.Point(notime, notime, sum.NewMonotonicFloat64(3), cumulative, attr),
			),
			test.Instrument(
				test.Descriptor("iupcount", sdkinstrument.SyncUpDownCounter, number.Int64Kind),
				test.Point(notime, notime, sum.NewNonMonotonicInt64(4), cumulative, attr),
			),
			test.Instrument(
				test.Descriptor("fupcount", sdkinstrument.SyncUpDownCounter, number.Float64Kind),
				test.Point(notime, notime, sum.NewNonMonotonicFloat64(5), cumulative, attr),
			),
			test.Instrument(
				test.Descriptor("ihistogram", sdkinstrument.SyncHistogram, number.Int64Kind),
				test.Point(notime, notime, histogram.NewInt64(cfg.Histogram, 2, 4, 8), cumulative, attr),
			),
			test.Instrument(
				test.Descriptor("fhistogram", sdkinstrument.SyncHistogram, number.Float64Kind),
				test.Point(notime, notime, histogram.NewFloat64(cfg.Histogram, 8, 16, 32), cumulative, attr),
			),
		),
	)
}
