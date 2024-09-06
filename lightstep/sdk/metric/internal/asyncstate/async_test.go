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

package asyncstate // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/asyncstate"

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/sum"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/pipeline"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/test"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

var (
	testLibrary = instrumentation.Scope{
		Name: "test",
	}

	endTime    = time.Unix(100, 0)
	middleTime = endTime.Add(-time.Millisecond)
	startTime  = endTime.Add(-2 * time.Millisecond)

	testSequence = data.Sequence{
		Start: startTime,
		Last:  middleTime,
		Now:   endTime,
	}

	ignorePerf = sdkinstrument.Performance{
		IgnoreCollisions:   false,
		AttributeSizeLimit: sdkinstrument.DefaultAttributeSizeLimit,
	}
)

type testSDK struct {
	compilers []*viewstate.Compiler
}

func (tsdk *testSDK) compile(desc sdkinstrument.Descriptor) pipeline.Register[viewstate.Instrument] {
	reg := pipeline.NewRegister[viewstate.Instrument](len(tsdk.compilers))

	for i, comp := range tsdk.compilers {
		inst, err := comp.Compile(desc)
		if err != nil {
			panic(err)
		}
		reg[i] = inst
	}
	return reg
}

func testAsync(name string, opts ...view.Option) *testSDK {
	return &testSDK{
		compilers: []*viewstate.Compiler{
			viewstate.New(testLibrary, view.New(name, ignorePerf, opts...)),
			viewstate.New(testLibrary, view.New(name, ignorePerf, opts...)),
		},
	}
}

func testAsync2(name string, opts1, opts2 []view.Option) *testSDK {
	return &testSDK{
		compilers: []*viewstate.Compiler{
			viewstate.New(testLibrary, view.New(name, ignorePerf, opts1...)),
			viewstate.New(testLibrary, view.New(name, ignorePerf, opts2...)),
		},
	}
}

func testState(num int) *State {
	return NewState(num)
}

type intObserver struct {
	*Observer
	metric.Int64Observable
}

type floatObserver struct {
	*Observer
	metric.Float64Observable
}

func testIntObserver(tsdk *testSDK, name string, ik sdkinstrument.Kind) intObserver {
	desc := test.Descriptor(name, ik, number.Int64Kind)
	return intObserver{Observer: New(desc, ignorePerf, tsdk, tsdk.compile(desc))}
}

func testFloatObserver(tsdk *testSDK, name string, ik sdkinstrument.Kind) floatObserver {
	desc := test.Descriptor(name, ik, number.Float64Kind)
	return floatObserver{Observer: New(desc, ignorePerf, tsdk, tsdk.compile(desc))}
}

func nopCB(context.Context, metric.Observer) error {
	return nil
}

func TestNewCallbackError(t *testing.T) {
	tsdk := testAsync("test")

	// no instruments error
	cb, err := NewCallback(nil, tsdk, nil)
	require.Error(t, err)
	require.Nil(t, cb)

	// nil callback error
	cntr := testIntObserver(tsdk, "counter", sdkinstrument.AsyncCounter)
	cb, err = NewCallback([]metric.Observable{cntr}, tsdk, nil)
	require.Error(t, err)
	require.Nil(t, cb)
}

func TestNewCallbackProviderMismatch(t *testing.T) {
	test0 := testAsync("test0")
	test1 := testAsync("test1")

	instA0 := testIntObserver(test0, "A", sdkinstrument.AsyncCounter)
	instB1 := testFloatObserver(test1, "A", sdkinstrument.AsyncCounter)

	cb, err := NewCallback([]metric.Observable{instA0, instB1}, test0, nopCB)
	require.Error(t, err)
	require.Contains(t, err.Error(), "asynchronous instrument belongs to a different meter")
	require.Nil(t, cb)

	cb, err = NewCallback([]metric.Observable{instA0, instB1}, test1, nopCB)
	require.Error(t, err)
	require.Contains(t, err.Error(), "asynchronous instrument belongs to a different meter")
	require.Nil(t, cb)

	cb, err = NewCallback([]metric.Observable{instA0}, test0, nopCB)
	require.NoError(t, err)
	require.NotNil(t, cb)

	cb, err = NewCallback([]metric.Observable{instB1}, test1, nopCB)
	require.NoError(t, err)
	require.NotNil(t, cb)

	// nil value not of this SDK
	var fake0 metric.Observable
	cb, err = NewCallback([]metric.Observable{fake0}, test0, nopCB)
	require.Error(t, err)
	require.Contains(t, err.Error(), "asynchronous instrument does not belong to this SDK")
	require.Nil(t, cb)

	// non-nil value not of this SDK
	var fake1 struct {
		metric.Observable
	}
	cb, err = NewCallback([]metric.Observable{fake1}, test0, nopCB)
	require.Error(t, err)
	require.Contains(t, err.Error(), "asynchronous instrument does not belong to this SDK")
	require.Nil(t, cb)
}

func TestCallbackInvalidation(t *testing.T) {
	errors := test.OTelErrors()

	tsdk := testAsync("test")

	var called int64
	var saveObs metric.Observer

	cntr := testIntObserver(tsdk, "counter", sdkinstrument.AsyncCounter)
	cb, err := NewCallback([]metric.Observable{cntr}, tsdk, func(ctx context.Context, obs metric.Observer) error {
		obs.ObserveInt64(cntr, called)
		saveObs = obs
		called++
		return nil
	})
	require.NoError(t, err)

	state := testState(0)

	// run the callback once legitimately
	cb.Run(context.Background(), state)

	// simulate use after callback return
	saveObs.ObserveInt64(cntr, 10000000)

	cntr.SnapshotAndProcess(state)

	require.Equal(t, int64(1), called)
	require.Equal(t, 1, len(*errors))
	require.Contains(t, (*errors)[0].Error(), "used after callback return")

	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			tsdk.compilers[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			cntr.descriptor,
			test.Point(startTime, endTime, sum.NewMonotonicInt64(0), aggregation.CumulativeTemporality),
		),
	)
}

func TestCallbackInstrumentUndeclaredForCalback(t *testing.T) {
	errors := test.OTelErrors()

	tt := testAsync("test")

	var called int64

	cntr1 := testIntObserver(tt, "counter1", sdkinstrument.AsyncCounter)
	cntr2 := testIntObserver(tt, "counter2", sdkinstrument.AsyncCounter)

	cb, err := NewCallback([]metric.Observable{cntr1}, tt, func(ctx context.Context, obs metric.Observer) error {
		obs.ObserveInt64(cntr2, called)
		called++
		return nil
	})
	require.NoError(t, err)

	state := testState(0)

	// run the callback once legitimately
	cb.Run(context.Background(), state)

	cntr1.SnapshotAndProcess(state)
	cntr2.SnapshotAndProcess(state)

	require.Equal(t, int64(1), called)
	require.Equal(t, 1, len(*errors))
	require.Contains(t, (*errors)[0].Error(), "instrument not declared for use in callback")

	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			tt.compilers[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			cntr1.descriptor,
		),
		test.Instrument(
			cntr2.descriptor,
		),
	)
}

func TestCallbackDisabledInstrument(t *testing.T) {
	tt := testAsync2(
		"test",
		[]view.Option{
			view.WithClause(
				view.MatchInstrumentName("drop1"),
				view.WithAggregation(aggregation.DropKind),
			),
			view.WithClause(
				view.MatchInstrumentName("drop2"),
				view.WithAggregation(aggregation.DropKind),
			),
		},
		[]view.Option{
			view.WithClause(
				view.MatchInstrumentName("drop2"),
				view.WithAggregation(aggregation.DropKind),
			),
		},
	)

	cntrDrop1 := testFloatObserver(tt, "drop1", sdkinstrument.AsyncCounter)
	cntrDrop2 := testFloatObserver(tt, "drop2", sdkinstrument.AsyncCounter)
	cntrKeep := testFloatObserver(tt, "keep", sdkinstrument.AsyncCounter)

	cb, _ := NewCallback([]metric.Observable{cntrDrop1, cntrDrop2, cntrKeep}, tt, func(ctx context.Context, obs metric.Observer) error {
		obs.ObserveFloat64(cntrKeep, 1000)
		obs.ObserveFloat64(cntrDrop1, 1001)
		obs.ObserveFloat64(cntrDrop2, 1002)
		return nil
	})

	runFor := func(num int) {
		state := testState(num)

		cb.Run(context.Background(), state)

		cntrKeep.SnapshotAndProcess(state)
		cntrDrop1.SnapshotAndProcess(state)
		cntrDrop2.SnapshotAndProcess(state)
	}

	runFor(0)
	runFor(1)

	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			tt.compilers[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			cntrKeep.descriptor,
			test.Point(startTime, endTime, sum.NewMonotonicFloat64(1000), aggregation.CumulativeTemporality),
		),
	)
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			tt.compilers[1].Collectors(),
			testSequence,
		),
		test.Instrument(
			cntrDrop1.descriptor,
			test.Point(startTime, endTime, sum.NewMonotonicFloat64(1001), aggregation.CumulativeTemporality),
		),
		test.Instrument(
			cntrKeep.descriptor,
			test.Point(startTime, endTime, sum.NewMonotonicFloat64(1000), aggregation.CumulativeTemporality),
		),
	)
}

func TestOutOfRangeValues(t *testing.T) {
	otelErrs := test.OTelErrors()

	tt := testAsync("test")

	c := testFloatObserver(tt, "testPatternC", sdkinstrument.AsyncCounter)
	u := testFloatObserver(tt, "testPatternU", sdkinstrument.AsyncUpDownCounter)
	g := testFloatObserver(tt, "testPatternG", sdkinstrument.AsyncGauge)

	cb, _ := NewCallback([]metric.Observable{
		c, u, g,
	}, tt, func(ctx context.Context, obs metric.Observer) error {
		obs.ObserveFloat64(c, math.NaN())
		obs.ObserveFloat64(c, math.Inf(+1))
		obs.ObserveFloat64(u, math.NaN())
		obs.ObserveFloat64(u, math.Inf(+1))
		obs.ObserveFloat64(g, math.NaN())
		obs.ObserveFloat64(g, math.Inf(+1))
		return nil
	})

	runFor := func(num int) {
		state := testState(num)

		cb.Run(context.Background(), state)

		c.SnapshotAndProcess(state)
		u.SnapshotAndProcess(state)
		g.SnapshotAndProcess(state)
	}

	for i := 0; i < 2; i++ {
		runFor(i)

		test.RequireEqualMetrics(
			t,
			test.CollectScope(
				t,
				tt.compilers[i].Collectors(),
				testSequence,
			),
			test.Instrument(
				c.descriptor,
			),
			test.Instrument(
				u.descriptor,
			),
			test.Instrument(
				g.descriptor,
			),
		)
	}

	// Errors are rate limited, but this is the only test in this
	// package that uses invalid values.  We should have at least
	// one per class.
	require.LessOrEqual(t, 2, len(*otelErrs))

	haveNaN := false
	haveInf := false
	for _, err := range *otelErrs {
		isNaN := errors.Is(err, aggregator.ErrNaNInput)
		isInf := errors.Is(err, aggregator.ErrInfInput)

		require.True(t, isNaN || isInf)
		require.True(t, strings.HasPrefix(err.Error(), "testPattern"))

		haveNaN = haveNaN || isNaN
		haveInf = haveInf || isInf
	}
	require.True(t, haveNaN)
	require.True(t, haveInf)
}

func TestAttributeSizeLimit(t *testing.T) {
	tsdk := testAsync("test")
	cntr := testIntObserver(tsdk, "counter", sdkinstrument.AsyncCounter)
	cb, err := NewCallback([]metric.Observable{cntr}, tsdk, func(ctx context.Context, obs metric.Observer) error {
		obs.ObserveInt64(cntr, 1234,
			metric.WithAttributes(attribute.String(strings.Repeat("X", 1<<14), strings.Repeat("Y", 1<<14))),
		)
		return nil
	})
	require.NoError(t, err)

	state := testState(0)

	// run the callback once legitimately
	cb.Run(context.Background(), state)

	cntr.SnapshotAndProcess(state)

	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			tsdk.compilers[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			cntr.descriptor,
			test.Point(startTime, endTime, sum.NewMonotonicInt64(1234), aggregation.CumulativeTemporality,
				attribute.String(
					strings.Repeat("X", sdkinstrument.DefaultAttributeSizeLimit),
					strings.Repeat("Y", sdkinstrument.DefaultAttributeSizeLimit),
				),
			),
		),
	)
}
