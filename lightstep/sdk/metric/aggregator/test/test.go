package test

import (
	"context"
	"sync"
	"testing"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/stretchr/testify/require"
)

// GenericAggregatorTest contains tests that apply to multiple aggregator packages.
func GenericAggregatorTest[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]](t *testing.T, nf func(number.Number) N) {
	t.Run("init", func(t *testing.T) {
		var storage Storage
		var methods Methods
		methods.Init(&storage, aggregator.Config{})

		agg := methods.ToAggregation(&storage)
		if g, ok := agg.(aggregation.Gauge); ok {
			require.Equal(t, N(0), nf(g.Gauge()))
		} else if h, ok := agg.(aggregation.Histogram); ok {
			require.Equal(t, uint64(0), h.Count())
			require.Equal(t, N(0), nf(h.Sum()))
		} else if s, ok := agg.(aggregation.Sum); ok {
			require.Equal(t, N(0), nf(s.Sum()))
		} else {
			t.Fail()
		}

		require.Equal(t, methods.Kind(), agg.Kind())
		require.False(t, methods.HasChange(&storage))

		st, ok := methods.ToStorage(agg)
		require.True(t, ok)
		require.Equal(t, st, &storage)
	})

	t.Run("add_merge", func(t *testing.T) {
		var input Storage
		var intermediate Storage
		var output Storage
		var methods Methods

		methods.Init(&input, aggregator.Config{})
		methods.Init(&intermediate, aggregator.Config{})
		methods.Init(&output, aggregator.Config{})

		// Tests Counter and Histogram; excludes Gauge.
		if _, ok := methods.ToAggregation(&intermediate).(aggregation.HasASum); !ok {
			t.Skip()
			return
		}

		const ops = 1e5
		const workers = 10

		var updaters sync.WaitGroup
		var mergers sync.WaitGroup
		updaters.Add(workers)
		mergers.Add(workers)

		for i := 0; i < workers; i++ {
			go func() {
				defer updaters.Done()

				for j := 0; j < ops/workers; j++ {
					methods.Update(&input, 1)
				}
			}()
		}

		ctx, cancel := context.WithCancel(context.Background())

		for i := 0; i < workers; i++ {
			go func() {
				defer mergers.Done()

				var mine Storage
				methods.Init(&mine, aggregator.Config{})

				collect := func() {
					methods.Move(&input, &mine)
					methods.Merge(&mine, &intermediate)
				}

				for {
					select {
					case <-ctx.Done():
						collect()
						return
					default:
						collect()
					}
				}
			}()
		}

		updaters.Wait()
		cancel()
		mergers.Wait()
		methods.Copy(&intermediate, &output)

		s, ok := methods.ToAggregation(&output).(aggregation.HasASum)
		require.True(t, ok)
		require.Equal(t, N(ops), nf(s.Sum()))

		require.True(t, methods.HasChange(&output))
		require.True(t, !methods.HasChange(&input))
	})
}
