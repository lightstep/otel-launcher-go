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

package viewstate // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/test"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"
)

var oneConflict = Conflict{
	Semantic: SemanticError{
		Instrument:  sdkinstrument.CounterKind,
		Aggregation: aggregation.GaugeKind,
	},
}

func TestViewConflictsError(t *testing.T) {
	var err error
	err = ViewConflictsError{}
	require.Equal(t, noConflictsString, err.Error())
	require.True(t, errors.Is(err, ViewConflictsError{}))

	require.True(t, errors.Is(oneConflict.Semantic, SemanticError{}))
}

// TestViewConflictsError exercises the code paths that construct example
// error messages from duplicate instrument conditions.
func TestViewConflictsBuilder(t *testing.T) {
	// Note: These all use "no conflicts" strings, which happens
	// under artificial conditions such as conflicts w/ < 2 examples
	// and allows testing the code that avoids lengthy messages
	// when there is only one conflict or only one reader.
	rd0 := "test0"
	rd1 := "test1"

	// This is a synthetic case, for the sake of coverage.
	builder := ViewConflictsBuilder{
		rd0: []Conflict{},
	}
	err := builder.AsError()
	require.Equal(t, noConflictsString, err.Error())

	// Note: This test ignores duplicates, one semantic error is
	// enough to test the ViewConflicts logic.
	oneError := oneConflict.Semantic.Error()

	builder = ViewConflictsBuilder{}
	builder.Add(rd0, oneConflict)
	err = builder.AsError()
	require.True(t, strings.HasSuffix(err.Error(), oneError), err)

	builder = ViewConflictsBuilder{}
	builder.Add(rd0, oneConflict)
	builder.Add(rd0, oneConflict)
	err = builder.AsError()
	require.True(t, strings.HasSuffix(err.Error(), oneError), err)
	require.True(t, strings.HasPrefix(err.Error(), "2 conflicts, e.g. "), err)

	builder = ViewConflictsBuilder{}
	builder.Add(rd0, oneConflict)
	builder.Add(rd1, oneConflict)
	err = builder.AsError()
	require.True(t, strings.HasSuffix(err.Error(), oneError), err)
	require.True(t, strings.HasPrefix(err.Error(), "2 conflicts in 2 readers, e.g. "), err)
}

func TestConflictCombine(t *testing.T) {
	rd0 := "test0"
	rd1 := "test1"

	builder1 := ViewConflictsBuilder{}
	builder1.Add(rd0, oneConflict)

	builder2 := ViewConflictsBuilder{}
	builder2.Add(rd1, oneConflict)

	builder1.Combine(builder2)
	err1 := builder1.AsError()
	require.True(t, strings.HasSuffix(err1.Error(), oneConflict.Semantic.Error()), err1)
	require.True(t, strings.HasPrefix(err1.Error(), "2 conflicts in 2 readers, e.g. "), err1)

	var builder3 ViewConflictsBuilder
	builder3.Combine(ViewConflictsBuilder{}) // empty builder has no effect
	builder3.Combine(builder1)
	err3 := builder3.AsError()

	require.Equal(t, err1, err3)
}

// TestConflictError tests that both semantic errors and duplicate
// conflicts are printed.  Note this uses the real library to generate
// the conflict, to avoid creating a relatively large test-only type.
func TestConflictError(t *testing.T) {
	views := view.New(
		"problem",
		view.WithDefaultAggregationKindSelector(func(k sdkinstrument.Kind) aggregation.Kind {
			return aggregation.GaugeKind
		}),
		view.WithClause(
			// "bar" is renamed "foo" w/ histogram
			view.MatchInstrumentName("bar"),
			view.WithName("foo"),
			view.WithAggregation(aggregation.HistogramKind),
		),
	)

	vc := New(testLib, views)

	// Sync counter named bar becomes histogram
	inst2, conf2 := vc.Compile(test.Descriptor("bar", sdkinstrument.CounterKind, number.Int64Kind))
	require.NoError(t, conf2.AsError())
	require.NotNil(t, inst2)

	// Async counter named foo becomes gauge
	inst1, conf1 := vc.Compile(test.Descriptor("foo", sdkinstrument.CounterObserverKind, number.Int64Kind))
	require.Error(t, conf1.AsError())
	require.NotNil(t, inst1)
	require.Equal(t,
		"problem: CounterObserver instrument incompatible with Gauge aggregation; "+
			"name \"foo\" (original \"bar\") conflicts Counter-Int64-Histogram, CounterObserver-Int64-MonotonicSum",
		conf1.AsError().Error(),
	)

	require.NotEqual(t, inst1, inst2)
}
