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

package export

import (
	"context"
	"errors"

	"github.com/lightstep/go-expohisto/mapping/logarithm"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/internal"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/gauge"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/minmaxsumcount"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/sum"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	metricapi "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func toTemporality(t aggregation.Temporality) pmetric.AggregationTemporality {
	switch t {
	case aggregation.CumulativeTemporality:
		return pmetric.AggregationTemporalityCumulative
	case aggregation.DeltaTemporality:
		return pmetric.AggregationTemporalityDelta
	}
	return pmetric.AggregationTemporalityUnspecified
}

func copySumPoints(m pmetric.Metric, inM data.Instrument, mono bool) {
	s := m.SetEmptySum()
	s.SetIsMonotonic(mono)
	s.SetAggregationTemporality(toTemporality(inM.Points[0].Temporality))

	for _, inP := range inM.Points {
		dp := s.DataPoints().AppendEmpty()

		dp.SetStartTimestamp(pcommon.NewTimestampFromTime(inP.Start))
		dp.SetTimestamp(pcommon.NewTimestampFromTime(inP.End))

		internal.CopyAttributes(dp.Attributes(), inP.Attributes)

		switch t := inP.Aggregation.(type) {
		case *sum.MonotonicInt64:
			dp.SetIntValue(number.ToInt64(t.Sum()))
		case *sum.NonMonotonicInt64:
			dp.SetIntValue(number.ToInt64(t.Sum()))
		case *sum.MonotonicFloat64:
			dp.SetDoubleValue(number.ToFloat64(t.Sum()))
		case *sum.NonMonotonicFloat64:
			dp.SetDoubleValue(number.ToFloat64(t.Sum()))
		default:
			panic("unhandled case")
		}
	}
}

func copyGaugePoints(m pmetric.Metric, inM data.Instrument) {
	s := m.SetEmptyGauge()

	for _, inP := range inM.Points {
		dp := s.DataPoints().AppendEmpty()

		// Note: no start timestamp
		dp.SetTimestamp(pcommon.NewTimestampFromTime(inP.End))

		internal.CopyAttributes(dp.Attributes(), inP.Attributes)

		switch t := inP.Aggregation.(type) {
		case *gauge.Int64:
			dp.SetIntValue(number.ToInt64(t.Gauge()))
		case *gauge.Float64:
			dp.SetDoubleValue(number.ToFloat64(t.Gauge()))
		default:
			panic("unhandled case")
		}
	}
}

func copyExplicitHistogramPoints(m pmetric.Metric, inM data.Instrument) {
	s := m.SetEmptyHistogram()
	s.SetAggregationTemporality(toTemporality(inM.Points[0].Temporality))

	for _, inP := range inM.Points {
		dp := s.DataPoints().AppendEmpty()

		dp.SetStartTimestamp(pcommon.NewTimestampFromTime(inP.Start))
		dp.SetTimestamp(pcommon.NewTimestampFromTime(inP.End))
		internal.CopyAttributes(dp.Attributes(), inP.Attributes)

		switch t := inP.Aggregation.(type) {
		case *histogram.Int64:
			dp.SetSum(t.Sum().CoerceToFloat64(number.Int64Kind))
			dp.SetCount(t.Count())

			if t.Count() != 0 {
				dp.SetMax(t.Max().CoerceToFloat64(number.Int64Kind))
				dp.SetMin(t.Min().CoerceToFloat64(number.Int64Kind))
			}

			copyExplicitHistogramBuckets(t.Positive(), t.ZeroCount(), t.Scale(), dp)

		case *histogram.Float64:
			dp.SetSum(number.ToFloat64(t.Sum()))
			dp.SetCount(t.Count())
			if t.Count() != 0 {
				dp.SetMax(number.ToFloat64(t.Max()))
				dp.SetMin(number.ToFloat64(t.Min()))
			}

			copyExplicitHistogramBuckets(t.Positive(), t.ZeroCount(), t.Scale(), dp)
		default:
			panic("unhandled case")
		}
	}
}

// copyExplicitHistogramBuckets converts a Lightstep exponential histogram to an OTel histogram with
// explicitly defined buckets.
func copyExplicitHistogramBuckets(
	sourcePositiveBuckets aggregation.Buckets,
	sourceZeroCount uint64,
	sourceScale int32,
	dest pmetric.HistogramDataPoint,
) {
	// add the zero bucket in: (-Inf, 0]
	dest.ExplicitBounds().Append(0)
	dest.BucketCounts().Append(sourceZeroCount)

	if sourcePositiveBuckets.Len() > 0 {
		positiveOffset := sourcePositiveBuckets.Offset()
		positiveNumElements := int32(sourcePositiveBuckets.Len())
		mapping, _ := logarithm.NewMapping(sourceScale)
		leftBound, _ := mapping.LowerBoundary(positiveOffset)

		for element := int32(0); element < positiveNumElements; element++ {
			index := element + positiveOffset

			rightBound, _ := mapping.LowerBoundary(index + 1)

			// We have to add a bucket to get from zero to the start of the first user-defined bucket.
			// This has no count.
			if element == 0 {
				dest.ExplicitBounds().Append(leftBound)
				dest.BucketCounts().Append(0)
			}

			dest.ExplicitBounds().Append(rightBound)
			dest.BucketCounts().Append(sourcePositiveBuckets.At(uint32(element)))

			leftBound = rightBound
		}
	}

	// There are no elements in the (..., +Inf] bucket.
	dest.BucketCounts().Append(0)
}

func copyExponentialHistogramPoints(m pmetric.Metric, inM data.Instrument) {
	s := m.SetEmptyExponentialHistogram()
	s.SetAggregationTemporality(toTemporality(inM.Points[0].Temporality))

	for _, inP := range inM.Points {
		dp := s.DataPoints().AppendEmpty()

		dp.SetStartTimestamp(pcommon.NewTimestampFromTime(inP.Start))
		dp.SetTimestamp(pcommon.NewTimestampFromTime(inP.End))

		internal.CopyAttributes(dp.Attributes(), inP.Attributes)

		switch t := inP.Aggregation.(type) {
		case *histogram.Int64:
			dp.SetSum(t.Sum().CoerceToFloat64(number.Int64Kind))
			dp.SetCount(t.Count())
			dp.SetZeroCount(t.ZeroCount())
			dp.SetScale(t.Scale())
			if t.Count() != 0 {
				dp.SetMax(t.Max().CoerceToFloat64(number.Int64Kind))
				dp.SetMin(t.Min().CoerceToFloat64(number.Int64Kind))
			}
			if t.Positive().Len() != 0 {
				copyExponentialHistogramBuckets(dp.Positive(), t.Positive())
			}
			if t.Negative().Len() != 0 {
				copyExponentialHistogramBuckets(dp.Negative(), t.Negative())
			}
		case *histogram.Float64:
			dp.SetSum(number.ToFloat64(t.Sum()))
			dp.SetCount(t.Count())
			dp.SetZeroCount(t.ZeroCount())
			dp.SetScale(t.Scale())
			if t.Count() != 0 {
				dp.SetMax(number.ToFloat64(t.Max()))
				dp.SetMin(number.ToFloat64(t.Min()))
			}
			if t.Positive().Len() != 0 {
				copyExponentialHistogramBuckets(dp.Positive(), t.Positive())
			}
			if t.Negative().Len() != 0 {
				copyExponentialHistogramBuckets(dp.Negative(), t.Negative())
			}
		default:
			panic("unhandled case")
		}
	}
}

func copyExponentialHistogramBuckets(dest pmetric.ExponentialHistogramDataPointBuckets, src aggregation.Buckets) {
	if src.Len() == 0 {
		return
	}
	dest.SetOffset(src.Offset())
	dest.BucketCounts().EnsureCapacity(int(src.Len()))
	for i := uint32(0); i < src.Len(); i++ {
		dest.BucketCounts().Append(src.At(i))
	}
}

func copyMMSCPoints(m pmetric.Metric, inM data.Instrument) {
	s := m.SetEmptyHistogram()
	s.SetAggregationTemporality(toTemporality(inM.Points[0].Temporality))

	for _, inP := range inM.Points {
		dp := s.DataPoints().AppendEmpty()

		dp.SetStartTimestamp(pcommon.NewTimestampFromTime(inP.Start))
		dp.SetTimestamp(pcommon.NewTimestampFromTime(inP.End))

		internal.CopyAttributes(dp.Attributes(), inP.Attributes)

		switch t := inP.Aggregation.(type) {
		case *minmaxsumcount.Int64:
			dp.SetSum(t.Sum().CoerceToFloat64(number.Int64Kind))
			dp.SetCount(t.Count())
			if t.Count() != 0 {
				dp.SetMax(number.ToFloat64(t.Max()))
				dp.SetMin(number.ToFloat64(t.Min()))
			}
		case *minmaxsumcount.Float64:
			dp.SetSum(number.ToFloat64(t.Sum()))
			dp.SetCount(t.Count())
			if t.Count() != 0 {
				dp.SetMax(number.ToFloat64(t.Max()))
				dp.SetMin(number.ToFloat64(t.Min()))
			}
		default:
			panic("unhandled case")
		}
	}
}

func d2pd(
	resourceMap *internal.ResourceMap,
	in data.Metrics,
	useExponentialHistogram bool,
) pmetric.Metrics {
	out := pmetric.NewMetrics()
	rm := out.ResourceMetrics().AppendEmpty()

	resourceMap.Get(in.Resource).CopyTo(rm.Resource())

	for _, inS := range in.Scopes {
		sm := rm.ScopeMetrics().AppendEmpty()
		sm.Scope().SetName(inS.Library.Name)
		sm.Scope().SetVersion(inS.Library.Version)
		sm.SetSchemaUrl(inS.Library.SchemaURL)

		for _, inM := range inS.Instruments {
			if len(inM.Points) == 0 {
				continue
			}

			m := sm.Metrics().AppendEmpty()
			m.SetName(inM.Descriptor.Name)
			m.SetUnit(string(inM.Descriptor.Unit))
			m.SetDescription(inM.Descriptor.Description)

			switch inM.Points[0].Aggregation.(type) {
			case *sum.MonotonicInt64, *sum.MonotonicFloat64:
				copySumPoints(m, inM, true)
			case *sum.NonMonotonicInt64, *sum.NonMonotonicFloat64:
				copySumPoints(m, inM, false)
			case *gauge.Int64, *gauge.Float64:
				copyGaugePoints(m, inM)
			case *histogram.Int64, *histogram.Float64:
				if useExponentialHistogram {
					copyExponentialHistogramPoints(m, inM)
				} else {
					copyExplicitHistogramPoints(m, inM)
				}
			case *minmaxsumcount.Int64, *minmaxsumcount.Float64:
				copyMMSCPoints(m, inM)
			}
		}
	}

	return out
}

func ExportMetrics(
	ctx context.Context,
	data data.Metrics,
	tracer trace.Tracer,
	telemetryItemsCounter metricapi.Int64Counter,
	resourceMap *internal.ResourceMap,
	exporter exporter.Metrics,
	useExponentialHistogram bool,
) error {
	ctx, span := tracer.Start(
		ctx,
		"otelsdk_export_metrics",
	)
	defer span.End()

	converted := d2pd(resourceMap, data, useExponentialHistogram)
	points := int64(converted.DataPointCount())

	err := exporter.ConsumeMetrics(ctx, converted)
	success := err == nil
	var state string
	if success {
		state = "ok"
	} else if errors.Is(err, context.Canceled) {
		state = "canceled"
	} else if errors.Is(err, context.DeadlineExceeded) {
		state = "timeout"
	} else {
		state = "error"
	}

	var attrs = []attribute.KeyValue{
		attribute.Bool("success", success),
		attribute.String("state", state),
	}
	telemetryItemsCounter.Add(ctx, points, metricapi.WithAttributes(attrs...))
	span.SetAttributes(append(attrs, attribute.Int64("num_points", points))...)
	if err == nil {
		span.SetStatus(otelcodes.Ok, state)
	} else {
		span.SetStatus(otelcodes.Error, err.Error())
	}
	return err
}
