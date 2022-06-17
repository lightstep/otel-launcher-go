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
	"fmt"
	"math"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/gauge"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/minmaxsumcount"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/sum"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/test"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

var (
	testLib = instrumentation.Library{
		Name: "test",
	}

	fooToBarView = view.WithClause(
		view.MatchInstrumentName("foo"),
		view.WithName("bar"),
	)

	testHistBoundaries = []float64{1, 2, 3}

	defaultAggregatorConfig = aggregator.Config{}

	altHistogramConfig = aggregator.Config{
		Histogram: aggregator.HistogramConfig{
			MaxSize: 15,
		},
	}

	fooToBarAltHistView = view.WithClause(
		view.MatchInstrumentName("foo"),
		view.WithName("bar"),
		view.WithAggregatorConfig(altHistogramConfig),
	)

	fooToBarFilteredView = view.WithClause(
		view.MatchInstrumentName("foo"),
		view.WithName("bar"),
		view.WithKeys([]attribute.Key{"a", "b"}),
	)

	fooToBarDifferentFiltersViews = []view.Option{
		fooToBarFilteredView,
		view.WithClause(
			view.MatchInstrumentName("bar"),
			view.WithKeys([]attribute.Key{"a"}),
		),
	}

	fooToBarSameFiltersViews = []view.Option{
		fooToBarFilteredView,
		view.WithClause(
			view.MatchInstrumentName("bar"),
			view.WithKeys([]attribute.Key{"a", "b"}),
		),
	}

	dropHistInstView = view.WithClause(
		view.MatchInstrumentKind(sdkinstrument.SyncHistogram),
		view.WithAggregation(aggregation.DropKind),
	)

	instrumentKinds = []sdkinstrument.Kind{
		sdkinstrument.SyncHistogram,
		sdkinstrument.AsyncGauge,
		sdkinstrument.SyncCounter,
		sdkinstrument.SyncUpDownCounter,
		sdkinstrument.AsyncCounter,
		sdkinstrument.AsyncUpDownCounter,
	}

	numberKinds = []number.Kind{
		number.Int64Kind,
		number.Float64Kind,
	}

	endTime    = time.Now()
	middleTime = endTime.Add(-time.Millisecond)
	startTime  = endTime.Add(-2 * time.Millisecond)

	testSequence = data.Sequence{
		Start: startTime,
		Last:  middleTime,
		Now:   endTime,
	}
)

const (
	cumulative = aggregation.CumulativeTemporality
	delta      = aggregation.DeltaTemporality
)

func testCompile(vc *Compiler, name string, ik sdkinstrument.Kind, nk number.Kind, opts ...instrument.Option) (Instrument, error) {
	inst, conflicts := vc.Compile(test.Descriptor(name, ik, nk, opts...))
	return inst, conflicts.AsError()
}

func testCollect(t *testing.T, vc *Compiler) []data.Instrument {
	return test.CollectScope(t, vc.Collectors(), testSequence)
}

func testCollectSequence(t *testing.T, vc *Compiler, seq data.Sequence) []data.Instrument {
	return test.CollectScope(t, vc.Collectors(), seq)
}

func testCollectSequenceReuse(t *testing.T, vc *Compiler, seq data.Sequence, output *data.Scope) []data.Instrument {
	return test.CollectScopeReuse(t, vc.Collectors(), seq, output)
}

// TestDeduplicateNoConflict verifies that two identical instruments
// have the same collector.
func TestDeduplicateNoConflict(t *testing.T) {
	vc := New(testLib, view.New("test"))

	inst1, err1 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Int64Kind)
	require.NoError(t, err1)
	require.NotNil(t, inst1)

	inst2, err2 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Int64Kind)
	require.NoError(t, err2)
	require.NotNil(t, inst2)

	require.Equal(t, inst1, inst2)
}

// TestDeduplicateRenameNoConflict verifies that one instrument can be renamed
// such that it becomes identical to another, so no conflict.
func TestDeduplicateRenameNoConflict(t *testing.T) {
	vc := New(testLib, view.New("test", fooToBarView))

	inst1, err1 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Int64Kind)
	require.NoError(t, err1)
	require.NotNil(t, inst1)

	inst2, err2 := testCompile(vc, "bar", sdkinstrument.SyncCounter, number.Int64Kind)
	require.NoError(t, err2)
	require.NotNil(t, inst2)

	require.Equal(t, inst1, inst2)
}

// TestNoRenameNoConflict verifies that one instrument does not
// conflict with another differently-named instrument.
func TestNoRenameNoConflict(t *testing.T) {
	vc := New(testLib, view.New("test"))

	inst1, err1 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Int64Kind)
	require.NoError(t, err1)
	require.NotNil(t, inst1)

	inst2, err2 := testCompile(vc, "bar", sdkinstrument.SyncCounter, number.Int64Kind)
	require.NoError(t, err2)
	require.NotNil(t, inst2)

	require.NotEqual(t, inst1, inst2)
}

// TestDuplicateNumberConflict verifies that two same instruments
// except different number kind conflict.
func TestDuplicateNumberConflict(t *testing.T) {
	vc := New(testLib, view.New("test"))

	inst1, err1 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Int64Kind)
	require.NoError(t, err1)
	require.NotNil(t, inst1)

	inst2, err2 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Float64Kind)
	require.Error(t, err2)
	require.NotNil(t, inst2)
	require.True(t, errors.Is(err2, ViewConflictsError{}))
	require.Equal(t, 1, len(err2.(ViewConflictsError)))
	require.Equal(t, 1, len(err2.(ViewConflictsError)["test"]))
	require.Equal(t, 2, len(err2.(ViewConflictsError)["test"][0].Duplicates))

	require.NotEqual(t, inst1, inst2)
}

// TestDuplicateSyncAsyncConflict verifies that two same instruments
// except one synchonous, one asynchronous conflict.
func TestDuplicateSyncAsyncConflict(t *testing.T) {
	vc := New(testLib, view.New("test"))

	inst1, err1 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Float64Kind)
	require.NoError(t, err1)
	require.NotNil(t, inst1)

	inst2, err2 := testCompile(vc, "foo", sdkinstrument.AsyncCounter, number.Float64Kind)
	require.Error(t, err2)
	require.NotNil(t, inst2)
	require.True(t, errors.Is(err2, ViewConflictsError{}))

	require.NotEqual(t, inst1, inst2)
}

// TestDuplicateUnitConflict verifies that two same instruments
// except different units conflict.
func TestDuplicateUnitConflict(t *testing.T) {
	vc := New(testLib, view.New("test"))

	inst1, err1 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Float64Kind, instrument.WithUnit("gal_us"))
	require.NoError(t, err1)
	require.NotNil(t, inst1)

	inst2, err2 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Float64Kind, instrument.WithUnit("cft_i"))
	require.Error(t, err2)
	require.NotNil(t, inst2)
	require.True(t, errors.Is(err2, ViewConflictsError{}))
	require.Contains(t, err2.Error(), "test: name \"foo\" conflicts SyncCounter-Float64-MonotonicSum-gal_us")

	require.NotEqual(t, inst1, inst2)
}

// TestDuplicateMonotonicConflict verifies that two same instruments
// except different monotonic values.
func TestDuplicateMonotonicConflict(t *testing.T) {
	vc := New(testLib, view.New("test"))

	inst1, err1 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Float64Kind)
	require.NoError(t, err1)
	require.NotNil(t, inst1)

	inst2, err2 := testCompile(vc, "foo", sdkinstrument.SyncUpDownCounter, number.Float64Kind)
	require.Error(t, err2)
	require.NotNil(t, inst2)
	require.True(t, errors.Is(err2, ViewConflictsError{}))
	require.Contains(t, err2.Error(), "UpDownCounter-Float64-NonMonotonicSum")

	require.NotEqual(t, inst1, inst2)
}

// TestDuplicateAggregatorConfigConflict verifies that two same instruments
// except different aggregator.Config values.
func TestDuplicateAggregatorConfigConflict(t *testing.T) {
	vc := New(testLib, view.New("test", fooToBarAltHistView))

	inst1, err1 := testCompile(vc, "foo", sdkinstrument.SyncHistogram, number.Float64Kind)
	require.NoError(t, err1)
	require.NotNil(t, inst1)

	inst2, err2 := testCompile(vc, "bar", sdkinstrument.SyncHistogram, number.Float64Kind)
	require.Error(t, err2)
	require.NotNil(t, inst2)
	require.True(t, errors.Is(err2, ViewConflictsError{}))
	require.Contains(t, err2.Error(), "different aggregator configuration")

	require.NotEqual(t, inst1, inst2)
}

// TestDuplicateAggregatorConfigNoConflict verifies that two same instruments
// with same aggregator.Config values configured in different ways.
func TestDuplicateAggregatorConfigNoConflict(t *testing.T) {
	for _, nk := range numberKinds {
		t.Run(nk.String(), func(t *testing.T) {
			views := view.New(
				"test",
				view.WithDefaultAggregationConfigSelector(
					func(_ sdkinstrument.Kind) (int64Config, float64Config aggregator.Config) {
						if nk == number.Int64Kind {
							return altHistogramConfig, aggregator.Config{}
						}
						return aggregator.Config{}, altHistogramConfig
					},
				),
				fooToBarAltHistView,
			)

			vc := New(testLib, views)

			inst1, err1 := testCompile(vc, "foo", sdkinstrument.SyncHistogram, nk)
			require.NoError(t, err1)
			require.NotNil(t, inst1)

			inst2, err2 := testCompile(vc, "bar", sdkinstrument.SyncHistogram, nk)
			require.NoError(t, err2)
			require.NotNil(t, inst2)

			require.Equal(t, inst1, inst2)
		})
	}
}

// TestDuplicateAggregationKindConflict verifies that two instruments
// with different aggregation kinds conflict.
func TestDuplicateAggregationKindConflict(t *testing.T) {
	vc := New(testLib, view.New("test", fooToBarView))

	inst1, err1 := testCompile(vc, "foo", sdkinstrument.SyncHistogram, number.Int64Kind)
	require.NoError(t, err1)
	require.NotNil(t, inst1)

	inst2, err2 := testCompile(vc, "bar", sdkinstrument.SyncCounter, number.Int64Kind)
	require.Error(t, err2)
	require.NotNil(t, inst2)
	require.True(t, errors.Is(err2, ViewConflictsError{}))
	require.Contains(t, err2.Error(), "name \"bar\" (original \"foo\") conflicts SyncHistogram-Int64-Histogram, SyncCounter-Int64-MonotonicSum")

	require.NotEqual(t, inst1, inst2)
}

// TestDuplicateAggregationKindNoConflict verifies that two
// instruments with different aggregation kinds do not conflict when
// the view drops one of the instruments.
func TestDuplicateAggregationKindNoConflict(t *testing.T) {
	vc := New(testLib, view.New("test", dropHistInstView))

	inst1, err1 := testCompile(vc, "foo", sdkinstrument.SyncHistogram, number.Int64Kind)
	require.NoError(t, err1)
	require.Nil(t, inst1) // The viewstate.Instrument is nil, instruments become no-ops.

	inst2, err2 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Int64Kind)
	require.NoError(t, err2)
	require.NotNil(t, inst2)
}

// TestDuplicateMultipleConflicts verifies that multiple duplicate
// instrument conflicts include sufficient explanatory information.
func TestDuplicateMultipleConflicts(t *testing.T) {
	vc := New(testLib, view.New("test"))

	inst1, err1 := testCompile(vc, "foo", instrumentKinds[0], number.Float64Kind)
	require.NoError(t, err1)
	require.NotNil(t, inst1)

	for num, ik := range instrumentKinds[1:] {
		inst2, err2 := testCompile(vc, "foo", ik, number.Float64Kind)
		require.Error(t, err2)
		require.NotNil(t, inst2)
		require.True(t, errors.Is(err2, ViewConflictsError{}))
		// The total number of conflicting definitions is 1 in
		// the first place and num+1 for the iterations of this loop.
		require.Equal(t, num+2, len(err2.(ViewConflictsError)["test"][0].Duplicates))

		if num > 0 {
			require.Contains(t, err2.Error(), fmt.Sprintf("and %d more", num))
		}
	}
}

// TestDuplicateFilterConflicts verifies several cases where
// instruments output the same metric w/ different filters create conflicts.
func TestDuplicateFilterConflicts(t *testing.T) {
	for idx, vws := range [][]view.Option{
		// In the first case, foo has two attribute filters bar has 0.
		{fooToBarFilteredView},
		// In the second case, foo has two attribute filters bar has 1.
		fooToBarDifferentFiltersViews,
	} {
		t.Run(fmt.Sprint(idx), func(t *testing.T) {
			vc := New(testLib, view.New("test", vws...))

			inst1, err1 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Int64Kind)
			require.NoError(t, err1)
			require.NotNil(t, inst1)

			inst2, err2 := testCompile(vc, "bar", sdkinstrument.SyncCounter, number.Int64Kind)
			require.Error(t, err2)
			require.NotNil(t, inst2)

			require.True(t, errors.Is(err2, ViewConflictsError{}))
			require.Contains(t, err2.Error(), "name \"bar\" (original \"foo\") has conflicts: different attribute filters")
		})
	}
}

// TestDeduplicateSameFilters thests that when one instrument is
// renamed to match another exactly, including filters, they are not
// in conflict.
func TestDeduplicateSameFilters(t *testing.T) {
	vc := New(testLib, view.New("test", fooToBarSameFiltersViews...))

	inst1, err1 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Int64Kind)
	require.NoError(t, err1)
	require.NotNil(t, inst1)

	inst2, err2 := testCompile(vc, "bar", sdkinstrument.SyncCounter, number.Int64Kind)
	require.NoError(t, err2)
	require.NotNil(t, inst2)

	require.Equal(t, inst1, inst2)
}

// TestDuplicatesMergeDescriptor ensures that the longest description string is used.
func TestDuplicatesMergeDescriptor(t *testing.T) {
	vc := New(testLib, view.New("test", fooToBarSameFiltersViews...))

	inst1, err1 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Int64Kind)
	require.NoError(t, err1)
	require.NotNil(t, inst1)

	// This is the winning description:
	inst2, err2 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Int64Kind, instrument.WithDescription("very long"))
	require.NoError(t, err2)
	require.NotNil(t, inst2)

	inst3, err3 := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Int64Kind, instrument.WithDescription("shorter"))
	require.NoError(t, err3)
	require.NotNil(t, inst3)

	require.Equal(t, inst1, inst2)
	require.Equal(t, inst1, inst3)

	accUpp := inst1.NewAccumulator(attribute.NewSet())
	accUpp.(Updater[int64]).Update(1)

	accUpp.SnapshotAndProcess(false)

	output := testCollect(t, vc)

	require.Equal(t, 1, len(output))
	require.Equal(t, test.Instrument(
		test.Descriptor("bar", sdkinstrument.SyncCounter, number.Int64Kind, instrument.WithDescription("very long")),
		test.Point(startTime, endTime, sum.NewMonotonicInt64(1), cumulative)), output[0],
	)
}

// TestViewDescription ensures that a View can override the description.
func TestViewDescription(t *testing.T) {
	views := view.New(
		"test",
		view.WithClause(
			view.MatchInstrumentName("foo"),
			view.WithDescription("something helpful"),
		),
	)

	vc := New(testLib, views)

	inst1, err1 := testCompile(vc,
		"foo", sdkinstrument.SyncCounter, number.Int64Kind,
		instrument.WithDescription("other description"),
	)
	require.NoError(t, err1)
	require.NotNil(t, inst1)

	attrs := []attribute.KeyValue{
		attribute.String("K", "V"),
	}
	accUpp := inst1.NewAccumulator(attribute.NewSet(attrs...))
	accUpp.(Updater[int64]).Update(1)

	accUpp.SnapshotAndProcess(false)

	output := testCollect(t, vc)

	require.Equal(t, 1, len(output))
	require.Equal(t,
		test.Instrument(
			test.Descriptor(
				"foo", sdkinstrument.SyncCounter, number.Int64Kind,
				instrument.WithDescription("something helpful"),
			),
			test.Point(startTime, endTime, sum.NewMonotonicInt64(1), cumulative, attribute.String("K", "V")),
		),
		output[0],
	)
}

// TestKeyFilters verifies that keys are filtred and metrics are
// correctly aggregated.
func TestKeyFilters(t *testing.T) {
	views := view.New("test",
		view.WithClause(view.WithKeys([]attribute.Key{"a", "b"})),
	)

	vc := New(testLib, views)

	inst, err := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Int64Kind)
	require.NoError(t, err)
	require.NotNil(t, inst)

	accUpp1 := inst.NewAccumulator(
		attribute.NewSet(attribute.String("a", "1"), attribute.String("b", "2"), attribute.String("c", "3")),
	)
	accUpp2 := inst.NewAccumulator(
		attribute.NewSet(attribute.String("a", "1"), attribute.String("b", "2"), attribute.String("d", "4")),
	)

	accUpp1.(Updater[int64]).Update(1)
	accUpp2.(Updater[int64]).Update(1)
	accUpp1.SnapshotAndProcess(false)
	accUpp2.SnapshotAndProcess(false)

	output := testCollect(t, vc)

	require.Equal(t, 1, len(output))
	require.Equal(t, test.Instrument(
		test.Descriptor("foo", sdkinstrument.SyncCounter, number.Int64Kind),
		test.Point(
			startTime, endTime, sum.NewMonotonicInt64(2), cumulative,
			attribute.String("a", "1"), attribute.String("b", "2"),
		)), output[0],
	)
}

// TestTwoViewsOneInt64Instrument verifies that multiple int64
// instrument behaviors work; in this case, viewing a Sum in each
// of three independent dimensions.
func TestTwoViewsOneInt64Instrument(t *testing.T) {
	views := view.New(
		"test",
		view.WithClause(
			view.MatchInstrumentName("foo"),
			view.WithName("foo_a"),
			view.WithKeys([]attribute.Key{"a"}),
		),
		view.WithClause(
			view.MatchInstrumentName("foo"),
			view.WithName("foo_b"),
			view.WithKeys([]attribute.Key{"b"}),
		),
		view.WithClause(
			view.MatchInstrumentName("foo"),
			view.WithName("foo_c"),
			view.WithKeys([]attribute.Key{"c"}),
		),
	)

	vc := New(testLib, views)

	inst, err := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Int64Kind)
	require.NoError(t, err)

	for _, acc := range []Accumulator{
		inst.NewAccumulator(attribute.NewSet(attribute.String("a", "1"), attribute.String("b", "1"))),
		inst.NewAccumulator(attribute.NewSet(attribute.String("a", "1"), attribute.String("b", "2"))),
		inst.NewAccumulator(attribute.NewSet(attribute.String("a", "2"), attribute.String("b", "1"))),
		inst.NewAccumulator(attribute.NewSet(attribute.String("a", "2"), attribute.String("b", "2"))),
	} {
		acc.(Updater[int64]).Update(1)
		acc.SnapshotAndProcess(false)
	}

	output := testCollect(t, vc)

	test.RequireEqualMetrics(t,
		output,
		test.Instrument(
			test.Descriptor("foo_a", sdkinstrument.SyncCounter, number.Int64Kind),
			test.Point(
				startTime, endTime, sum.NewMonotonicInt64(2), cumulative, attribute.String("a", "1"),
			),
			test.Point(
				startTime, endTime, sum.NewMonotonicInt64(2), cumulative, attribute.String("a", "2"),
			),
		),
		test.Instrument(
			test.Descriptor("foo_b", sdkinstrument.SyncCounter, number.Int64Kind),
			test.Point(
				startTime, endTime, sum.NewMonotonicInt64(2), cumulative, attribute.String("b", "1"),
			),
			test.Point(
				startTime, endTime, sum.NewMonotonicInt64(2), cumulative, attribute.String("b", "2"),
			),
		),
		test.Instrument(
			test.Descriptor("foo_c", sdkinstrument.SyncCounter, number.Int64Kind),
			test.Point(
				startTime, endTime, sum.NewMonotonicInt64(4), cumulative,
			),
		),
	)
}

// TestHistogramTwoAggregations verifies that two float64 instrument
// behaviors are correctly combined, in this case one sum and one histogram.
func TestHistogramTwoAggregations(t *testing.T) {
	views := view.New(
		"test",
		view.WithClause(
			view.MatchInstrumentName("foo"),
			view.WithName("foo_sum"),
			view.WithAggregation(aggregation.MonotonicSumKind),
			view.WithKeys([]attribute.Key{}),
		),
		view.WithClause(
			view.MatchInstrumentName("foo"),
			view.WithName("foo_hist"),
			view.WithAggregation(aggregation.HistogramKind),
		),
	)

	vc := New(testLib, views)

	inst, err := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Float64Kind)
	require.NoError(t, err)

	acc := inst.NewAccumulator(attribute.NewSet())
	acc.(Updater[float64]).Update(1)
	acc.(Updater[float64]).Update(2)
	acc.(Updater[float64]).Update(3)
	acc.(Updater[float64]).Update(4)
	acc.SnapshotAndProcess(false)

	output := testCollect(t, vc)

	test.RequireEqualMetrics(t, output,
		test.Instrument(
			test.Descriptor("foo_sum", sdkinstrument.SyncCounter, number.Float64Kind),
			test.Point(
				startTime, endTime, sum.NewMonotonicFloat64(10), cumulative,
			),
		),
		test.Instrument(
			test.Descriptor("foo_hist", sdkinstrument.SyncCounter, number.Float64Kind),
			test.Point(
				startTime, endTime, histogram.NewFloat64(defaultAggregatorConfig.Histogram, 1, 2, 3, 4), cumulative,
			),
		),
	)
}

// TestAllKeysFilter tests that view.WithKeys([]attribute.Key{})
// correctly erases all keys.
func TestAllKeysFilter(t *testing.T) {
	views := view.New(
		"test",
		view.WithClause(view.WithKeys([]attribute.Key{})),
	)

	vc := New(testLib, views)

	inst, err := testCompile(vc, "foo", sdkinstrument.SyncCounter, number.Float64Kind)
	require.NoError(t, err)

	acc1 := inst.NewAccumulator(attribute.NewSet(attribute.String("a", "1")))
	acc1.(Updater[float64]).Update(1)
	acc1.SnapshotAndProcess(false)

	acc2 := inst.NewAccumulator(attribute.NewSet(attribute.String("b", "2")))
	acc2.(Updater[float64]).Update(1)
	acc2.SnapshotAndProcess(false)

	output := testCollect(t, vc)

	test.RequireEqualMetrics(t, output,
		test.Instrument(
			test.Descriptor("foo", sdkinstrument.SyncCounter, number.Float64Kind),
			test.Point(
				startTime, endTime, sum.NewMonotonicFloat64(2), cumulative,
			),
		),
	)
}

// TestAnySumAggregation checks that the proper aggregation inference
// is performed for each of the inbstrument types when
// aggregation.AnySum kind is configured.
func TestAnySumAggregation(t *testing.T) {
	views := view.New(
		"test",
		view.WithClause(view.WithAggregation(aggregation.AnySumKind)),
	)

	vc := New(testLib, views)

	for _, ik := range []sdkinstrument.Kind{
		sdkinstrument.SyncCounter,
		sdkinstrument.AsyncCounter,
		sdkinstrument.SyncUpDownCounter,
		sdkinstrument.AsyncUpDownCounter,
		sdkinstrument.SyncHistogram,
		sdkinstrument.AsyncGauge,
	} {
		inst, err := testCompile(vc, ik.String(), ik, number.Float64Kind)
		if ik == sdkinstrument.AsyncGauge {
			// semantic conflict, Gauge can't handle AnySum aggregation!
			require.Error(t, err)
			require.Contains(t,
				err.Error(),
				"AsyncGauge instrument incompatible with Undefined aggregation",
			)
		} else {
			require.NoError(t, err)
		}

		acc := inst.NewAccumulator(attribute.NewSet())
		acc.(Updater[float64]).Update(1)
		acc.SnapshotAndProcess(false)
	}

	output := testCollect(t, vc)

	test.RequireEqualMetrics(t, output,
		test.Instrument(
			test.Descriptor("SyncCounter", sdkinstrument.SyncCounter, number.Float64Kind),
			test.Point(startTime, endTime, sum.NewMonotonicFloat64(1), cumulative), // AnySum -> Monotonic
		),
		test.Instrument(
			test.Descriptor("AsyncCounter", sdkinstrument.AsyncCounter, number.Float64Kind),
			test.Point(startTime, endTime, sum.NewMonotonicFloat64(1), cumulative), // AnySum -> Monotonic
		),
		test.Instrument(
			test.Descriptor("SyncUpDownCounter", sdkinstrument.SyncUpDownCounter, number.Float64Kind),
			test.Point(startTime, endTime, sum.NewNonMonotonicFloat64(1), cumulative), // AnySum -> Non-Monotonic
		),
		test.Instrument(
			test.Descriptor("AsyncUpDownCounter", sdkinstrument.AsyncUpDownCounter, number.Float64Kind),
			test.Point(startTime, endTime, sum.NewNonMonotonicFloat64(1), cumulative), // AnySum -> Non-Monotonic
		),
		test.Instrument(
			test.Descriptor("SyncHistogram", sdkinstrument.SyncHistogram, number.Float64Kind),
			test.Point(startTime, endTime, sum.NewMonotonicFloat64(1), cumulative), // Histogram to Monotonic Sum
		),
		test.Instrument(
			test.Descriptor("AsyncGauge", sdkinstrument.AsyncGauge, number.Float64Kind),
			test.Point(startTime, endTime, gauge.NewFloat64(1), cumulative), // This stays a Gauge!
		),
	)
}

// TestDuplicateAsyncMeasurementsIngored tests that asynchronous
// instrument accumulators keep only the last observed value, while
// synchronous instruments correctly snapshotAndProcess them all.
func TestDuplicateAsyncMeasurementsIngored(t *testing.T) {
	vc := New(testLib, view.New("test"))

	inst1, err := testCompile(vc, "async", sdkinstrument.AsyncCounter, number.Float64Kind)
	require.NoError(t, err)

	inst2, err := testCompile(vc, "sync", sdkinstrument.SyncCounter, number.Float64Kind)
	require.NoError(t, err)

	for _, inst := range []Instrument{inst1, inst2} {
		acc := inst.NewAccumulator(attribute.NewSet())
		acc.(Updater[float64]).Update(1)
		acc.(Updater[float64]).Update(10)
		acc.(Updater[float64]).Update(100)
		acc.(Updater[float64]).Update(1000)
		acc.(Updater[float64]).Update(10000)
		acc.(Updater[float64]).Update(100000)
		acc.SnapshotAndProcess(false)
	}

	output := testCollect(t, vc)

	test.RequireEqualMetrics(t, output,
		test.Instrument(
			test.Descriptor("async", sdkinstrument.AsyncCounter, number.Float64Kind),
			test.Point(
				startTime, endTime, sum.NewMonotonicFloat64(100000), cumulative,
			),
		),
		test.Instrument(
			test.Descriptor("sync", sdkinstrument.SyncCounter, number.Float64Kind),
			test.Point(
				startTime, endTime, sum.NewMonotonicFloat64(111111), cumulative,
			),
		),
	)
}

// TestCumulativeTemporality ensures that synchronous instruments
// snapshotAndProcess data over time, whereas asynchronous instruments do not.
func TestCumulativeTemporality(t *testing.T) {
	views := view.New(
		"test",
		view.WithClause(
			// Dropping all keys
			view.WithKeys([]attribute.Key{}),
		),
		view.WithDefaultAggregationTemporalitySelector(view.StandardTemporality),
	)

	vc := New(testLib, views)

	inst1, err := testCompile(vc, "sync", sdkinstrument.SyncCounter, number.Float64Kind)
	require.NoError(t, err)

	inst2, err := testCompile(vc, "async", sdkinstrument.AsyncCounter, number.Float64Kind)
	require.NoError(t, err)

	setA := attribute.NewSet(attribute.String("A", "1"))
	setB := attribute.NewSet(attribute.String("B", "1"))

	for rounds := 1; rounds <= 2; rounds++ {
		for _, acc := range []Accumulator{
			inst1.NewAccumulator(setA),
			inst1.NewAccumulator(setB),
			inst2.NewAccumulator(setA),
			inst2.NewAccumulator(setB),
		} {
			acc.(Updater[float64]).Update(1)
			acc.SnapshotAndProcess(false)
		}

		test.RequireEqualMetrics(t, testCollect(t, vc),
			test.Instrument(
				test.Descriptor("sync", sdkinstrument.SyncCounter, number.Float64Kind),
				test.Point(
					// Because synchronous instruments snapshotAndProcess, the
					// rounds multiplier is used here but not in the case below.
					startTime, endTime, sum.NewMonotonicFloat64(float64(rounds)*2), cumulative,
				),
			),
			test.Instrument(
				test.Descriptor("async", sdkinstrument.AsyncCounter, number.Float64Kind),
				test.Point(
					startTime, endTime, sum.NewMonotonicFloat64(2), cumulative,
				),
			),
		)
	}
}

// TestDeltaTemporality ensures that synchronous instruments
// snapshotAndProcess data over time, whereas asynchronous instruments do not.
func TestDeltaTemporalityCounter(t *testing.T) {
	views := view.New(
		"test",
		view.WithClause(
			// Dropping all keys
			view.WithKeys([]attribute.Key{}),
		),
		view.WithDefaultAggregationTemporalitySelector(view.DeltaPreferredTemporality),
	)

	vc := New(testLib, views)

	inst1, err := testCompile(vc, "sync", sdkinstrument.SyncCounter, number.Float64Kind)
	require.NoError(t, err)

	inst2, err := testCompile(vc, "async", sdkinstrument.AsyncCounter, number.Float64Kind)
	require.NoError(t, err)

	setA := attribute.NewSet(attribute.String("A", "1"))
	setB := attribute.NewSet(attribute.String("B", "1"))

	seq := testSequence

	for rounds := 1; rounds <= 3; rounds++ {
		for _, acc := range []Accumulator{
			inst1.NewAccumulator(setA),
			inst1.NewAccumulator(setB),
			inst2.NewAccumulator(setA),
			inst2.NewAccumulator(setB),
		} {
			acc.(Updater[float64]).Update(float64(rounds))
			acc.SnapshotAndProcess(false)
		}

		test.RequireEqualMetrics(t, testCollectSequence(t, vc, seq),
			test.Instrument(
				test.Descriptor("sync", sdkinstrument.SyncCounter, number.Float64Kind),
				test.Point(
					// By construction, the change is rounds per attribute set == 2*rounds
					seq.Last, seq.Now, sum.NewMonotonicFloat64(2*float64(rounds)), delta,
				),
			),
			test.Instrument(
				test.Descriptor("async", sdkinstrument.AsyncCounter, number.Float64Kind),
				test.Point(
					// By construction, the change is 1 per attribute set == 2
					seq.Last, seq.Now, sum.NewMonotonicFloat64(2), delta,
				),
			),
		)

		// Update the test sequence
		seq.Last = seq.Now
		seq.Now = time.Now()
	}
}

// TestDeltaTemporalityGauge ensures that the asynchronous gauge
// when used with delta temporalty only reports changed values.
func TestDeltaTemporalityGauge(t *testing.T) {
	views := view.New(
		"test",
		view.WithDefaultAggregationTemporalitySelector(view.DeltaPreferredTemporality),
	)

	vc := New(testLib, views)

	instF, err := testCompile(vc, "gaugeF", sdkinstrument.AsyncGauge, number.Float64Kind)
	require.NoError(t, err)

	instI, err := testCompile(vc, "gaugeI", sdkinstrument.AsyncGauge, number.Int64Kind)
	require.NoError(t, err)

	set := attribute.NewSet()

	observe := func(x int) {
		accI := instI.NewAccumulator(set)
		accI.(Updater[int64]).Update(int64(x))
		accI.SnapshotAndProcess(false)

		accF := instF.NewAccumulator(set)
		accF.(Updater[float64]).Update(float64(x))
		accF.SnapshotAndProcess(false)
	}

	expectValues := func(x int, seq data.Sequence) {
		test.RequireEqualMetrics(t,
			testCollectSequence(t, vc, seq),
			test.Instrument(
				test.Descriptor("gaugeF", sdkinstrument.AsyncGauge, number.Float64Kind),
				test.Point(seq.Last, seq.Now, gauge.NewFloat64(float64(x)), delta),
			),
			test.Instrument(
				test.Descriptor("gaugeI", sdkinstrument.AsyncGauge, number.Int64Kind),
				test.Point(seq.Last, seq.Now, gauge.NewInt64(int64(x)), delta),
			),
		)
	}
	expectNone := func(seq data.Sequence) {
		test.RequireEqualMetrics(t,
			testCollectSequence(t, vc, seq),
			test.Instrument(
				test.Descriptor("gaugeF", sdkinstrument.AsyncGauge, number.Float64Kind),
			),
			test.Instrument(
				test.Descriptor("gaugeI", sdkinstrument.AsyncGauge, number.Int64Kind),
			),
		)
	}
	seq := testSequence
	tick := func() {
		// Update the test sequence
		seq.Last = seq.Now
		seq.Now = time.Now()
	}

	observe(10)
	expectValues(10, seq)
	tick()

	observe(10)
	expectNone(seq)
	tick()

	observe(10)
	expectNone(seq)
	tick()

	observe(11)
	expectValues(11, seq)
	tick()

	observe(11)
	expectNone(seq)
	tick()

	observe(10)
	expectValues(10, seq)
	tick()
}

// TestSyncDeltaTemporalityCounter ensures that counter and updowncounter
// skip points with delta temporality and no change.
func TestSyncDeltaTemporalityCounter(t *testing.T) {
	views := view.New(
		"test",
		view.WithDefaultAggregationTemporalitySelector(
			func(ik sdkinstrument.Kind) aggregation.Temporality {
				return aggregation.DeltaTemporality // Always delta
			}),
	)

	vc := New(testLib, views)

	instCF, err := testCompile(vc, "counterF", sdkinstrument.SyncCounter, number.Float64Kind)
	require.NoError(t, err)

	instCI, err := testCompile(vc, "counterI", sdkinstrument.SyncCounter, number.Int64Kind)
	require.NoError(t, err)

	instUF, err := testCompile(vc, "updowncounterF", sdkinstrument.SyncUpDownCounter, number.Float64Kind)
	require.NoError(t, err)

	instUI, err := testCompile(vc, "updowncounterI", sdkinstrument.SyncUpDownCounter, number.Int64Kind)
	require.NoError(t, err)

	set := attribute.NewSet()

	var output data.Scope

	observe := func(mono, nonMono int) {
		accCI := instCI.NewAccumulator(set)
		accCI.(Updater[int64]).Update(int64(mono))
		accCI.SnapshotAndProcess(false)

		accCF := instCF.NewAccumulator(set)
		accCF.(Updater[float64]).Update(float64(mono))
		accCF.SnapshotAndProcess(false)

		accUI := instUI.NewAccumulator(set)
		accUI.(Updater[int64]).Update(int64(nonMono))
		accUI.SnapshotAndProcess(false)

		accUF := instUF.NewAccumulator(set)
		accUF.(Updater[float64]).Update(float64(nonMono))
		accUF.SnapshotAndProcess(false)
	}

	expectValues := func(mono, nonMono int, seq data.Sequence) {
		test.RequireEqualMetrics(t,
			testCollectSequenceReuse(t, vc, seq, &output),
			test.Instrument(
				test.Descriptor("counterF", sdkinstrument.SyncCounter, number.Float64Kind),
				test.Point(seq.Last, seq.Now, sum.NewMonotonicFloat64(float64(mono)), delta),
			),
			test.Instrument(
				test.Descriptor("counterI", sdkinstrument.SyncCounter, number.Int64Kind),
				test.Point(seq.Last, seq.Now, sum.NewMonotonicInt64(int64(mono)), delta),
			),
			test.Instrument(
				test.Descriptor("updowncounterF", sdkinstrument.SyncUpDownCounter, number.Float64Kind),
				test.Point(seq.Last, seq.Now, sum.NewNonMonotonicFloat64(float64(nonMono)), delta),
			),
			test.Instrument(
				test.Descriptor("updowncounterI", sdkinstrument.SyncUpDownCounter, number.Int64Kind),
				test.Point(seq.Last, seq.Now, sum.NewNonMonotonicInt64(int64(nonMono)), delta),
			),
		)
	}
	expectNone := func(seq data.Sequence) {
		test.RequireEqualMetrics(t,
			testCollectSequenceReuse(t, vc, seq, &output),
			test.Instrument(
				test.Descriptor("counterF", sdkinstrument.SyncCounter, number.Float64Kind),
			),
			test.Instrument(
				test.Descriptor("counterI", sdkinstrument.SyncCounter, number.Int64Kind),
			),
			test.Instrument(
				test.Descriptor("updowncounterF", sdkinstrument.SyncUpDownCounter, number.Float64Kind),
			),
			test.Instrument(
				test.Descriptor("updowncounterI", sdkinstrument.SyncUpDownCounter, number.Int64Kind),
			),
		)
	}
	seq := testSequence
	tick := func() {
		// Update the test sequence
		seq.Last = seq.Now
		seq.Now = time.Now()
	}

	observe(10, 10)
	expectValues(10, 10, seq)
	tick()

	observe(0, 100)
	observe(0, -100)
	expectNone(seq)
	tick()

	observe(100, 100)
	expectValues(100, 100, seq)
	tick()
}

func TestSyncDeltaTemporalityMapDeletion(t *testing.T) {
	views := view.New(
		"test",
		view.WithDefaultAggregationTemporalitySelector(
			func(ik sdkinstrument.Kind) aggregation.Temporality {
				return aggregation.DeltaTemporality // Always delta
			}),
	)

	vc := New(testLib, views)

	inst, err := testCompile(vc, "counter", sdkinstrument.SyncCounter, number.Float64Kind)
	require.NoError(t, err)

	attr := attribute.String("A", "1")
	set := attribute.NewSet(attr)

	acc1 := inst.NewAccumulator(set)
	acc2 := inst.NewAccumulator(set)

	acc1.(Updater[float64]).Update(1)
	acc2.(Updater[float64]).Update(1)

	// There are two references to one entry in the map.
	require.Equal(t, 1, len(inst.(*statelessSyncInstrument[float64, sum.MonotonicFloat64, sum.MonotonicFloat64Methods]).data))

	acc1.SnapshotAndProcess(false)
	acc2.SnapshotAndProcess(true)

	var output data.Scope

	test.RequireEqualMetrics(t,
		testCollectSequenceReuse(t, vc, testSequence, &output),
		test.Instrument(
			test.Descriptor("counter", sdkinstrument.SyncCounter, number.Float64Kind),
			test.Point(middleTime, endTime, sum.NewMonotonicFloat64(2), delta, attr),
		),
	)

	require.Equal(t, 1, len(inst.(*statelessSyncInstrument[float64, sum.MonotonicFloat64, sum.MonotonicFloat64Methods]).data))

	acc1.SnapshotAndProcess(true)

	test.RequireEqualMetrics(t,
		testCollectSequenceReuse(t, vc, testSequence, &output),
		test.Instrument(
			test.Descriptor("counter", sdkinstrument.SyncCounter, number.Float64Kind),
		),
	)

	require.Equal(t, 0, len(inst.(*statelessSyncInstrument[float64, sum.MonotonicFloat64, sum.MonotonicFloat64Methods]).data))

}

func TestRegexpMatch(t *testing.T) {
	views := view.New(
		"test",
		view.WithClause(
			view.MatchInstrumentNameRegexp(regexp.MustCompile(".*_rate")),
			view.WithAggregation(aggregation.DropKind),
		),
	)

	vc := New(testLib, views)

	inst0, err := testCompile(vc, "foo_rate", sdkinstrument.AsyncGauge, number.Float64Kind)
	require.NoError(t, err)
	inst1, err := testCompile(vc, "bar_rate", sdkinstrument.AsyncGauge, number.Float64Kind)
	require.NoError(t, err)
	inst2, err := testCompile(vc, "notarate", sdkinstrument.AsyncGauge, number.Float64Kind)
	require.NoError(t, err)

	require.Nil(t, inst0)
	require.Nil(t, inst1)
	require.NotNil(t, inst2)
}

func TestSingleInstrumentWarning(t *testing.T) {
	views := view.New(
		"test",
		view.WithClause(
			view.MatchInstrumentNameRegexp(regexp.MustCompile(".*_rate")),
			view.WithName("fixed"),
		),
	)

	_, err := view.Validate(views)
	require.Error(t, err)
	require.Contains(t, err.Error(), "multi-instrument view specifies a single name")
}

func TestDeltaTemporalityMinMaxSumCount(t *testing.T) {
	views := view.New(
		"test",
		view.WithClause(
			view.MatchInstrumentKind(sdkinstrument.SyncHistogram),
			view.WithAggregation(aggregation.MinMaxSumCountKind),
		),
		view.WithDefaultAggregationTemporalitySelector(view.DeltaPreferredTemporality),
	)

	vc := New(testLib, views)

	inst1, err := testCompile(vc, "lowcost", sdkinstrument.SyncHistogram, number.Float64Kind)
	require.NoError(t, err)

	setA := attribute.NewSet(attribute.String("A", "1"))
	setB := attribute.NewSet(attribute.String("B", "1"))

	seq := testSequence

	const rounds = 10
	const expectCount uint64 = rounds
	const expectMax = 1.0
	const expectMin = 0x1p-9

	expectSum := 0.0
	expectMMSC := minmaxsumcount.NewFloat64()

	for round := 0; round < rounds; round++ {
		value := math.Exp2(float64(-round))
		expectSum += value
		minmaxsumcount.Float64Methods{}.Update(expectMMSC, value)
	}
	require.Equal(t, expectCount, expectMMSC.Count())
	require.Equal(t, expectSum, expectMMSC.Sum().CoerceToFloat64(number.Float64Kind))
	require.Equal(t, expectMin, expectMMSC.Min().CoerceToFloat64(number.Float64Kind))
	require.Equal(t, expectMax, expectMMSC.Max().CoerceToFloat64(number.Float64Kind))

	for _, acc := range []Accumulator{
		inst1.NewAccumulator(setA),
		inst1.NewAccumulator(setB),
	} {
		for round := 0; round < 10; round++ {
			acc.(Updater[float64]).Update(math.Exp2(float64(-round)))
			acc.SnapshotAndProcess(false)
		}
	}

	test.RequireEqualMetrics(t, testCollectSequence(t, vc, seq),
		test.Instrument(
			test.Descriptor("lowcost", sdkinstrument.SyncHistogram, number.Float64Kind),
			test.Point(seq.Last, seq.Now, expectMMSC, delta, setA.ToSlice()...),
			test.Point(seq.Last, seq.Now, expectMMSC, delta, setB.ToSlice()...),
		),
	)
}

func TestViewHints(t *testing.T) {
	views := view.New("test")
	vc := New(testLib, views)
	otelErrs := test.OTelErrors()

	histo, err := testCompile(
		vc,
		"histo",
		sdkinstrument.SyncCounter, // counter->small histogram
		number.Float64Kind,
		instrument.WithDescription(`{
  "aggregation": "histogram",
  "config": {
    "histogram": {
      "max_size": 3
    }
  }
}`))
	require.NoError(t, err)

	mmsc, err := testCompile(
		vc,
		"mmsc",
		sdkinstrument.SyncHistogram, // histogram->minmaxsumcount
		number.Float64Kind,
		instrument.WithDescription(`{
  "description": "heyyy",
  "aggregation": "minmaxsumcount"
}`))
	require.NoError(t, err)

	gg, err := testCompile(
		vc,
		"gauge",
		sdkinstrument.SyncUpDownCounter, // updowncounter->gauge
		number.Float64Kind,
		instrument.WithDescription(`{
  "description": "check it",
  "aggregation": "gauge"
}`))
	require.NoError(t, err)

	set := attribute.NewSet(attribute.String("test", "attr"))
	seq := testSequence
	inputs := []float64{1, 2, 3, 4, 5, 6, 7, 8}
	sum := 0.0
	numInputs := len(inputs)
	for _, inp := range inputs {
		sum += inp
	}

	for _, acc := range []Accumulator{
		histo.NewAccumulator(set),
		mmsc.NewAccumulator(set),
		gg.NewAccumulator(set),
	} {
		for _, inp := range inputs {
			acc.(Updater[float64]).Update(inp)
		}
		acc.SnapshotAndProcess(false)
	}

	test.RequireEqualMetrics(t, testCollectSequence(t, vc, seq),
		test.Instrument(
			test.Descriptor("histo", sdkinstrument.SyncCounter, number.Float64Kind),
			test.Point(seq.Start, seq.Now, histogram.NewFloat64(aggregator.HistogramConfig{
				MaxSize: 3,
			}, inputs...), cumulative, set.ToSlice()...),
		),
		test.Instrument(
			test.Descriptor("mmsc", sdkinstrument.SyncHistogram, number.Float64Kind, instrument.WithDescription("heyyy")),
			test.Point(seq.Start, seq.Now, minmaxsumcount.NewFloat64(inputs...), cumulative, set.ToSlice()...),
		),
		test.Instrument(
			test.Descriptor("gauge", sdkinstrument.SyncUpDownCounter, number.Float64Kind, instrument.WithDescription("check it")),
			test.Point(seq.Start, seq.Now, gauge.NewFloat64(inputs[numInputs-1]), cumulative, set.ToSlice()...),
		),
	)

	require.Nil(t, *otelErrs)
}

func TestViewHintErrors(t *testing.T) {
	views := view.New("test")
	vc := New(testLib, views)
	otelErrs := test.OTelErrors()

	_, err := testCompile(
		vc,
		"extra_comma",
		sdkinstrument.SyncCounter,
		number.Float64Kind,
		instrument.WithDescription(`{
  "aggregation": "histogram",
}`))
	require.NoError(t, err)

	_, err = testCompile(
		vc,
		"accidental_json_parse",
		sdkinstrument.SyncHistogram,
		number.Float64Kind,
		instrument.WithDescription("accidental { parse"))
	require.NoError(t, err)

	_, err = testCompile(
		vc,
		"invalid_aggregation",
		sdkinstrument.SyncUpDownCounter,
		number.Float64Kind,
		instrument.WithDescription(`{
  "aggregation": "cardinality"
}`))
	require.NoError(t, err)

	_, err = testCompile(
		vc,
		"bad_max_size",
		sdkinstrument.SyncCounter,
		number.Float64Kind,
		instrument.WithDescription(`{
  "aggregation": "histogram",
  "config": {
    "histogram": {
      "max_size": -3
    }
  }
}`))
	require.NoError(t, err)

	require.Equal(t, 4, len(*otelErrs))
	require.Contains(t, (*otelErrs)[0].Error(), "invalid character")
	require.Contains(t, (*otelErrs)[1].Error(), "looking for beginning")
	require.Contains(t, (*otelErrs)[2].Error(), "invalid aggregation")
	require.Contains(t, (*otelErrs)[3].Error(), "invalid aggregator config")
}
