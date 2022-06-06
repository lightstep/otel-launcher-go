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
	case sdkinstrument.HistogramKind:
		return aggregation.HistogramKind
	case sdkinstrument.GaugeObserverKind:
		return aggregation.GaugeKind
	case sdkinstrument.UpDownCounterKind, sdkinstrument.UpDownCounterObserverKind:
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
	case sdkinstrument.UpDownCounterKind, sdkinstrument.UpDownCounterObserverKind:
		return aggregation.CumulativeTemporality
	default:
		return aggregation.DeltaTemporality
	}
}

// StandardConfig returns a function that configures two default aggregator.Configs.
func StandardConfig(ik sdkinstrument.Kind) (ints, floats aggregator.Config) {
	return aggregator.Config{}, aggregator.Config{}
}
