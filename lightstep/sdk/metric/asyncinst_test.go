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
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func TestAsyncInstsMultiCallback(t *testing.T) {
	rdr := NewManualReader("test")
	res := resource.Empty()
	provider := NewMeterProvider(WithReader(rdr), WithResource(res))

	ci := must(provider.Meter("test").Int64ObservableCounter("icount"))
	cf := must(provider.Meter("test").Float64ObservableCounter("fcount"))
	ui := must(provider.Meter("test").Int64ObservableUpDownCounter("iupcount"))
	uf := must(provider.Meter("test").Float64ObservableUpDownCounter("fupcount"))
	gi := must(provider.Meter("test").Int64ObservableGauge("igauge"))
	gf := must(provider.Meter("test").Float64ObservableGauge("fgauge"))

	attr := attribute.String("a", "B")

	reg, err := provider.Meter("test").RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		observer.ObserveInt64(ci, 2, metric.WithAttributes(attr))
		observer.ObserveFloat64(cf, 3, metric.WithAttributes(attr))
		observer.ObserveInt64(ui, 4, metric.WithAttributes(attr))
		observer.ObserveFloat64(uf, 5, metric.WithAttributes(attr))
		observer.ObserveInt64(gi, 6, metric.WithAttributes(attr))
		observer.ObserveFloat64(gf, 7, metric.WithAttributes(attr))
		return nil
	}, ci, cf, ui, uf, gi, gf)

	require.NoError(t, err)

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

	// Unregister it, get no points in the following collection.
	require.NoError(t, reg.Unregister())

	data = rdr.Produce(nil)
	test.RequireEqualResourceMetrics(
		t, data, res,
		test.Scope(
			test.Library("test"),
			test.Instrument(
				test.Descriptor("icount", sdkinstrument.AsyncCounter, number.Int64Kind),
			),
			test.Instrument(
				test.Descriptor("fcount", sdkinstrument.AsyncCounter, number.Float64Kind),
			),
			test.Instrument(
				test.Descriptor("iupcount", sdkinstrument.AsyncUpDownCounter, number.Int64Kind),
			),
			test.Instrument(
				test.Descriptor("fupcount", sdkinstrument.AsyncUpDownCounter, number.Float64Kind),
			),
			test.Instrument(
				test.Descriptor("igauge", sdkinstrument.AsyncGauge, number.Int64Kind),
			),
			test.Instrument(
				test.Descriptor("fgauge", sdkinstrument.AsyncGauge, number.Float64Kind),
			),
		),
	)

	// Unregister it again, get an error.
	err = reg.Unregister()
	require.Error(t, err)
	require.Contains(t, err.Error(), "already unregistered")
}

func TestAsyncInstsSingleCallback(t *testing.T) {
	rdr := NewManualReader("test")
	res := resource.Empty()
	provider := NewMeterProvider(WithReader(rdr), WithResource(res))
	tm := provider.Meter("test")

	attr := attribute.String("a", "B")

	_ = must(tm.Int64ObservableCounter("icount",
		metric.WithInt64Callback(
			func(ctx context.Context, obs metric.Int64Observer) error {
				obs.Observe(2, metric.WithAttributes(attr))
				return nil
			},
		),
	))
	_ = must(tm.Float64ObservableCounter("fcount",
		metric.WithFloat64Callback(
			func(ctx context.Context, obs metric.Float64Observer) error {
				obs.Observe(3, metric.WithAttributes(attr))
				return nil
			},
		),
	))
	_ = must(tm.Int64ObservableUpDownCounter("iupcount",
		metric.WithInt64Callback(
			func(ctx context.Context, obs metric.Int64Observer) error {
				obs.Observe(4, metric.WithAttributes(attr))
				return nil
			},
		),
	))
	_ = must(tm.Float64ObservableUpDownCounter("fupcount",
		metric.WithFloat64Callback(
			func(ctx context.Context, obs metric.Float64Observer) error {
				obs.Observe(5, metric.WithAttributes(attr))
				return nil
			},
		),
	))
	_ = must(tm.Int64ObservableGauge("igauge",
		metric.WithInt64Callback(
			func(ctx context.Context, obs metric.Int64Observer) error {
				obs.Observe(6, metric.WithAttributes(attr))
				return nil
			},
		),
	))
	_ = must(tm.Float64ObservableGauge("fgauge",
		metric.WithFloat64Callback(
			func(ctx context.Context, obs metric.Float64Observer) error {
				obs.Observe(7, metric.WithAttributes(attr))
				return nil
			},
		),
	))

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
