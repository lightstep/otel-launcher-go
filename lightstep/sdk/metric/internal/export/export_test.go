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
	"fmt"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/internal"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"testing"
)

// Test_d2pd tests the conversion from Lightstep exponential histogram to OTel explicit bucketed histogram.
func Test_d2pd(t *testing.T) {
	tcs := []struct {
		name  string
		input map[float64]uint64
	}{
		{
			name: "zero bucket",
			input: map[float64]uint64{
				0: 10,
			},
		},
		{
			name: "multiple values in the same bucket",
			input: map[float64]uint64{
				1.00001: 14,
				1.00002: 4,
				1.00003: 5,

				2.00001: 14,
				2.00002: 4,
				2.00003: 5,
			},
		},
		{
			name: "complex",
			input: map[float64]uint64{
				1:    14,
				6:    4,
				1000: 5,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			h := histogram.NewFloat64(histogram.NewConfig())
			for num, incr := range tc.input {
				h.Histogram.UpdateByIncr(num, incr)
			}

			out := d2pd(&internal.ResourceMap{}, pointToMetric(h), false)
			outBuckets := populateBuckets(getSingleHistPoint(t, out))

			for number, incr := range tc.input {
				ok := func() bool {
					// iterate through buckets until we find one this falls in.
					for i := range outBuckets {
						if outBuckets[i].contains(number) {
							reduction := min(incr, outBuckets[i].value)
							incr -= reduction
							outBuckets[i].value -= reduction

							if incr == 0 {
								return true
							}
						}
					}

					return false
				}()
				require.True(t, ok, fmt.Sprintf("could not find %f (incr %d) in histogram", number, incr))
			}

			// make sure all buckets are now empty
			for _, bucket := range outBuckets {
				require.Equal(t, bucket.value, uint64(0))
			}
		})
	}
}

type bucket struct {
	start *float64
	end   *float64 // inclusive

	value uint64
}

func (b *bucket) contains(v float64) bool {
	// due to floating point rounding errors, make all the buckets *slightly* larger than
	// they actually are.
	fudgeFactor := .00001

	if b.start == nil && b.end == nil {
		panic("bucket must have start or end set")
	}

	if b.start == nil && b.end != nil {
		return v <= (*b.end + fudgeFactor)
	}

	if b.start != nil && b.end == nil {
		return v > (*b.start - fudgeFactor)
	}

	return v > (*b.start-fudgeFactor) && v <= (*b.end+fudgeFactor)
}

func (b *bucket) String() string {
	start := "..."
	if b.start != nil {
		start = fmt.Sprintf("%f", *b.start)
	}

	end := "..."
	if b.end != nil {
		end = fmt.Sprintf("%f", *b.end)
	}

	return fmt.Sprintf("(%s, %s]: %d", start, end, b.value)
}

func populateBuckets(h pmetric.HistogramDataPoint) []bucket {
	var buckets []bucket
	for i := 0; i < h.BucketCounts().Len(); i++ {
		bucketCount := h.BucketCounts().At(i)
		if bucketCount == 0 {
			continue
		}

		switch i {
		case 0: // the first bucket, upper bound
			rightBound := h.ExplicitBounds().At(0)

			buckets = append(buckets, bucket{
				end:   &rightBound,
				value: bucketCount,
			})
		case h.BucketCounts().Len() - 1: // the last bucket, lower bound
			rightBound := h.ExplicitBounds().At(i - 1)

			buckets = append(buckets, bucket{
				start: &rightBound,
				value: bucketCount,
			})
		default:
			leftBound := h.ExplicitBounds().At(i - 1)
			rightBound := h.ExplicitBounds().At(i)

			buckets = append(buckets, bucket{
				start: &leftBound,
				end:   &rightBound,
				value: bucketCount,
			})
		}
	}

	return buckets
}

func pointToMetric(point aggregation.Aggregation) data.Metrics {
	return data.Metrics{
		Scopes: []data.Scope{
			{
				Instruments: []data.Instrument{
					{
						Points: []data.Point{
							{
								Temporality: aggregation.DeltaTemporality,
								Aggregation: point,
							},
						},
					},
				},
			},
		},
	}
}

func getSingleHistPoint(t *testing.T, m pmetric.Metrics) pmetric.HistogramDataPoint {
	resourceMetrics := m.ResourceMetrics()
	require.Equal(t, resourceMetrics.Len(), 1)
	scopeMetrics := resourceMetrics.At(0).ScopeMetrics()
	require.Equal(t, scopeMetrics.Len(), 1)
	metrics := scopeMetrics.At(0).Metrics()
	require.Equal(t, metrics.Len(), 1)
	dataPoints := metrics.At(0).Histogram().DataPoints()
	require.Equal(t, dataPoints.Len(), 1)

	return dataPoints.At(0)
}
