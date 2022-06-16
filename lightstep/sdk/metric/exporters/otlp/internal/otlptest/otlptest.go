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

package otlptest

import (
	"time"

	collectorpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

func ExportRequest(rms ...*metricspb.ResourceMetrics) *collectorpb.ExportMetricsServiceRequest {
	return &collectorpb.ExportMetricsServiceRequest{
		ResourceMetrics: rms,
	}
}

func ResourceMetrics(resource *resourcepb.Resource, schemaURL string, ilms ...*metricspb.ScopeMetrics) *metricspb.ResourceMetrics {
	return &metricspb.ResourceMetrics{
		Resource:     resource,
		SchemaUrl:    schemaURL,
		ScopeMetrics: ilms,
	}
}

func KeyValue(k, v string) *commonpb.KeyValue {
	return &commonpb.KeyValue{
		Key: k,
		Value: &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: v,
			},
		},
	}
}

func KeyIntValue(k string, v int) *commonpb.KeyValue {
	return &commonpb.KeyValue{
		Key: k,
		Value: &commonpb.AnyValue{
			Value: &commonpb.AnyValue_IntValue{
				IntValue: int64(v),
			},
		},
	}
}

func Resource(kvs ...*commonpb.KeyValue) *resourcepb.Resource {
	return &resourcepb.Resource{
		Attributes: kvs,
	}
}

func Attributes(kvs ...*commonpb.KeyValue) []*commonpb.KeyValue {
	return kvs

}

func Scope(name, version string) *commonpb.InstrumentationScope {
	return &commonpb.InstrumentationScope{
		Name:    name,
		Version: version,
	}
}

func ScopeMetrics(scope *commonpb.InstrumentationScope, ms ...*metricspb.Metric) *metricspb.ScopeMetrics {
	return &metricspb.ScopeMetrics{
		Scope:   scope,
		Metrics: ms,
	}
}

func Sum(name, desc, unit string, tempo metricspb.AggregationTemporality, monotone bool, idps ...*metricspb.NumberDataPoint) *metricspb.Metric {

	return &metricspb.Metric{
		Name:        name,
		Description: desc,
		Unit:        unit,
		Data: &metricspb.Metric_Sum{
			Sum: &metricspb.Sum{
				AggregationTemporality: tempo,
				IsMonotonic:            monotone,
				DataPoints:             idps,
			},
		},
	}
}

func toNanos(t time.Time) uint64 {
	if t.IsZero() {
		return 0
	}
	return uint64(t.UnixNano())
}

func Int64DataPoint(attributes []*commonpb.KeyValue, start, end time.Time, value int64) *metricspb.NumberDataPoint {
	return &metricspb.NumberDataPoint{
		Attributes:        attributes,
		StartTimeUnixNano: toNanos(start),
		TimeUnixNano:      toNanos(end),
		Value:             &metricspb.NumberDataPoint_AsInt{AsInt: value},
	}
}

func Float64DataPoint(attributes []*commonpb.KeyValue, start, end time.Time, value float64) *metricspb.NumberDataPoint {
	return &metricspb.NumberDataPoint{
		Attributes:        attributes,
		StartTimeUnixNano: toNanos(start),
		TimeUnixNano:      toNanos(end),
		Value:             &metricspb.NumberDataPoint_AsDouble{AsDouble: value},
	}
}

func Gauge(name, desc, unit string, idps ...*metricspb.NumberDataPoint) *metricspb.Metric {
	return &metricspb.Metric{
		Name:        name,
		Description: desc,
		Unit:        unit,
		Data: &metricspb.Metric_Gauge{
			Gauge: &metricspb.Gauge{
				DataPoints: idps,
			},
		},
	}
}

func HistogramDataPoint(attributes []*commonpb.KeyValue, start, end time.Time, sum float64, count, zeroCount uint64, scale, posOffset int32, posBucketCounts []uint64, negOffset int32, negBucketCounts []uint64) *metricspb.ExponentialHistogramDataPoint {
	dp := &metricspb.ExponentialHistogramDataPoint{
		Attributes:        attributes,
		StartTimeUnixNano: toNanos(start),
		TimeUnixNano:      toNanos(end),
		Sum:               &sum,
		Count:             count,
		ZeroCount:         zeroCount,
		Scale:             scale,
	}
	if posBucketCounts != nil {
		dp.Positive = &metricspb.ExponentialHistogramDataPoint_Buckets{
			Offset:       posOffset,
			BucketCounts: posBucketCounts,
		}
	}
	if negBucketCounts != nil {
		dp.Negative = &metricspb.ExponentialHistogramDataPoint_Buckets{
			Offset:       negOffset,
			BucketCounts: negBucketCounts,
		}
	}
	return dp
}

func Histogram(name, desc, unit string, tempo metricspb.AggregationTemporality, idps ...*metricspb.ExponentialHistogramDataPoint) *metricspb.Metric {
	return &metricspb.Metric{
		Name:        name,
		Description: desc,
		Unit:        unit,
		Data: &metricspb.Metric_ExponentialHistogram{
			ExponentialHistogram: &metricspb.ExponentialHistogram{
				AggregationTemporality: tempo,
				DataPoints:             idps,
			},
		},
	}
}
