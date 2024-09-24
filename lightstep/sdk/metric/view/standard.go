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

package view // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"

import (
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
)

// StandardAggregationKind returns a function that configures the
// specified default aggregation Kind for each instrument Kind.
func StandardAggregationKind(ik sdkinstrument.Kind) aggregation.Kind {
	switch ik {
	case sdkinstrument.SyncHistogram:
		// Note: the default is Exponential Histogram, not MinMaxSumCount.
		return aggregation.HistogramKind
	case sdkinstrument.AsyncGauge, sdkinstrument.SyncGauge:
		return aggregation.GaugeKind
	case sdkinstrument.SyncUpDownCounter, sdkinstrument.AsyncUpDownCounter:
		return aggregation.NonMonotonicSumKind
	default:
		return aggregation.MonotonicSumKind
	}
}

// StandardTemporality returns a function that conifigures the
// specified default Cumulative temporality for all instrument kinds.
func StandardTemporality(ik sdkinstrument.Kind) aggregation.Temporality {
	return aggregation.CumulativeTemporality
}

// DeltaPreferredTemporality returns a function that configures a
// preference for Delta temporality for all instrument kinds except
// UpDownCounter, which remain Cumulative.
func DeltaPreferredTemporality(ik sdkinstrument.Kind) aggregation.Temporality {
	switch ik {
	case sdkinstrument.SyncUpDownCounter, sdkinstrument.AsyncUpDownCounter:
		return aggregation.CumulativeTemporality
	default:
		return aggregation.DeltaTemporality
	}
}

// StandardConfigForPerformance returns a function that configures two
// default aggregator.Configs using the specified performance
// defaults.
func StandardConfigForPerformance(perf sdkinstrument.Performance) func(ik sdkinstrument.Kind) (ints, floats aggregator.Config) {
	perf = perf.Validate()

	// Histogram settings are defaulted.
	noExs := aggregator.Config{
		CardinalityLimit: perf.AggregatorCardinalityLimit,
		Exemplar: aggregator.ExemplarConfig{
			Filter: aggregator.AlwaysOffKind,
		},
	}
	// Exemplar settings are restricted by instrument kind.
	withExs := noExs
	if perf.ExemplarsEnabled > 0 {
		withExs.Exemplar = aggregator.ExemplarConfig{
			Filter: aggregator.WhenTracedKind,
			Size:   perf.ExemplarsEnabled,
		}
	}
	return func(ik sdkinstrument.Kind) (ints, floats aggregator.Config) {
		switch ik {
		case sdkinstrument.SyncCounter, sdkinstrument.SyncHistogram:
			return withExs, withExs
		default:
			return noExs, noExs
		}
	}
}
