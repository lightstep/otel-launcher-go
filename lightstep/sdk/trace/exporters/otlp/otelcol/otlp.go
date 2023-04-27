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
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/gauge"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/minmaxsumcount"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/sum"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/otel/attribute"
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

func copyAttributes(dest pcommon.Map, src attribute.Set) {
	for iter := src.Iter(); iter.Next(); {
		inA := iter.Attribute()
		key := string(inA.Key)
		switch inA.Value.Type() {
		case attribute.BOOL:
			dest.PutBool(key, inA.Value.AsBool())
		case attribute.INT64:
			dest.PutInt(key, inA.Value.AsInt64())
		case attribute.FLOAT64:
			dest.PutDouble(key, inA.Value.AsFloat64())
		case attribute.STRING:
			dest.PutStr(key, inA.Value.AsString())
		case attribute.BOOLSLICE:
			sl := dest.PutEmptySlice(key)
			sl.EnsureCapacity(len(inA.Value.AsBoolSlice()))
			for _, v := range inA.Value.AsBoolSlice() {
				sl.AppendEmpty().SetBool(v)
			}
		case attribute.INT64SLICE:
			sl := dest.PutEmptySlice(key)
			sl.EnsureCapacity(len(inA.Value.AsInt64Slice()))
			for _, v := range inA.Value.AsInt64Slice() {
				sl.AppendEmpty().SetInt(v)
			}
		case attribute.FLOAT64SLICE:
			sl := dest.PutEmptySlice(key)
			sl.EnsureCapacity(len(inA.Value.AsFloat64Slice()))
			for _, v := range inA.Value.AsFloat64Slice() {
				sl.AppendEmpty().SetDouble(v)
			}
		case attribute.STRINGSLICE:
			sl := dest.PutEmptySlice(key)
			sl.EnsureCapacity(len(inA.Value.AsStringSlice()))
			for _, v := range inA.Value.AsStringSlice() {
				sl.AppendEmpty().SetStr(v)
			}
		}
	}
}

func copySumPoints(m pmetric.Metric, inM data.Instrument, mono bool) {
	s := m.SetEmptySum()
	s.SetIsMonotonic(mono)
	s.SetAggregationTemporality(toTemporality(inM.Points[0].Temporality))

	for _, inP := range inM.Points {
		dp := s.DataPoints().AppendEmpty()

		dp.SetStartTimestamp(pcommon.NewTimestampFromTime(inP.Start))
		dp.SetTimestamp(pcommon.NewTimestampFromTime(inP.End))

		copyAttributes(dp.Attributes(), inP.Attributes)

		switch t := inP.Aggregation.(type) {
		case *sum.MonotonicInt64:
			dp.SetIntValue(number.ToInt64(t.Sum()))
		case *sum.NonMonotonicInt64:
			dp.SetIntValue(number.ToInt64(t.Sum()))
		case *sum.MonotonicFloat64:
			dp.SetDoubleValue(number.ToFloat64(t.Sum()))
		case *sum.NonMonotonicFloat64:
			dp.SetDoubleValue(number.ToFloat64(t.Sum()))
		}
	}
}

func copyGaugePoints(m pmetric.Metric, inM data.Instrument) {
	s := m.SetEmptyGauge()

	for _, inP := range inM.Points {
		dp := s.DataPoints().AppendEmpty()

		// Note: no start timestamp
		dp.SetTimestamp(pcommon.NewTimestampFromTime(inP.End))

		copyAttributes(dp.Attributes(), inP.Attributes)

		switch t := inP.Aggregation.(type) {
		case *gauge.Int64:
			dp.SetIntValue(number.ToInt64(t.Gauge()))
		case *gauge.Float64:
			dp.SetDoubleValue(number.ToFloat64(t.Gauge()))
		}
	}
}

func copyHistogramPoints(m pmetric.Metric, inM data.Instrument) {
	s := m.SetEmptyExponentialHistogram()

	for _, inP := range inM.Points {
		dp := s.DataPoints().AppendEmpty()

		dp.SetStartTimestamp(pcommon.NewTimestampFromTime(inP.Start))
		dp.SetTimestamp(pcommon.NewTimestampFromTime(inP.End))

		copyAttributes(dp.Attributes(), inP.Attributes)

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
			copyHistogramBuckets(dp.Positive(), t.Positive())
			copyHistogramBuckets(dp.Negative(), t.Negative())
		case *histogram.Float64:
			dp.SetSum(number.ToFloat64(t.Sum()))
			dp.SetCount(t.Count())
			dp.SetZeroCount(t.ZeroCount())
			dp.SetScale(t.Scale())
			if t.Count() != 0 {
				dp.SetMax(number.ToFloat64(t.Max()))
				dp.SetMin(number.ToFloat64(t.Min()))
			}
			copyHistogramBuckets(dp.Positive(), t.Positive())
			copyHistogramBuckets(dp.Negative(), t.Negative())
		}
	}
}

func copyHistogramBuckets(dest pmetric.ExponentialHistogramDataPointBuckets, src aggregation.Buckets) {
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

	for _, inP := range inM.Points {
		dp := s.DataPoints().AppendEmpty()

		dp.SetStartTimestamp(pcommon.NewTimestampFromTime(inP.Start))
		dp.SetTimestamp(pcommon.NewTimestampFromTime(inP.End))

		copyAttributes(dp.Attributes(), inP.Attributes)

		switch t := inP.Aggregation.(type) {
		case *histogram.Int64:
			dp.SetSum(t.Sum().CoerceToFloat64(number.Int64Kind))
			dp.SetCount(t.Count())
			if t.Count() != 0 {
				dp.SetMax(number.ToFloat64(t.Max()))
				dp.SetMin(number.ToFloat64(t.Min()))
			}
		case *histogram.Float64:
			dp.SetSum(number.ToFloat64(t.Sum()))
			dp.SetCount(t.Count())
			if t.Count() != 0 {
				dp.SetMax(number.ToFloat64(t.Max()))
				dp.SetMin(number.ToFloat64(t.Min()))
			}
		}
	}
}

func (c *client) d2pd(in data.Metrics) pmetric.Metrics {
	c.once.Do(func() {
		c.resource = pcommon.NewResource()
		copyAttributes(
			c.resource.Attributes(),
			attribute.NewSet(in.Resource.Attributes()...),
		)
	})
	out := pmetric.NewMetrics()
	rm := out.ResourceMetrics().AppendEmpty()

	c.resource.CopyTo(rm.Resource())

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
				copyHistogramPoints(m, inM)
			case *minmaxsumcount.Int64, *minmaxsumcount.Float64:
				copyMMSCPoints(m, inM)
			}
		}
	}

	return out
}
