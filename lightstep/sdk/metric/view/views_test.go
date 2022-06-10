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

package view

import (
	"regexp"
	"testing"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

func TestClauseMatches(t *testing.T) {
	re := regexp.MustCompile("s.+")
	lib0 := instrumentation.Library{
		Name: "fancy",
	}
	lib1 := instrumentation.Library{
		Name:    "limited",
		Version: "vF.G.H",
	}
	views := New("test",
		WithClause(MatchInstrumentName("single")),
		WithClause(MatchInstrumentNameRegexp(re)),
		WithClause(MatchInstrumentKind(sdkinstrument.AsyncCounter)),
		WithClause(MatchNumberKind(number.Float64Kind)),
		WithClause(MatchInstrumentationLibrary(lib0)),
		WithClause(
			MatchInstrumentationLibrary(instrumentation.Library{Version: "vF.G.H"}),
			MatchInstrumentKind(sdkinstrument.AsyncGauge),
		),
	)

	views, err := Validate(views)
	require.NoError(t, err)
	require.Equal(t, "single", views.Clauses[0].instrumentName)
	require.Equal(t, re, views.Clauses[1].instrumentNameRegexp)
	require.Equal(t, sdkinstrument.AsyncCounter, views.Clauses[2].instrumentKind)
	require.Equal(t, number.Float64Kind, views.Clauses[3].numberKind)
	require.Equal(t, lib0, views.Clauses[4].library)

	desc0 := sdkinstrument.NewDescriptor("something", sdkinstrument.AsyncCounter, number.Int64Kind, "", "")
	desc1 := sdkinstrument.NewDescriptor("single", sdkinstrument.SyncCounter, number.Int64Kind, "", "")
	desc2 := sdkinstrument.NewDescriptor("other", sdkinstrument.AsyncGauge, number.Float64Kind, "", "")

	// exact name match
	require.False(t, views.Clauses[0].Matches(lib0, desc0))
	require.True(t, views.Clauses[0].Matches(lib0, desc1))
	require.False(t, views.Clauses[0].Matches(lib0, desc2))

	// regexp name match
	require.True(t, views.Clauses[1].Matches(lib0, desc0))
	require.True(t, views.Clauses[1].Matches(lib0, desc1))
	require.False(t, views.Clauses[1].Matches(lib0, desc2))

	// instrument kind match
	require.True(t, views.Clauses[2].Matches(lib0, desc0))
	require.False(t, views.Clauses[2].Matches(lib0, desc1))
	require.False(t, views.Clauses[2].Matches(lib0, desc2))

	// number kind match
	require.False(t, views.Clauses[3].Matches(lib0, desc0))
	require.False(t, views.Clauses[3].Matches(lib0, desc1))
	require.True(t, views.Clauses[3].Matches(lib0, desc2))

	// exact library match
	require.True(t, views.Clauses[4].Matches(lib0, desc0))
	require.True(t, views.Clauses[4].Matches(lib0, desc1))
	require.True(t, views.Clauses[4].Matches(lib0, desc2))
	require.False(t, views.Clauses[4].Matches(lib1, desc0))
	require.False(t, views.Clauses[4].Matches(lib1, desc1))
	require.False(t, views.Clauses[4].Matches(lib1, desc2))

	// partial library match AND instrument kind match
	require.False(t, views.Clauses[5].Matches(lib0, desc0))
	require.False(t, views.Clauses[5].Matches(lib0, desc1))
	require.False(t, views.Clauses[5].Matches(lib0, desc2))
	require.False(t, views.Clauses[5].Matches(lib1, desc0))
	require.False(t, views.Clauses[5].Matches(lib1, desc1))
	require.True(t, views.Clauses[5].Matches(lib1, desc2))
}

func TestClauseProperties(t *testing.T) {
	views := New("test",
		WithClause(WithName("longname"), MatchInstrumentName("single")),
		WithClause(WithDescription("very interesting")),
		WithClause(WithKeys(nil)),
		WithClause(WithKeys([]attribute.Key{})),
		WithClause(WithAggregation(aggregation.DropKind)),
		WithClause(WithAggregatorConfig(aggregator.Config{
			Histogram: aggregator.HistogramConfig{
				MaxSize: 177,
			},
		})),
	)

	views, err := Validate(views)
	require.NoError(t, err)
	require.Equal(t, "longname", views.Clauses[0].Name())
	require.Equal(t, "very interesting", views.Clauses[1].Description())
	require.Equal(t, []attribute.Key(nil), views.Clauses[2].Keys())
	require.Equal(t, []attribute.Key{}, views.Clauses[3].Keys())
	require.Equal(t, aggregation.DropKind, views.Clauses[4].Aggregation())
	require.Equal(t, aggregator.Config{Histogram: aggregator.HistogramConfig{MaxSize: 177}}, views.Clauses[5].AggregatorConfig())
}

func TestNameAndRegexp(t *testing.T) {
	views := New("test", WithClause(
		MatchInstrumentName("yes"),
		MatchInstrumentNameRegexp(regexp.MustCompile("no")),
	))

	_, err := Validate(views)

	require.Error(t, err)
	require.Contains(t, err.Error(), "view has instrument name and regexp matches")
}

func TestEmptyKeyString(t *testing.T) {
	views := New("test", WithClause(
		WithKeys([]attribute.Key{
			attribute.Key(""),
		}),
	))

	_, err := Validate(views)

	require.Error(t, err)
	require.Contains(t, err.Error(), "view has empty string in keys")
}

func TestSingleNameConflict(t *testing.T) {
	views := New("test", WithClause(
		WithName("aha"),
	))

	_, err := Validate(views)

	require.Error(t, err)
	require.Contains(t, err.Error(), "multi-instrument view specifies a single name")
}

func TestStandardTemporality(t *testing.T) {
	views := New("test",
		WithDefaultAggregationTemporalitySelector(StandardTemporality),
	)
	expectStandardTemporality(t, views)
}

func expectStandardTemporality(t *testing.T, v *Views) {
	for i := sdkinstrument.Kind(0); i < sdkinstrument.NumKinds; i++ {
		require.Equal(t, aggregation.CumulativeTemporality, v.Defaults.Temporality(i))
	}
}

func TestDeltaPreferredTemporality(t *testing.T) {
	views := New("test",
		WithDefaultAggregationTemporalitySelector(DeltaPreferredTemporality),
	)
	for i := sdkinstrument.Kind(0); i < sdkinstrument.NumKinds; i++ {
		switch i {
		case sdkinstrument.AsyncUpDownCounter, sdkinstrument.SyncUpDownCounter:
			require.Equal(t, aggregation.CumulativeTemporality, views.Defaults.Temporality(i))
		default:
			require.Equal(t, aggregation.DeltaTemporality, views.Defaults.Temporality(i))
		}
	}
}

func TestStandardAggregation(t *testing.T) {
	views := New("test",
		WithDefaultAggregationKindSelector(StandardAggregationKind),
	)
	expectStandardAggregation(t, views)
}

func expectStandardAggregation(t *testing.T, v *Views) {
	for i := sdkinstrument.Kind(0); i < sdkinstrument.NumKinds; i++ {
		switch i {
		case sdkinstrument.AsyncGauge:
			require.Equal(t, aggregation.GaugeKind, v.Defaults.Aggregation(i))
		case sdkinstrument.SyncCounter, sdkinstrument.AsyncCounter:
			require.Equal(t, aggregation.MonotonicSumKind, v.Defaults.Aggregation(i))
		case sdkinstrument.SyncUpDownCounter, sdkinstrument.AsyncUpDownCounter:
			require.Equal(t, aggregation.NonMonotonicSumKind, v.Defaults.Aggregation(i))
		case sdkinstrument.SyncHistogram:
			require.Equal(t, aggregation.HistogramKind, v.Defaults.Aggregation(i))
		default:
			t.Fail()
		}
	}
}

func TestInvalidViewDefaults(t *testing.T) {
	views := New("",
		WithDefaultAggregationKindSelector(func(_ sdkinstrument.Kind) aggregation.Kind {
			return 1999
		}),
		WithDefaultAggregationTemporalitySelector(func(_ sdkinstrument.Kind) aggregation.Temporality {
			return 199
		}),
		WithDefaultAggregationConfigSelector(func(_ sdkinstrument.Kind) (aggregator.Config, aggregator.Config) {
			inv := aggregator.Config{
				Histogram: aggregator.HistogramConfig{
					MaxSize: -3,
				},
			}
			return inv, inv
		}),
	)
	views, err := Validate(views)
	require.Equal(t, "", views.Name)

	require.Error(t, err)

	expectStandardAggregation(t, views)
	expectStandardTemporality(t, views)

	for i := sdkinstrument.Kind(0); i < sdkinstrument.NumKinds; i++ {
		icfg := views.Defaults.AggregationConfig(i, number.Int64Kind)
		fcfg := views.Defaults.AggregationConfig(i, number.Float64Kind)
		require.Equal(t, histogram.DefaultMaxSize, icfg.Histogram.MaxSize)
		require.Equal(t, histogram.DefaultMaxSize, fcfg.Histogram.MaxSize)
	}

	require.Contains(t, err.Error(), "invalid temporality")
	require.Contains(t, err.Error(), "invalid aggregation")
	require.Contains(t, err.Error(), "invalid histogram size")
}
