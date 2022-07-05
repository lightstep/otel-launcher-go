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

package gauge // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/gauge"

import (
	"sync"
	"testing"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/test"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/stretchr/testify/require"
)

func TestInt64Gauge(t *testing.T) {
	test.GenericAggregatorTest[int64, Int64, Int64Methods](t, number.ToInt64)
}

func TestFloat64Gauge(t *testing.T) {
	test.GenericAggregatorTest[float64, Float64, Float64Methods](t, number.ToFloat64)
}

func TestLastValue(t *testing.T) {
	genericLastValueTest[float64, Float64, Float64Methods](t, number.ToFloat64)
	genericLastValueTest[int64, Int64, Int64Methods](t, number.ToInt64)
}

func genericLastValueTest[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]](t *testing.T, nf func(number.Number) N) {
	var methods Methods
	init := func() *Storage {
		var s Storage
		methods.Init(&s, aggregator.Config{})
		return &s
	}

	t.Run("copy1", func(t *testing.T) {
		input := init()
		output := init()

		const ops = 1e5
		const workers = 10

		var updaters sync.WaitGroup
		updaters.Add(workers)

		for i := 0; i < workers; i++ {
			go func(i int) {
				defer updaters.Done()

				for j := 0; j < ops/workers; j++ {
					methods.Update(input, N(i+1))
				}
			}(i)
		}

		updaters.Wait()

		methods.Move(input, output)

		g, ok := methods.ToAggregation(output).(aggregation.Gauge)
		require.True(t, ok)
		require.LessOrEqual(t, N(1), nf(g.Gauge()))
		require.GreaterOrEqual(t, N(workers+1), nf(g.Gauge()))

		require.True(t, methods.HasChange(output))
		require.True(t, !methods.HasChange(input))
	})

	t.Run("copy2", func(t *testing.T) {
		in := init()
		out := init()
		methods.Update(in, 17)
		methods.Copy(in, out)

		require.Equal(t, in, out)
	})

	t.Run("merge1", func(t *testing.T) {
		first := init()
		second := init()
		methods.Update(first, 17)
		methods.Update(second, 23)

		methods.Merge(first, second)

		var methods Methods
		require.True(t, methods.HasChange(second))
		agg := methods.ToAggregation(second)
		require.Equal(t, N(23), nf(agg.(aggregation.Gauge).Gauge()))
	})

	t.Run("merge2", func(t *testing.T) {
		first := init()
		second := init()
		methods.Update(second, 23)
		methods.Update(first, 17)

		methods.Merge(first, second)

		var methods Methods
		require.True(t, methods.HasChange(second))
		agg := methods.ToAggregation(second)
		require.Equal(t, N(17), nf(agg.(aggregation.Gauge).Gauge()))
	})
}
