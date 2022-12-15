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

package histogram // import "github.com/lightstep/go-expohisto"

import (
	"testing"

	"github.com/lightstep/go-expohisto/structure"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/test"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/stretchr/testify/require"
)

func RequireEqualValues[N structure.ValueType, Traits number.Traits[N]](t *testing.T, a, b *Histogram[N, Traits]) {
	require.Equal(t, a.Scale(), b.Scale())
	require.Equal(t, a.Count(), b.Count())
	require.Equal(t, a.Sum(), b.Sum())
	require.Equal(t, a.Min(), b.Min())
	require.Equal(t, a.Max(), b.Max())
	requireEqualBuckets(t, a.Positive(), b.Positive())
	requireEqualBuckets(t, a.Negative(), b.Negative())
}

func requireEqualBuckets(t *testing.T, a, b aggregation.Buckets) {
	require.Equal(t, a.Len(), b.Len())
	require.Equal(t, a.Offset(), b.Offset())
	for i := uint32(0); i < a.Len(); i++ {
		require.Equal(t, a.At(i), b.At(i))
	}
}

func TestCopyMove(t *testing.T) {
	var mf Float64Methods

	h1 := NewFloat64(NewConfig(), 1, 3, 5, 7, 9, -1, -3, -5)
	h2 := NewFloat64(NewConfig())
	h3 := NewFloat64(NewConfig())

	require.Equal(t, -5.0, number.ToFloat64(h1.Min()))
	require.Equal(t, 9.0, number.ToFloat64(h1.Max()))

	mf.Move(h1, h2)
	mf.Copy(h2, h3)

	RequireEqualValues(t, h2, h3)
}

func TestMerge(t *testing.T) {
	var mf Float64Methods

	h1 := NewFloat64(NewConfig(), 1, 2, 3)
	h2 := NewFloat64(NewConfig(), 4, 5, 6)
	h3 := NewFloat64(NewConfig(), 7, 8, 9)
	h4 := NewFloat64(NewConfig())

	mf.Merge(h1, h4)
	mf.Merge(h2, h4)
	mf.Merge(h3, h4)

	require.Equal(t, 1.0, number.ToFloat64(h4.Min()))
	require.Equal(t, 9.0, number.ToFloat64(h4.Max()))

	h5 := NewFloat64(NewConfig(), 1, 2, 3, 4, 5, 6, 7, 8, 9)

	RequireEqualValues(t, h5, h4)
}

func TestAggregatorToFrom(t *testing.T) {
	var mi Int64Methods
	var mf Float64Methods
	var hi Int64

	hs, ok := mi.ToStorage(mi.ToAggregation(&hi))
	require.Equal(t, &hi, hs)
	require.True(t, ok)

	_, ok = mf.ToStorage(mi.ToAggregation(&hi))
	require.False(t, ok)
}

// Tests that the aggregation kind is correct.
func TestAggregationKind(t *testing.T) {
	require.Equal(t, aggregation.HistogramKind, NewInt64(NewConfig()).Kind())
	require.Equal(t, aggregation.HistogramKind, NewFloat64(NewConfig()).Kind())
}

func TestInt64Histogram(t *testing.T) {
	test.GenericAggregatorTest[int64, Int64, Int64Methods](t, number.ToInt64)
}

func TestFloat64Histogram(t *testing.T) {
	test.GenericAggregatorTest[float64, Float64, Float64Methods](t, number.ToFloat64)
}
