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

package metrictransform

import (
	"math"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/gauge"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/minmaxsumcount"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/sum"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp/internal/otlptest"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/test"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/resource"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/testing/protocmp"
)

const (
	testCumulative = aggregation.CumulativeTemporality
	testDelta      = aggregation.DeltaTemporality
	testDontCare   = aggregation.UndefinedTemporality

	expectCumulative = metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE
	expectDelta      = metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA

	testName   = "it"
	testUnit   = "tu"
	testDesc   = "td"
	testSchema = "tschema"
	noSchema   = ""
)

var (
	testResource0   = resource.NewWithAttributes(testSchema)
	expectResource0 = otlptest.Resource()
	testResource1   = resource.NewSchemaless(attribute.String("Res", "Val"))
	expectResource1 = otlptest.Resource(otlptest.KeyValue("Res", "Val"))
	testResource2   = resource.NewWithAttributes(
		testSchema,
		attribute.String("A", "B"),
		attribute.String("C", "D"),
	)
	expectResource2 = otlptest.Resource(otlptest.KeyValue("A", "B"), otlptest.KeyValue("C", "D"))

	testScope0   = test.Library("lib0", metric.WithInstrumentationVersion("v1.2.3"))
	expectScope0 = otlptest.Scope("lib0", "v1.2.3")

	testScope1   = test.Library("lib1")
	expectScope1 = otlptest.Scope("lib1", "")

	testAttrs0 = []attribute.KeyValue{attribute.String("K", "V")}
	testAttrs1 = []attribute.KeyValue{attribute.String("K0", "0"), attribute.Int("K1", 1)}

	expectAttrs0 = otlptest.Attributes(otlptest.KeyValue("K", "V"))
	expectAttrs1 = otlptest.Attributes(otlptest.KeyValue("K0", "0"), otlptest.KeyIntValue("K1", 1))

	noTime     = time.Time{}
	endTime    = time.Now()
	middleTime = endTime.Add(-time.Millisecond)
	startTime  = endTime.Add(-2 * time.Millisecond)
)

func testInst(nk number.Kind) sdkinstrument.Descriptor {
	kind := sdkinstrument.Kind(-1) // the exporter doesn't inspect this field, test doesn't care
	return test.Descriptor(testName, kind, nk, instrument.WithUnit(testUnit), instrument.WithDescription(testDesc))
}

func testInt64() sdkinstrument.Descriptor {
	return testInst(number.Int64Kind)
}

func testFloat64() sdkinstrument.Descriptor {
	return testInst(number.Float64Kind)
}

func TestMetricTransform(t *testing.T) {
	for _, test := range []struct {
		input   data.Metrics
		encoded *metricspb.ResourceMetrics
	}{
		// int64 counter, resource0, scope0, attrs0, cumulative
		{
			input: test.Metrics(
				testResource0,
				test.Scope(
					testScope0,
					test.Instrument(
						testInt64(),
						test.Point(startTime, endTime, sum.NewMonotonicInt64(2), testCumulative, testAttrs0...),
					),
				),
			),
			encoded: otlptest.ResourceMetrics(
				expectResource0,
				testSchema,
				otlptest.ScopeMetrics(
					expectScope0,
					otlptest.Sum(
						testName,
						testDesc,
						testUnit,
						expectCumulative,
						true, // monotonic
						otlptest.Int64DataPoint(expectAttrs0, startTime, endTime, 2),
					),
				),
			),
		},
		// float64 updowncounter, resource1, scope1, attrs1, delta
		{
			input: test.Metrics(
				testResource1,
				test.Scope(
					testScope1,
					test.Instrument(
						testFloat64(),
						test.Point(middleTime, endTime, sum.NewNonMonotonicFloat64(0.5), testDelta, testAttrs1...),
					),
				),
			),
			encoded: otlptest.ResourceMetrics(
				expectResource1,
				noSchema,
				otlptest.ScopeMetrics(
					expectScope1,
					otlptest.Sum(
						testName,
						testDesc,
						testUnit,
						expectDelta,
						false, // monotonic
						otlptest.Float64DataPoint(expectAttrs1, middleTime, endTime, 0.5),
					),
				),
			),
		},
		// int64 gauge, resource2, scope0, attrs0
		{
			input: test.Metrics(
				testResource2,
				test.Scope(
					testScope0,
					test.Instrument(
						testInt64(),
						test.Point(startTime, endTime, gauge.NewInt64(5), testDontCare, testAttrs0...),
					),
				),
			),
			encoded: otlptest.ResourceMetrics(
				expectResource2,
				testSchema,
				otlptest.ScopeMetrics(
					expectScope0,
					otlptest.Gauge(
						testName,
						testDesc,
						testUnit,
						otlptest.Int64DataPoint(expectAttrs0, noTime, endTime, 5),
					),
				),
			),
		},
		// histogram int64, resource1, scope1, attrs1, cumulative, only positive
		{
			input: test.Metrics(
				testResource1,
				test.Scope(
					testScope1,
					test.Instrument(
						testInt64(),
						test.Point(
							startTime, endTime,
							histogram.NewInt64(
								histogram.NewConfig(histogram.WithMaxSize(3)),
								1, 2, 4,
							),
							testCumulative, testAttrs1...,
						),
					),
				),
			),
			encoded: otlptest.ResourceMetrics(
				expectResource1,
				noSchema,
				otlptest.ScopeMetrics(
					expectScope1,
					otlptest.Histogram(
						testName,
						testDesc,
						testUnit,
						expectCumulative,
						otlptest.HistogramDataPoint(
							expectAttrs1, startTime, endTime, 7, 3, 0, 1, 4, 0, 0, []uint64{1, 1, 1}, 0, nil,
						),
					),
				),
			),
		},
		// histogram float64, resource0, scope0, attrs0, delta, positive and negative
		{
			input: test.Metrics(
				testResource0,
				test.Scope(
					testScope0,
					test.Instrument(
						testFloat64(),
						test.Point(
							middleTime, endTime,
							histogram.NewFloat64(
								histogram.NewConfig(histogram.WithMaxSize(3)),
								2, 4, 8, 0, 0, -0.5, -1, -2,
							),
							testDelta, testAttrs0...,
						),
					),
				),
			),
			encoded: otlptest.ResourceMetrics(
				expectResource0,
				testSchema,
				otlptest.ScopeMetrics(
					expectScope0,
					otlptest.Histogram(
						testName,
						testDesc,
						testUnit,
						expectDelta,
						otlptest.HistogramDataPoint(
							expectAttrs0,
							middleTime,
							endTime,
							10.5,
							8,
							2,
							-2, // min
							8,  // max
							0,
							// positive offset by 1
							1,
							[]uint64{1, 1, 1},
							// negative offset by -1
							-1,
							[]uint64{1, 1, 1},
						),
					),
				),
			),
		},
		// no points
		{
			input: test.Metrics(
				testResource0,
				test.Scope(
					testScope0,
					test.Instrument(
						testFloat64(),
					),
				),
			),
			encoded: otlptest.ResourceMetrics(
				expectResource0,
				testSchema,
				// Note: there is no attempt to avoid a scope w/ 0 instruments
				otlptest.ScopeMetrics(
					expectScope0,
				),
			),
		},
		// minmaxsumcount w/ ints, cumulative so no min/max
		{
			input: test.Metrics(
				testResource1,
				test.Scope(
					testScope0,
					test.Instrument(
						testInt64(),
						test.Point(
							startTime,
							endTime,
							minmaxsumcount.NewInt64(3, 2, 4, 1, 5),
							testCumulative,
							testAttrs1...,
						),
					),
				),
			),
			encoded: otlptest.ResourceMetrics(
				expectResource1,
				noSchema,
				otlptest.ScopeMetrics(
					expectScope0,
					otlptest.MinMaxSumCount(
						testName,
						testDesc,
						testUnit,
						expectCumulative,
						otlptest.MinMaxSumCountDataPoint(
							expectAttrs1, startTime, endTime,
							15, 5, math.NaN(), math.NaN(),
						),
					),
				),
			),
		},
		// minmaxsumcount w/ ints
		{
			input: test.Metrics(
				testResource1,
				test.Scope(
					testScope0,
					test.Instrument(
						testInt64(),
						test.Point(
							startTime,
							endTime,
							minmaxsumcount.NewInt64(3, 2, 4, 1, 5),
							testDelta,
							testAttrs1...,
						),
					),
				),
			),
			encoded: otlptest.ResourceMetrics(
				expectResource1,
				noSchema,
				otlptest.ScopeMetrics(
					expectScope0,
					otlptest.MinMaxSumCount(
						testName,
						testDesc,
						testUnit,
						expectDelta,
						otlptest.MinMaxSumCountDataPoint(
							expectAttrs1, startTime, endTime,
							15, 5, 1, 5,
						),
					),
				),
			),
		},
		// minmaxsumcount empty with no min/max
		{
			input: test.Metrics(
				testResource1,
				test.Scope(
					testScope0,
					test.Instrument(
						testInt64(),
						test.Point(
							startTime,
							endTime,
							minmaxsumcount.NewInt64(),
							testDelta,
							testAttrs1...,
						),
					),
				),
			),
			encoded: otlptest.ResourceMetrics(
				expectResource1,
				noSchema,
				otlptest.ScopeMetrics(
					expectScope0,
					otlptest.MinMaxSumCount(
						testName,
						testDesc,
						testUnit,
						expectDelta,
						otlptest.MinMaxSumCountDataPoint(
							expectAttrs1, startTime, endTime,
							0, 0, math.NaN(), math.NaN(),
						),
					),
				),
			),
		},
	} {
		asproto, err := Metrics(test.input)
		require.NoError(t, err)

		require.Equal(t, "", cmp.Diff(asproto, test.encoded, protocmp.Transform()))
	}
}
