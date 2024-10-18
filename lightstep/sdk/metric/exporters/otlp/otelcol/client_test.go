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

package otelcol

import (
	"context"
	"math"
	"math/rand"
	"net"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	sdkmetric "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/export"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/otelarrowreceiver"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

// Note: unclear which test support library we should use until this
// moves into an otel repo.  Some simple test supports are developed
// here anyway to defer this question.

var (
	testResourceAttrs = []attribute.KeyValue{
		attribute.String("service.name", "tester"),
		attribute.String("property", "value"),
	}

	testStmtAttrs = []attribute.KeyValue{
		filteredAttr,
		unfilteredAttr,
	}

	filteredAttr   = attribute.String("A", "1")
	unfilteredAttr = attribute.Int("B", 2)
)

const (
	exemplarTimestamp pcommon.Timestamp = 12345
)

func attrs2otlp(kvs ...attribute.KeyValue) []*commonpb.KeyValue {
	var r []*commonpb.KeyValue
	for _, kv := range kvs {
		switch kv.Value.Type() {
		case attribute.BOOL:
			r = append(r, &commonpb.KeyValue{
				Key: string(kv.Key),
				Value: &commonpb.AnyValue{
					Value: &commonpb.AnyValue_BoolValue{
						BoolValue: kv.Value.AsBool(),
					},
				},
			})
		case attribute.INT64:
			r = append(r, &commonpb.KeyValue{
				Key: string(kv.Key),
				Value: &commonpb.AnyValue{
					Value: &commonpb.AnyValue_IntValue{
						IntValue: kv.Value.AsInt64(),
					},
				},
			})
		case attribute.FLOAT64:
			r = append(r, &commonpb.KeyValue{
				Key: string(kv.Key),
				Value: &commonpb.AnyValue{
					Value: &commonpb.AnyValue_DoubleValue{
						DoubleValue: kv.Value.AsFloat64(),
					},
				},
			})
		case attribute.STRING:
			r = append(r, &commonpb.KeyValue{
				Key: string(kv.Key),
				Value: &commonpb.AnyValue{
					Value: &commonpb.AnyValue_StringValue{
						StringValue: kv.Value.AsString(),
					},
				},
			})
		default:
			// Note: Missing tests here.
			panic("untested cases")
		}
	}
	return r
}

type clientTestSuite struct {
	suite.Suite

	addr   string
	recv   receiver.Metrics
	sink   *consumertest.MetricsSink
	sdk    *sdkmetric.MeterProvider
	before time.Time
	after  time.Time
}

type timedPoint interface {
	StartTimestamp() pcommon.Timestamp
	Timestamp() pcommon.Timestamp
	SetStartTimestamp(pcommon.Timestamp)
	SetTimestamp(pcommon.Timestamp)
}

func TestExporterSuite(t *testing.T) {
	suite.Run(t, new(clientTestSuite))
}

func (t *clientTestSuite) SetupTest() {
	ctx := context.Background()

	t.sink.Reset()
	t.before = timeNow()

	exp, err := NewExporter(
		ctx,
		NewConfig(
			WithInsecure(),
			WithEndpoint(t.addr),
			WithHeaders(map[string]string{"lightstep-access-token": "${TOKEN}"}),
		),
	)
	t.NoError(err)

	t.sdk = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exp, math.MaxInt64),
			view.WithDefaultAggregationTemporalitySelector(aggregation.LowMemoryTemporality),
			view.WithClause(view.WithKeys([]attribute.Key{"B"})),
		),
		sdkmetric.WithResource(
			resource.NewSchemaless(testResourceAttrs...),
		),
	)
}

func (t *clientTestSuite) SetupSuite() {
	ctx := context.Background()

	listener, err := net.Listen("tcp", "127.0.0.1:")
	t.NoError(err)
	t.addr = listener.Addr().String()

	t.NoError(listener.Close())

	factory := otelarrowreceiver.NewFactory()
	cfg := factory.CreateDefaultConfig().(*otelarrowreceiver.Config)
	cfg.Protocols.Arrow = otelarrowreceiver.ArrowConfig{}
	cfg.GRPC.NetAddr = confignet.AddrConfig{Endpoint: t.addr, Transport: "tcp"}

	set := receivertest.NewNopSettings()
	tc := &consumertest.MetricsSink{}

	mr, err := factory.CreateMetrics(ctx, set, cfg, tc)
	t.NoError(err)

	err = mr.Start(ctx, componenttest.NewNopHost())
	t.NoError(err)

	t.recv = mr
	t.sink = tc
}

func (t *clientTestSuite) checkTimedPoint(p timedPoint) {
	if p.StartTimestamp() != 0 {
		t.LessOrEqual(t.before, p.StartTimestamp().AsTime())
	}

	t.LessOrEqual(t.before, t.after)
	t.LessOrEqual(p.Timestamp().AsTime(), t.after)

	t.GreaterOrEqual(t.after, p.Timestamp().AsTime())
	t.LessOrEqual(t.before, p.Timestamp().AsTime())

	// Set these fields to zero, making them omitted fields in JSON,
	// which allows cmp.Diff() uses following this call to work.
	p.SetStartTimestamp(0)
	p.SetTimestamp(0)
}

func (t *clientTestSuite) assertTimestamps() {
	t.after = timeNow()
	for _, export := range t.sink.AllMetrics() {
		for ri := 0; ri < export.ResourceMetrics().Len(); ri++ {
			rm := export.ResourceMetrics().At(ri)
			for si := 0; si < rm.ScopeMetrics().Len(); si++ {
				sm := rm.ScopeMetrics().At(si)
				for mi := 0; mi < sm.Metrics().Len(); mi++ {
					m := sm.Metrics().At(mi)

					switch m.Type() {
					case pmetric.MetricTypeGauge:
						for pi := 0; pi < m.Gauge().DataPoints().Len(); pi++ {
							t.checkTimedPoint(m.Gauge().DataPoints().At(pi))
						}
					case pmetric.MetricTypeSum:
						for pi := 0; pi < m.Sum().DataPoints().Len(); pi++ {
							t.checkTimedPoint(m.Sum().DataPoints().At(pi))
						}
					case pmetric.MetricTypeHistogram:
						for pi := 0; pi < m.Histogram().DataPoints().Len(); pi++ {
							t.checkTimedPoint(m.Histogram().DataPoints().At(pi))
						}
					case pmetric.MetricTypeExponentialHistogram:
						for pi := 0; pi < m.ExponentialHistogram().DataPoints().Len(); pi++ {
							t.checkTimedPoint(m.ExponentialHistogram().DataPoints().At(pi))
						}
					default:
						// e.g., summary point
						t.Fail("data type not used")
					}
				}
			}
		}
	}
}

func timeNow() time.Time {
	return time.Now()
}

func newFloat64(x float64) *float64 {
	return &x
}

func (t *clientTestSuite) TestCounterAndGauge() {
	ctx := context.Background()

	meter := t.sdk.Meter("test-meter")

	counter, _ := meter.Int64Counter("how-many")

	meter.Int64ObservableGauge("pressure",
		metric.WithInt64Callback(func(_ context.Context, obs metric.Int64Observer) error {
			obs.Observe(2, metric.WithAttributes(testStmtAttrs...))
			return nil
		}))

	counter.Add(ctx, 1, metric.WithAttributes(testStmtAttrs...))

	t.NoError(t.sdk.Shutdown(ctx))

	t.Equal(1, len(t.sink.AllMetrics()))

	t.assertTimestamps()

	data, err := pmetricotlp.NewExportRequestFromMetrics(t.sink.AllMetrics()[0]).MarshalProto()
	t.NoError(err)

	expect := colmetricspb.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricspb.ResourceMetrics{
			{
				Resource: &resourcepb.Resource{
					Attributes: attrs2otlp(testResourceAttrs...),
				},
				ScopeMetrics: []*metricspb.ScopeMetrics{
					{
						Scope: &commonpb.InstrumentationScope{
							Name: "test-meter",
						},
						Metrics: []*metricspb.Metric{
							{
								Name: "how-many",
								Data: &metricspb.Metric_Sum{
									Sum: &metricspb.Sum{
										IsMonotonic:            true,
										AggregationTemporality: metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA,
										DataPoints: []*metricspb.NumberDataPoint{
											{
												Attributes: attrs2otlp(unfilteredAttr),
												Value: &metricspb.NumberDataPoint_AsInt{
													AsInt: 1,
												},
											},
										},
									},
								},
							},
							{
								Name: "pressure",
								Data: &metricspb.Metric_Gauge{
									Gauge: &metricspb.Gauge{
										DataPoints: []*metricspb.NumberDataPoint{
											{
												Attributes: attrs2otlp(unfilteredAttr),
												Value: &metricspb.NumberDataPoint_AsInt{
													AsInt: 2,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	var export colmetricspb.ExportMetricsServiceRequest
	t.NoError(proto.Unmarshal(data, &export))

	t.Empty(cmp.Diff(prototext.Format(&expect), prototext.Format(&export)))
}

func (t *clientTestSuite) TestUpDownCounters() {
	ctx := context.Background()

	meter := t.sdk.Meter("test-meter")

	counter, _ := meter.Float64UpDownCounter("in-flight",
		metric.WithDescription(`{ "temporality": "delta" }`),
	)

	meter.Float64ObservableUpDownCounter("in-use",
		metric.WithFloat64Callback(func(_ context.Context, obs metric.Float64Observer) error {
			obs.Observe(2, metric.WithAttributes(testStmtAttrs...))
			return nil
		}))

	counter.Add(ctx, 1, metric.WithAttributes(testStmtAttrs...))

	_ = t.sdk.ForceFlush(ctx)

	counter.Add(ctx, 1, metric.WithAttributes(testStmtAttrs...))

	_ = t.sdk.Shutdown(ctx)

	t.Equal(2, len(t.sink.AllMetrics()))

	expect := colmetricspb.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricspb.ResourceMetrics{
			{
				Resource: &resourcepb.Resource{
					Attributes: attrs2otlp(testResourceAttrs...),
				},
				ScopeMetrics: []*metricspb.ScopeMetrics{
					{
						Scope: &commonpb.InstrumentationScope{
							Name: "test-meter",
						},
						Metrics: []*metricspb.Metric{
							{
								Name: "in-flight",
								Data: &metricspb.Metric_Sum{
									Sum: &metricspb.Sum{
										AggregationTemporality: metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA,
										DataPoints: []*metricspb.NumberDataPoint{
											{
												Attributes: attrs2otlp(unfilteredAttr),
												Value: &metricspb.NumberDataPoint_AsDouble{
													AsDouble: 1,
												},
											},
										},
									},
								},
							},
							{
								Name: "in-use",
								Data: &metricspb.Metric_Sum{
									Sum: &metricspb.Sum{
										AggregationTemporality: metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
										DataPoints: []*metricspb.NumberDataPoint{
											{
												Attributes: attrs2otlp(unfilteredAttr),
												Value: &metricspb.NumberDataPoint_AsDouble{
													AsDouble: 2,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	t.assertTimestamps()

	for _, m := range t.sink.AllMetrics() {
		var export colmetricspb.ExportMetricsServiceRequest
		data, err := pmetricotlp.NewExportRequestFromMetrics(m).MarshalProto()
		t.NoError(err)
		t.NoError(proto.Unmarshal(data, &export))

		t.Empty(cmp.Diff(prototext.Format(&expect), prototext.Format(&export)))
	}
}

func (t *clientTestSuite) TestHistograms() {
	ctx := context.Background()

	meter := t.sdk.Meter("test-meter")

	histo1, _ := meter.Int64Histogram("duration",
		metric.WithUnit("ms"),
		metric.WithDescription(`{ "config": { "histogram": { "max_size": 4 } } }`))
	histo2, _ := meter.Float64Histogram("latency",
		metric.WithDescription(`{ "aggregation": "minmaxsumcount", "temporality": "cumulative" }`),
	)

	histo1.Record(ctx, 1, metric.WithAttributes(testStmtAttrs...))
	histo1.Record(ctx, 2, metric.WithAttributes(testStmtAttrs...))
	histo1.Record(ctx, 4, metric.WithAttributes(testStmtAttrs...))

	histo2.Record(ctx, 1, metric.WithAttributes(testStmtAttrs...))
	histo2.Record(ctx, 2, metric.WithAttributes(testStmtAttrs...))
	histo2.Record(ctx, 4, metric.WithAttributes(testStmtAttrs...))

	_ = t.sdk.Shutdown(ctx)

	expect := colmetricspb.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricspb.ResourceMetrics{
			{
				Resource: &resourcepb.Resource{
					Attributes: attrs2otlp(testResourceAttrs...),
				},
				ScopeMetrics: []*metricspb.ScopeMetrics{
					{
						Scope: &commonpb.InstrumentationScope{
							Name: "test-meter",
						},
						Metrics: []*metricspb.Metric{
							{
								Name: "duration",
								Unit: "ms",
								Data: &metricspb.Metric_ExponentialHistogram{
									ExponentialHistogram: &metricspb.ExponentialHistogram{
										AggregationTemporality: metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA,
										DataPoints: []*metricspb.ExponentialHistogramDataPoint{
											{
												Attributes: attrs2otlp(unfilteredAttr),
												Count:      3,
												Min:        newFloat64(1),
												Max:        newFloat64(4),
												Scale:      0,
												Sum:        newFloat64(7),
												Positive: &metricspb.ExponentialHistogramDataPoint_Buckets{
													Offset:       -1,
													BucketCounts: []uint64{1, 1, 1},
												},
												// TODO: Not sure why prototext shows this, workaround? (or gogo bug?).
												Negative: &metricspb.ExponentialHistogramDataPoint_Buckets{},
											},
										},
									},
								},
							},
							{
								Name: "latency",
								Data: &metricspb.Metric_Histogram{
									Histogram: &metricspb.Histogram{
										AggregationTemporality: metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE,
										DataPoints: []*metricspb.HistogramDataPoint{
											{
												Attributes: attrs2otlp(unfilteredAttr),
												Count:      3,
												Min:        newFloat64(1),
												Max:        newFloat64(4),
												Sum:        newFloat64(7),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	t.Equal(1, len(t.sink.AllMetrics()))

	t.assertTimestamps()

	var export colmetricspb.ExportMetricsServiceRequest
	data, err := pmetricotlp.NewExportRequestFromMetrics(t.sink.AllMetrics()[0]).MarshalProto()
	t.NoError(err)
	t.NoError(proto.Unmarshal(data, &export))

	t.Empty(cmp.Diff(prototext.Format(&expect), prototext.Format(&export)))
}

func (t *clientTestSuite) TestSumExemplars() {
	// Note: only sum exemplars are tested.  TODO: the other data
	// points' code is nearly identical, but could be more tested.
	ctx := context.Background()

	meter := t.sdk.Meter("test-meter")

	counter, err := meter.Int64Counter("how-many",
		metric.WithDescription(`{
	  "config": {
	    "exemplar": {
	      "filter": "always_on",
	      "size": 1
	    }
	  },
	  "description": "incredible"
	}`),
	)
	t.NoError(err)

	before := time.Now()
	counter.Add(ctx, 17, metric.WithAttributes(testStmtAttrs...))
	after := time.Now()

	_ = t.sdk.Shutdown(ctx)

	t.Equal(1, len(t.sink.AllMetrics()))

	t.assertTimestamps()

	data, err := pmetricotlp.NewExportRequestFromMetrics(t.sink.AllMetrics()[0]).MarshalProto()
	t.NoError(err)

	exemplars := []*metricspb.Exemplar{
		{
			TimeUnixNano:       uint64(exemplarTimestamp),
			Value:              &metricspb.Exemplar_AsInt{AsInt: 17},
			FilteredAttributes: attrs2otlp(filteredAttr),
		},
	}

	expect := colmetricspb.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricspb.ResourceMetrics{
			{
				Resource: &resourcepb.Resource{
					Attributes: attrs2otlp(testResourceAttrs...),
				},
				ScopeMetrics: []*metricspb.ScopeMetrics{
					{
						Scope: &commonpb.InstrumentationScope{
							Name: "test-meter",
						},
						Metrics: []*metricspb.Metric{
							{
								Name:        "how-many",
								Description: "incredible",
								Data: &metricspb.Metric_Sum{
									Sum: &metricspb.Sum{
										IsMonotonic:            true,
										AggregationTemporality: metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA,
										DataPoints: []*metricspb.NumberDataPoint{
											{
												Attributes: attrs2otlp(unfilteredAttr),
												Value: &metricspb.NumberDataPoint_AsInt{
													AsInt: 17,
												},
												Exemplars: exemplars,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	var export colmetricspb.ExportMetricsServiceRequest
	t.NoError(proto.Unmarshal(data, &export))

	// The following is similar to assertTimestamps, but for the
	// exemplar timestamp.  The timestamp should be in-range.
	// Reset it to exemplarTimestamp for the cmp.Diff to succeed.
	ts := pcommon.Timestamp(export.ResourceMetrics[0].ScopeMetrics[0].Metrics[0].Data.(*metricspb.Metric_Sum).Sum.DataPoints[0].Exemplars[0].TimeUnixNano)
	t.True(!before.After(ts.AsTime()))
	t.True(!after.Before(ts.AsTime()))
	export.ResourceMetrics[0].ScopeMetrics[0].Metrics[0].Data.(*metricspb.Metric_Sum).Sum.DataPoints[0].Exemplars[0].TimeUnixNano = uint64(exemplarTimestamp)

	t.Empty(cmp.Diff(prototext.Format(&expect), prototext.Format(&export)))
}

func (t *clientTestSuite) TestFilteredAttributes() {
	allAttrs := []attribute.KeyValue{attribute.String("A", "1"),
		attribute.String("B", "2"),
		attribute.String("C", "3"),
	}

	// exhaustively
	for nk := number.Int64Kind; nk <= number.Float64Kind; nk++ {
		for perm := 0; perm < 1<<len(allAttrs); perm++ {
			for shuf := 0; shuf < 8; shuf++ {
				rand.Shuffle(len(allAttrs), func(i, j int) {
					allAttrs[i], allAttrs[j] = allAttrs[j], allAttrs[i]
				})

				var useAttrs []attribute.KeyValue
				var filtAttrs []attribute.KeyValue

				for bit := 0; bit < len(allAttrs); bit++ {
					if (1<<bit)&perm != 0 {
						useAttrs = append(useAttrs, allAttrs[bit])
					} else {
						filtAttrs = append(filtAttrs, allAttrs[bit])
					}
				}
				var num number.Number
				if nk == number.Int64Kind {
					num = number.FromInt64(1000)
				} else {
					num = number.FromFloat64(1000)
				}

				attrs := attribute.NewSet(useAttrs...)

				dest := pmetric.NewExemplarSlice()
				export.CopyExemplars(
					dest, attrs, nk,
					[]aggregator.WeightedExemplarBits{
						{
							ExemplarBits: aggregator.ExemplarBits{
								Time:       exemplarTimestamp.AsTime(),
								Attributes: allAttrs,
								Number:     num,
							},
							Weight: 0, // w/ reservoir size == 1, weight == 0
						},
					})

				expect := pmetric.NewExemplarSlice()
				exex := expect.AppendEmpty()
				exex.SetTimestamp(exemplarTimestamp)
				if nk == number.Int64Kind {
					exex.SetIntValue(1000)
				} else {
					exex.SetDoubleValue(1000)
				}
				// Sort the expected filtered attributes so they match
				sort.Slice(filtAttrs, func(i, j int) bool {
					return filtAttrs[i].Key < filtAttrs[j].Key
				})

				for _, oattr := range attrs2otlp(filtAttrs...) {
					exex.FilteredAttributes().PutStr(oattr.Key, oattr.Value.Value.(*commonpb.AnyValue_StringValue).StringValue)
				}
				t.Equal(expect, dest)
			}
		}
	}
}
