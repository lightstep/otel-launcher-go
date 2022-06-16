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

// Package metrictransform provides translations for opentelemetry-go concepts and
// structures to otlp structures.
package metrictransform // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp/internal/metrictransform"

import (
	"errors"
	"time"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

var (
	// ErrUnimplementedAgg is returned when a transformation of an unimplemented
	// aggregator is attempted.
	ErrUnimplementedAgg = errors.New("unimplemented aggregator")
)

// result is the product of transforming Records into OTLP Metrics.
type result struct {
	Metric *metricspb.Metric
	Err    error
}

// toNanos returns the number of nanoseconds since the UNIX epoch.
func toNanos(t time.Time) uint64 {
	if t.IsZero() {
		return 0
	}
	return uint64(t.UnixNano())
}

// Metrics transforms one batch of metrics into an OTLP ResourceMetrics.
func Metrics(metrics data.Metrics) (*metricspb.ResourceMetrics, error) {
	rm := &metricspb.ResourceMetrics{
		Resource:     Resource(metrics.Resource),
		SchemaUrl:    metrics.Resource.SchemaURL(),
		ScopeMetrics: make([]*metricspb.ScopeMetrics, 0, len(metrics.Scopes)),
	}
	for _, scope := range metrics.Scopes {
		sc := &metricspb.ScopeMetrics{
			Scope:     Library(scope.Library),
			Metrics:   make([]*metricspb.Metric, 0, len(scope.Instruments)),
			SchemaUrl: scope.Library.SchemaURL,
		}
		rm.ScopeMetrics = append(rm.ScopeMetrics, sc)

		for _, inst := range scope.Instruments {
			if len(inst.Points) == 0 {
				continue
			}
			mm := &metricspb.Metric{
				Name:        inst.Descriptor.Name,
				Unit:        string(inst.Descriptor.Unit),
				Description: inst.Descriptor.Description,
			}
			point0 := inst.Points[0]
			switch point0.Aggregation.Kind() {
			case aggregation.MonotonicSumKind:
				mm.Data = &metricspb.Metric_Sum{
					Sum: &metricspb.Sum{
						AggregationTemporality: Temporality(point0.Temporality),
						IsMonotonic:            true,
						DataPoints:             NumberPoints(&inst.Descriptor, inst.Points, sumToValue),
					},
				}
			case aggregation.NonMonotonicSumKind:
				mm.Data = &metricspb.Metric_Sum{
					Sum: &metricspb.Sum{
						AggregationTemporality: Temporality(point0.Temporality),
						IsMonotonic:            false,
						DataPoints:             NumberPoints(&inst.Descriptor, inst.Points, sumToValue),
					},
				}
			case aggregation.HistogramKind:
				mm.Data = &metricspb.Metric_ExponentialHistogram{
					ExponentialHistogram: &metricspb.ExponentialHistogram{
						AggregationTemporality: Temporality(point0.Temporality),
						DataPoints:             HistogramPoints(&inst.Descriptor, inst.Points),
					},
				}
			case aggregation.MinMaxSumCountKind:
				mm.Data = &metricspb.Metric_Histogram{
					Histogram: &metricspb.Histogram{
						AggregationTemporality: Temporality(point0.Temporality),
						DataPoints:             MinMaxSumCountPoints(&inst.Descriptor, inst.Points, point0.Temporality),
					},
				}
			case aggregation.GaugeKind:
				mm.Data = &metricspb.Metric_Gauge{
					Gauge: &metricspb.Gauge{
						DataPoints: NumberPoints(&inst.Descriptor, inst.Points, gaugeToValue),
					},
				}
			default:
				return nil, ErrUnimplementedAgg
			}
			sc.Metrics = append(sc.Metrics, mm)
		}
	}

	return rm, nil

}

func sumToValue(pt data.Point) number.Number {
	return pt.Aggregation.(aggregation.Sum).Sum()
}

func gaugeToValue(pt data.Point) number.Number {
	return pt.Aggregation.(aggregation.Gauge).Gauge()
}

func NumberPoints(desc *sdkinstrument.Descriptor, points []data.Point, p2v func(data.Point) number.Number) []*metricspb.NumberDataPoint {
	results := make([]*metricspb.NumberDataPoint, len(points))
	for i, pt := range points {
		results[i] = &metricspb.NumberDataPoint{
			Attributes:        Attributes(pt.Attributes),
			StartTimeUnixNano: toNanos(pt.Start),
			TimeUnixNano:      toNanos(pt.End),
		}
		value := p2v(pt)
		if desc.NumberKind == number.Float64Kind {
			results[i].Value = &metricspb.NumberDataPoint_AsDouble{
				AsDouble: number.ToFloat64(value),
			}
		} else {
			results[i].Value = &metricspb.NumberDataPoint_AsInt{
				AsInt: number.ToInt64(value),
			}
		}
	}
	return results
}

func HistogramPoints(desc *sdkinstrument.Descriptor, points []data.Point) []*metricspb.ExponentialHistogramDataPoint {
	results := make([]*metricspb.ExponentialHistogramDataPoint, len(points))
	for i, pt := range points {
		hist := pt.Aggregation.(aggregation.Histogram)
		// Note: We assume that inputs are non-negative by the
		// OTel API contract; If inputs are negative, we're
		// supposed to drop the sum.
		sum := hist.Sum().CoerceToFloat64(desc.NumberKind)

		results[i] = &metricspb.ExponentialHistogramDataPoint{
			Attributes:        Attributes(pt.Attributes),
			StartTimeUnixNano: toNanos(pt.Start),
			TimeUnixNano:      toNanos(pt.End),
			Count:             hist.Count(),
			Sum:               &sum,
			ZeroCount:         hist.ZeroCount(),
			Scale:             hist.Scale(),
			Positive:          HistogramBuckets(hist.Positive()),
			Negative:          HistogramBuckets(hist.Negative()),
		}
	}
	return results
}

func HistogramBuckets(b aggregation.Buckets) *metricspb.ExponentialHistogramDataPoint_Buckets {
	if b.Len() == 0 {
		return nil
	}
	result := &metricspb.ExponentialHistogramDataPoint_Buckets{
		Offset:       b.Offset(),
		BucketCounts: make([]uint64, b.Len()),
	}
	for i := range result.BucketCounts {
		result.BucketCounts[i] = b.At(uint32(i))
	}
	return result
}

func float64Ptr(x float64) *float64 {
	return &x
}

func MinMaxSumCountPoints(desc *sdkinstrument.Descriptor, points []data.Point, tempo aggregation.Temporality) []*metricspb.HistogramDataPoint {
	results := make([]*metricspb.HistogramDataPoint, len(points))
	for i, pt := range points {
		mmsc := pt.Aggregation.(aggregation.MinMaxSumCount)

		// See note about optional sum at top of minmaxsumcount.go
		sum := mmsc.Sum().CoerceToFloat64(desc.NumberKind)

		var min, max *float64

		if mmsc.Count() != 0 {
			min = float64Ptr(mmsc.Min().CoerceToFloat64(desc.NumberKind))
			max = float64Ptr(mmsc.Max().CoerceToFloat64(desc.NumberKind))
		}

		results[i] = &metricspb.HistogramDataPoint{
			Attributes:        Attributes(pt.Attributes),
			StartTimeUnixNano: toNanos(pt.Start),
			TimeUnixNano:      toNanos(pt.End),
			Count:             mmsc.Count(),
			Sum:               &sum,
			Min:               min,
			Max:               max,
		}
	}
	return results
}
