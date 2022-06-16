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

package minmaxsumcount // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/minmaxsumcount"

import (
	"testing"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/test"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/stretchr/testify/require"
)

func TestInt64Minmaxsumcount(t *testing.T) {
	test.GenericAggregatorTest[int64, Int64, Int64Methods](t, number.ToInt64)
}

func TestFloat64Minmaxsumcount(t *testing.T) {
	test.GenericAggregatorTest[float64, Float64, Float64Methods](t, number.ToFloat64)
}

func TestLastValue(t *testing.T) {
	genericMinMaxSumCountTest[float64, Float64, Float64Methods](t, number.ToFloat64)
	genericMinMaxSumCountTest[int64, Int64, Int64Methods](t, number.ToInt64)
}

func genericMinMaxSumCountTest[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]](t *testing.T, nf func(number.Number) N) {
	var methods Methods
	init := func(vals ...N) *Storage {
		var s Storage
		methods.Init(&s, aggregator.Config{})
		for _, val := range vals {
			methods.Update(&s, val)
		}
		return &s
	}

	t.Run("correct", func(t *testing.T) {
		in := init(3, 1, 10, 20, 500)
		agg := methods.ToAggregation(in).(aggregation.MinMaxSumCount)

		require.Equal(t, N(500), nf(agg.Max()))
		require.Equal(t, N(1), nf(agg.Min()))
		require.Equal(t, N(534), nf(agg.Sum()))
		require.Equal(t, uint64(5), agg.Count())
	})

	t.Run("copy", func(t *testing.T) {
		in := init(1, 2, 3)
		out := init()
		methods.Update(in, 4)
		methods.Copy(in, out)

		require.Equal(t, in, out)
	})

	t.Run("merge", func(t *testing.T) {
		first := init(1, 2, 3)
		second := init(4, 5, 6)

		methods.Update(first, 7)
		methods.Update(second, 8)

		expect := init(1, 2, 3, 4, 5, 6, 7, 8)

		methods.Merge(first, second)

		require.Equal(t, expect, second)
	})
}
