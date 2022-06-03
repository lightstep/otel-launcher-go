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
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/sum"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/pipeline"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/test"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

var (
	testLibrary = instrumentation.Library{
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
			viewstate.New(testLibrary, view.New(name, opts...)),
			viewstate.New(testLibrary, view.New(name, opts...)),
		},
	}
}

func testAsync2(name string, opts1, opts2 []view.Option) *testSDK {
	return &testSDK{
		compilers: []*viewstate.Compiler{
			viewstate.New(testLibrary, view.New(name, opts1...)),
			viewstate.New(testLibrary, view.New(name, opts2...)),
		},
	}
}

func testState(num int) *State {
	return NewState(num)
}

func testObserver[N number.Any, Traits number.Traits[N]](tsdk *testSDK, name string, ik sdkinstrument.Kind, opts ...instrument.Option) Observer[N, Traits] {
	var t Traits
	desc := test.Descriptor(name, ik, t.Kind(), opts...)
	impl := NewInstrument(desc, tsdk, tsdk.compile(desc))
	return NewObserver[N, Traits](impl)
}

func TestNewCallbackError(t *testing.T) {
	tsdk := testAsync("test")

	// no instruments error
	cb, err := NewCallback(nil, tsdk, nil)
	require.Error(t, err)
	require.Nil(t, cb)

	// nil callback error
	cntr := testObserver[int64, number.Int64Traits](tsdk, "counter", sdkinstrument.CounterObserverKind)
	cb, err = NewCallback([]instrument.Asynchronous{cntr}, tsdk, nil)
	require.Error(t, err)
	require.Nil(t, cb)
}

func TestNewCallbackProviderMismatch(t *testing.T) {
	test0 := testAsync("test0")
	test1 := testAsync("test1")

	instA0 := testObserver[int64, number.Int64Traits](test0, "A", sdkinstrument.CounterObserverKind)
	instB1 := testObserver[float64, number.Float64Traits](test1, "A", sdkinstrument.CounterObserverKind)

	cb, err := NewCallback([]instrument.Asynchronous{instA0, instB1}, test0, func(context.Context) {})
	require.Error(t, err)
	require.Contains(t, err.Error(), "asynchronous instrument belongs to a different meter")
	require.Nil(t, cb)

	cb, err = NewCallback([]instrument.Asynchronous{instA0, instB1}, test1, func(context.Context) {})
	require.Error(t, err)
	require.Contains(t, err.Error(), "asynchronous instrument belongs to a different meter")
	require.Nil(t, cb)

	cb, err = NewCallback([]instrument.Asynchronous{instA0}, test0, func(context.Context) {})
	require.NoError(t, err)
	require.NotNil(t, cb)

	cb, err = NewCallback([]instrument.Asynchronous{instB1}, test1, func(context.Context) {})
	require.NoError(t, err)
	require.NotNil(t, cb)

	// nil value not of this SDK
	var fake0 instrument.Asynchronous
	cb, err = NewCallback([]instrument.Asynchronous{fake0}, test0, func(context.Context) {})
	require.Error(t, err)
	require.Contains(t, err.Error(), "asynchronous instrument does not belong to this SDK")
	require.Nil(t, cb)

	// non-nil value not of this SDK
	var fake1 struct {
		instrument.Asynchronous
	}
	cb, err = NewCallback([]instrument.Asynchronous{fake1}, test0, func(context.Context) {})
	require.Error(t, err)
	require.Contains(t, err.Error(), "asynchronous instrument does not belong to this SDK")
	require.Nil(t, cb)
}

func TestCallbackInvalidation(t *testing.T) {
	errors := test.OTelErrors()

	tsdk := testAsync("test")

	var called int64
	var saveCtx context.Context

	cntr := testObserver[int64, number.Int64Traits](tsdk, "counter", sdkinstrument.CounterObserverKind)
	cb, err := NewCallback([]instrument.Asynchronous{cntr}, tsdk, func(ctx context.Context) {
		cntr.Observe(ctx, called)
		saveCtx = ctx
		called++
	})
	require.NoError(t, err)

	state := testState(0)

	// run the callback once legitimately
	cb.Run(context.Background(), state)

	// simulate use after callback return
	cntr.Observe(saveCtx, 10000000)

	cntr.inst.SnapshotAndProcess(state)

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
			cntr.inst.descriptor,
			test.Point(startTime, endTime, sum.NewMonotonicInt64(0), aggregation.CumulativeTemporality),
		),
	)
}

func TestCallbackInstrumentUndeclaredForCalback(t *testing.T) {
	errors := test.OTelErrors()

	tt := testAsync("test")

	var called int64

	cntr1 := testObserver[int64, number.Int64Traits](tt, "counter1", sdkinstrument.CounterObserverKind)
	cntr2 := testObserver[int64, number.Int64Traits](tt, "counter2", sdkinstrument.CounterObserverKind)

	cb, err := NewCallback([]instrument.Asynchronous{cntr1}, tt, func(ctx context.Context) {
		cntr2.Observe(ctx, called)
		called++
	})
	require.NoError(t, err)

	state := testState(0)

	// run the callback once legitimately
	cb.Run(context.Background(), state)

	cntr1.inst.SnapshotAndProcess(state)
	cntr2.inst.SnapshotAndProcess(state)

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
			cntr1.inst.descriptor,
		),
		test.Instrument(
			cntr2.inst.descriptor,
		),
	)
}

func TestInstrumentUseOutsideCallback(t *testing.T) {
	errors := test.OTelErrors()

	tt := testAsync("test")

	cntr := testObserver[float64, number.Float64Traits](tt, "cntr", sdkinstrument.CounterObserverKind)

	cntr.Observe(context.Background(), 1000)

	state := testState(0)

	cntr.inst.SnapshotAndProcess(state)

	require.Equal(t, 1, len(*errors))
	require.Contains(t, (*errors)[0].Error(), "async instrument used outside of callback")

	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			tt.compilers[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			cntr.inst.descriptor,
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

	cntrDrop1 := testObserver[float64, number.Float64Traits](tt, "drop1", sdkinstrument.CounterObserverKind)
	cntrDrop2 := testObserver[float64, number.Float64Traits](tt, "drop2", sdkinstrument.CounterObserverKind)
	cntrKeep := testObserver[float64, number.Float64Traits](tt, "keep", sdkinstrument.CounterObserverKind)

	cb, _ := NewCallback([]instrument.Asynchronous{cntrDrop1, cntrDrop2, cntrKeep}, tt, func(ctx context.Context) {
		cntrKeep.Observe(ctx, 1000)
		cntrDrop1.Observe(ctx, 1001)
		cntrDrop2.Observe(ctx, 1002)
	})

	runFor := func(num int) {
		state := testState(num)

		cb.Run(context.Background(), state)

		cntrKeep.inst.SnapshotAndProcess(state)
		cntrDrop1.inst.SnapshotAndProcess(state)
		cntrDrop2.inst.SnapshotAndProcess(state)
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
			cntrKeep.inst.descriptor,
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
			cntrDrop1.inst.descriptor,
			test.Point(startTime, endTime, sum.NewMonotonicFloat64(1001), aggregation.CumulativeTemporality),
		),
		test.Instrument(
			cntrKeep.inst.descriptor,
			test.Point(startTime, endTime, sum.NewMonotonicFloat64(1000), aggregation.CumulativeTemporality),
		),
	)
}
