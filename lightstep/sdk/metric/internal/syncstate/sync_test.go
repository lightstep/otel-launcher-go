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

package syncstate

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	histostruct "github.com/lightstep/go-expohisto/structure"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/gauge"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/sum"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/pipeline"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/test"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/trace"
)

var (
	endTime    = time.Unix(100, 0)
	middleTime = time.Unix(99, 0)
	startTime  = time.Unix(98, 0)

	testSequence = data.Sequence{
		Start: startTime,
		Last:  middleTime,
		Now:   endTime,
	}

	testKeyVal = attribute.String("test", "test-value")

	safePerf = sdkinstrument.Performance{
		IgnoreCollisions:          false,
		InactiveCollectionPeriods: 1,
		MeasurementProcessor: &testMeasurementProcessor{
			attrs: []attribute.KeyValue{
				testKeyVal,
			},
		},
	}

	unsafePerf = sdkinstrument.Performance{
		IgnoreCollisions:          true,
		InactiveCollectionPeriods: 1,
	}

	noAttrsCfg = attrsConfig()
)

func attrsConfig(attrs ...attribute.KeyValue) OpConfig {
	acfg := metric.NewAddConfig([]metric.AddOption{
		metric.WithAttributes(attrs...),
	})
	return OpConfig{
		Attributes: acfg.Attributes(),
	}
}

func deltaUpdate[N number.Any](old, new N) N {
	return old + new
}

func cumulativeUpdate[N number.Any](_, new N) N {
	return new
}

const testAttr = attribute.Key("key")

var (
	deltaSelector = view.WithDefaultAggregationTemporalitySelector(func(_ sdkinstrument.Kind) aggregation.Temporality {
		return aggregation.DeltaTemporality
	})

	cumulativeSelector = view.WithDefaultAggregationTemporalitySelector(func(_ sdkinstrument.Kind) aggregation.Temporality {
		return aggregation.CumulativeTemporality
	})

	keyFilter = view.WithClause(
		view.WithKeys([]attribute.Key{}),
	)
)

type testMeasurementProcessor struct {
	attrs []attribute.KeyValue
}

func (mp *testMeasurementProcessor) Process(ctx context.Context, inAttrs []attribute.KeyValue) []attribute.KeyValue {
	outAttrs := make([]attribute.KeyValue, 0, len(mp.attrs)+len(inAttrs))
	outAttrs = append(outAttrs, inAttrs...)
	return append(outAttrs, mp.attrs...)
}

func TestSyncStateDeltaConcurrencyInt(t *testing.T) {
	testSyncStateConcurrency[int64, number.Int64Traits](t, deltaUpdate[int64], deltaSelector)
}

func TestSyncStateCumulativeConcurrencyInt(t *testing.T) {
	testSyncStateConcurrency[int64, number.Int64Traits](t, cumulativeUpdate[int64], cumulativeSelector)
}

func TestSyncStateCumulativeConcurrencyIntFiltered(t *testing.T) {
	testSyncStateConcurrency[int64, number.Int64Traits](t, cumulativeUpdate[int64], cumulativeSelector, keyFilter)
}

func TestSyncStateDeltaConcurrencyFloat(t *testing.T) {
	testSyncStateConcurrency[float64, number.Float64Traits](t, deltaUpdate[float64], deltaSelector)
}

func TestSyncStateCumulativeConcurrencyFloat(t *testing.T) {
	testSyncStateConcurrency[float64, number.Float64Traits](t, cumulativeUpdate[float64], cumulativeSelector)
}

func TestSyncStateCumulativeConcurrencyFloatFiltered(t *testing.T) {
	testSyncStateConcurrency[float64, number.Float64Traits](t, cumulativeUpdate[float64], cumulativeSelector, keyFilter)
}

func testSyncStateConcurrency[N number.Any, Traits number.Traits[N]](t *testing.T, update func(old, new N) N, vopts ...view.Option) {
	t.Run("unsafe_collisions", func(t *testing.T) { testSyncStateConcurrencyWithPerf[N, Traits](t, unsafePerf, update, vopts...) })
	t.Run("safe_collisions", func(t *testing.T) { testSyncStateConcurrencyWithPerf[N, Traits](t, safePerf, update, vopts...) })
}

func testSyncStateConcurrencyWithPerf[N number.Any, Traits number.Traits[N]](t *testing.T, perf sdkinstrument.Performance, update func(old, new N) N, vopts ...view.Option) {
	// Note: prior to
	// https://github.com/lightstep/otel-launcher-go/pull/206 this
	// code was able to reproduce the race condition handled in
	// this code.  The race condition still exists, but the test
	// no longer covers the call to Gosched() in acquireRecord()
	// or the return-nil branch in acquireWrite().  This is
	// because the current code is much less racey or better
	// tests are needed.
	const (
		numReaders  = 2
		numRoutines = 10
		numAttrs    = 10
		numUpdates  = 1e5
	)

	var traits Traits
	var writers sync.WaitGroup
	var readers sync.WaitGroup

	readers.Add(numReaders)
	writers.Add(numRoutines)

	lib := instrumentation.Scope{
		Name: "testlib",
	}
	vcs := make([]*viewstate.Compiler, numReaders)
	for vci := range vcs {
		vcs[vci] = viewstate.New(lib, view.New("test", safePerf, vopts...))
	}
	attrs := make([]attribute.KeyValue, numAttrs)
	for i := range attrs {
		attrs[i] = testAttr.Int(i)
	}

	desc := test.Descriptor("tester", sdkinstrument.SyncCounter, traits.Kind())

	pipes := make(pipeline.Register[viewstate.Instrument], numReaders)
	for vci := range vcs {
		pipes[vci], _ = vcs[vci].Compile(desc)
	}

	inst := New(desc, perf, nil, pipes)
	require.NotNil(t, inst)

	ctx, cancel := context.WithCancel(context.Background())

	partialCounts := make([]map[attribute.Set]N, numReaders)

	for vci := range vcs {
		partialCounts[vci] = map[attribute.Set]N{}
	}

	// Reader loops
	for vci := range vcs {
		go func(vci int, partial map[attribute.Set]N, vc *viewstate.Compiler) {
			defer readers.Done()

			// scope will be reused by this reader
			var scope data.Scope
			seq := data.Sequence{
				Start: time.Now(),
			}
			seq.Now = seq.Start

			collect := func() {
				seq.Last = seq.Now
				seq.Now = time.Now()

				inst.SnapshotAndProcess()

				scope.Reset()

				vc.Collectors()[0].Collect(seq, &scope.Instruments)

				for _, pt := range scope.Instruments[0].Points {
					partial[pt.Attributes] = update(partial[pt.Attributes], traits.FromNumber(pt.Aggregation.(*sum.State[N, Traits, sum.Monotonic]).Sum()))
				}
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
		}(vci, partialCounts[vci], vcs[vci])
	}

	// Writer loops
	for i := 0; i < numRoutines; i++ {
		go func() {
			defer writers.Done()
			rnd := rand.New(rand.NewSource(rand.Int63()))

			cfg := attrsConfig(attrs[rnd.Intn(len(attrs))])

			for j := 0; j < numUpdates/numRoutines; j++ {
				Observe[N, Traits](ctx, inst, 1, cfg)
			}
		}()
	}

	writers.Wait()
	cancel()
	readers.Wait()

	for vci := range vcs {
		var sum N
		for _, count := range partialCounts[vci] {
			sum += count
		}
		require.Equal(t, N(numUpdates), sum, "vci==%d", vci)
	}
}

func TestSyncStatePartialNoopInstrument(t *testing.T) {
	ctx := context.Background()
	vopts := []view.Option{
		view.WithClause(
			view.MatchInstrumentName("dropme"),
			view.WithAggregation(aggregation.DropKind),
		),
	}
	lib := instrumentation.Scope{
		Name: "testlib",
	}
	vcs := make([]*viewstate.Compiler, 2)
	vcs[0] = viewstate.New(lib, view.New("dropper", safePerf, vopts...))
	vcs[1] = viewstate.New(lib, view.New("keeper", safePerf))

	desc := test.Descriptor("dropme", sdkinstrument.SyncHistogram, number.Float64Kind)

	pipes := make(pipeline.Register[viewstate.Instrument], 2)
	pipes[0], _ = vcs[0].Compile(desc)
	pipes[1], _ = vcs[1].Compile(desc)

	require.Nil(t, pipes[0])
	require.NotNil(t, pipes[1])

	inst := New(desc, safePerf, nil, pipes)
	require.NotNil(t, inst)

	inst.ObserveFloat64(ctx, 1, noAttrsCfg)
	inst.ObserveFloat64(ctx, 2, noAttrsCfg)
	inst.ObserveFloat64(ctx, 3, noAttrsCfg)

	inst.SnapshotAndProcess()

	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
	)

	// Note: Create a merged histogram that is exactly equal to
	// the one we expect.  Merging creates a slightly different
	// struct, despite identical value, so we merge to create the
	// expected value:
	expectHist := histogram.NewFloat64(histogram.NewConfig())
	mergeIn := histogram.NewFloat64(histogram.NewConfig(), 1, 2, 3)
	var methods histogram.Float64Methods
	methods.Copy(mergeIn, expectHist)

	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[1].Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
			test.Point(startTime, endTime,
				expectHist,
				aggregation.CumulativeTemporality,
				testKeyVal,
			),
		),
	)
}

func TestSyncStateFullNoopInstrument(t *testing.T) {
	ctx := context.Background()
	vopts := []view.Option{
		view.WithClause(
			view.MatchInstrumentName("dropme"),
			view.WithAggregation(aggregation.DropKind),
		),
	}
	lib := instrumentation.Scope{
		Name: "testlib",
	}
	vcs := make([]*viewstate.Compiler, 2)
	vcs[0] = viewstate.New(lib, view.New("dropper", safePerf, vopts...))
	vcs[1] = viewstate.New(lib, view.New("keeper", safePerf, vopts...))

	desc := test.Descriptor("dropme", sdkinstrument.SyncHistogram, number.Float64Kind)

	pipes := make(pipeline.Register[viewstate.Instrument], 2)
	pipes[0], _ = vcs[0].Compile(desc)
	pipes[1], _ = vcs[1].Compile(desc)

	require.Nil(t, pipes[0])
	require.Nil(t, pipes[1])

	inst := New(desc, safePerf, nil, pipes)
	require.Nil(t, inst)

	inst.ObserveFloat64(ctx, 1, noAttrsCfg)
	inst.ObserveFloat64(ctx, 2, noAttrsCfg)
	inst.ObserveFloat64(ctx, 3, noAttrsCfg)

	// There's no instrument, nothing to Snapshot
	require.Equal(t, 0, len(vcs[0].Collectors()))
	require.Equal(t, 0, len(vcs[1].Collectors()))
}

func TestOutOfRangeValues(t *testing.T) {
	otelErrs := test.OTelErrors()

	for _, desc := range []sdkinstrument.Descriptor{
		test.Descriptor("cf", sdkinstrument.SyncCounter, number.Float64Kind),
		test.Descriptor("uf", sdkinstrument.SyncUpDownCounter, number.Float64Kind),
		test.Descriptor("hf", sdkinstrument.SyncHistogram, number.Float64Kind),
		test.Descriptor("ci", sdkinstrument.SyncCounter, number.Int64Kind),
		test.Descriptor("ui", sdkinstrument.SyncUpDownCounter, number.Int64Kind),
		test.Descriptor("hi", sdkinstrument.SyncHistogram, number.Int64Kind),
	} {
		ctx := context.Background()
		lib := instrumentation.Scope{
			Name: "testlib",
		}
		vcs := make([]*viewstate.Compiler, 1)
		vcs[0] = viewstate.New(lib, view.New("test", safePerf))

		pipes := make(pipeline.Register[viewstate.Instrument], 1)
		pipes[0], _ = vcs[0].Compile(desc)

		inst := New(desc, safePerf, nil, pipes)
		require.NotNil(t, inst)

		var negOne aggregation.Aggregation

		if desc.NumberKind == number.Float64Kind {
			inst.ObserveFloat64(ctx, -1, noAttrsCfg)
			inst.ObserveFloat64(ctx, math.NaN(), noAttrsCfg)
			inst.ObserveFloat64(ctx, math.Inf(+1), noAttrsCfg)
			inst.ObserveFloat64(ctx, math.Inf(-1), noAttrsCfg)
			negOne = sum.NewNonMonotonicFloat64(-1)
		} else {
			inst.ObserveInt64(ctx, -1, noAttrsCfg)
			negOne = sum.NewNonMonotonicInt64(-1)
		}

		inst.SnapshotAndProcess()

		var expectPoints []data.Point

		if desc.Kind == sdkinstrument.SyncUpDownCounter {
			expectPoints = append(expectPoints, test.Point(
				startTime, endTime,
				negOne,
				aggregation.CumulativeTemporality,
				testKeyVal,
			))
		}

		test.RequireEqualMetrics(
			t,
			test.CollectScope(
				t,
				vcs[0].Collectors(),
				testSequence,
			),
			test.Instrument(
				desc,
				expectPoints...,
			),
		)
	}

	// Errors are rate limited, but this is the only test in this
	// package that uses invalid values.  We should have at least
	// one per class.
	require.LessOrEqual(t, 3, len(*otelErrs))

	haveNaN := false
	haveInf := false
	haveNeg := false
	for _, err := range *otelErrs {
		isNaN := errors.Is(err, aggregator.ErrNaNInput)
		isInf := errors.Is(err, aggregator.ErrInfInput)
		isNeg := errors.Is(err, aggregator.ErrNegativeInput)

		require.True(t, isNaN || isInf || isNeg)

		haveNaN = haveNaN || isNaN
		haveInf = haveInf || isInf
		haveNeg = haveNeg || isNeg
	}
	require.True(t, haveNaN)
	require.True(t, haveInf)
	require.True(t, haveNeg)
}

func TestSyncGaugeDeltaInstrument(t *testing.T) {
	ctx := context.Background()
	lib := instrumentation.Scope{
		Name: "testlib",
	}
	vcs := make([]*viewstate.Compiler, 2)
	vcs[0] = viewstate.New(lib, view.New(
		"test",
		safePerf,
		deltaSelector,
		view.WithClause(
			view.WithKeys([]attribute.Key{"A", "C"}),
		),
	))

	indesc := test.Descriptor(
		"syncgauge",
		sdkinstrument.SyncUpDownCounter,
		number.Float64Kind)
	indesc.Description = `{
  "aggregation": "gauge",
  "description": "incredible"
}`

	outdesc := test.Descriptor(
		"syncgauge",
		sdkinstrument.SyncUpDownCounter,
		number.Float64Kind)
	outdesc.Description = "incredible"

	pipes := make(pipeline.Register[viewstate.Instrument], 1)
	pipes[0], _ = vcs[0].Compile(indesc)

	require.NotNil(t, pipes[0])

	inst := New(indesc, safePerf, nil, pipes)
	require.NotNil(t, inst)

	inst.ObserveFloat64(ctx, 1, noAttrsCfg)
	inst.ObserveFloat64(ctx, 2, noAttrsCfg)
	inst.ObserveFloat64(ctx, 3, noAttrsCfg)

	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			outdesc,
			test.Point(middleTime, endTime,
				gauge.NewFloat64(3),
				aggregation.DeltaTemporality,
			),
		),
	)

	// If not set, it disappears.
	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			outdesc,
		),
	)

	// Set again
	inst.ObserveFloat64(ctx, 172, noAttrsCfg)
	inst.ObserveFloat64(ctx, 175, noAttrsCfg)

	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			outdesc,
			test.Point(middleTime, endTime,
				gauge.NewFloat64(175),
				aggregation.DeltaTemporality,
			),
		),
	)

	// Set different attribute sets, leave the first (empty set) unused.
	inst.ObserveFloat64(ctx, 1333, attrsConfig(attribute.String("A", "B")))
	inst.ObserveFloat64(ctx, 1337, attrsConfig(attribute.String("C", "D")))

	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			outdesc,
			test.Point(middleTime, endTime,
				gauge.NewFloat64(1333),
				aggregation.DeltaTemporality,
				attribute.String("A", "B"),
			),
			test.Point(middleTime, endTime,
				gauge.NewFloat64(1337),
				aggregation.DeltaTemporality,
				attribute.String("C", "D"),
			),
		),
	)

	// Test the filters.  Last value should win due to the Gauge
	// sequence number (as opposed to random choice, which would
	// happen naturally b/c of map iteration).
	for i := 0; i < 1000; i++ {
		inst.ObserveFloat64(ctx, float64(i), attrsConfig(attribute.Int("ignored", i), attribute.String("A", "B")))
	}
	for i := 1000; i > 0; i-- {
		inst.ObserveFloat64(ctx, float64(i), attrsConfig(attribute.Int("ignored", i), attribute.String("C", "D")))
	}

	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			outdesc,
			test.Point(middleTime, endTime,
				gauge.NewFloat64(999),
				aggregation.DeltaTemporality,
				attribute.String("A", "B"),
			),
			test.Point(middleTime, endTime,
				gauge.NewFloat64(1),
				aggregation.DeltaTemporality,
				attribute.String("C", "D"),
			),
		),
	)
}

func TestFingerprinting(t *testing.T) {
	// Coverage
	require.NotEqual(
		t,
		fingerprintAttributes([]attribute.KeyValue{attribute.Bool("B", true)}),
		fingerprintAttributes([]attribute.KeyValue{attribute.Bool("B", false)}),
	)
	require.NotEqual(
		t,
		fingerprintAttributes([]attribute.KeyValue{attribute.Float64("F", 1.0)}),
		fingerprintAttributes([]attribute.KeyValue{attribute.Float64("F", 2.0)}),
	)
	require.NotEqual(
		t,
		fingerprintAttributes([]attribute.KeyValue{attribute.BoolSlice("BS", []bool{true, false})}),
		fingerprintAttributes([]attribute.KeyValue{attribute.BoolSlice("BS", []bool{true, true})}),
	)
	require.NotEqual(
		t,
		fingerprintAttributes([]attribute.KeyValue{attribute.Float64Slice("FS", []float64{1, 2})}),
		fingerprintAttributes([]attribute.KeyValue{attribute.Float64Slice("FS", []float64{1, 4})}),
	)
	require.NotEqual(
		t,
		fingerprintAttributes([]attribute.KeyValue{attribute.StringSlice("SS", []string{"a", "b"})}),
		fingerprintAttributes([]attribute.KeyValue{attribute.StringSlice("SS", []string{"a", "c"})}),
	)
	require.NotEqual(
		t,
		fingerprintAttributes([]attribute.KeyValue{attribute.Int64Slice("IS", []int64{10, 11})}),
		fingerprintAttributes([]attribute.KeyValue{attribute.Int64Slice("IS", []int64{10, 12})}),
	)
	// Empty
	require.Equal(
		t,
		uint64(0),
		fingerprintAttributes([]attribute.KeyValue{}),
	)
	// Uninitialized key/value
	require.NotEqual(
		t,
		uint64(0),
		fingerprintAttributes([]attribute.KeyValue{{}}),
	)
	// Two uninitialized
	require.Equal(
		t,
		fingerprintAttributes([]attribute.KeyValue{{}, {}}),
		fingerprintAttributes([]attribute.KeyValue{{}, {}}),
	)

}

func TestAttributesEqual(t *testing.T) {
	// Note: these code paths are difficult to test indirectly, because
	// of the difficult of finding hash collisions.  There is a single
	// test case for the collision path, the other forms of attributesEqual
	// are tested directly.
	var (
		kv = func(k attribute.Key, v attribute.Value) attribute.KeyValue {
			return attribute.KeyValue{Key: k, Value: v}
		}

		attrs = func(as ...attribute.KeyValue) []attribute.KeyValue {
			return as
		}
		values = []attribute.Value{
			attribute.Float64Value(1),
			attribute.Float64Value(2),
			attribute.Int64Value(1),
			attribute.Int64Value(2),
			attribute.BoolValue(true),
			attribute.BoolValue(false),
			attribute.StringValue("s"),
			attribute.StringValue("t"),
			attribute.Float64SliceValue([]float64{1, 2, 3}),
			attribute.Float64SliceValue([]float64{1, 2}),
			attribute.Float64SliceValue([]float64{1, 3}),
			attribute.Int64SliceValue([]int64{1, 2, 3}),
			attribute.Int64SliceValue([]int64{1, 2}),
			attribute.Int64SliceValue([]int64{1, 3}),
			attribute.BoolSliceValue([]bool{true, false, true}),
			attribute.BoolSliceValue([]bool{true, false}),
			attribute.BoolSliceValue([]bool{true, true}),
			attribute.StringSliceValue([]string{"a", "b", "c"}),
			attribute.StringSliceValue([]string{"a", "b"}),
			attribute.StringSliceValue([]string{"a", "c"}),
		}
		k1 = attribute.Key("Q")
		k2 = attribute.Key("R")
	)

	for i := range values {
		for j := range values {
			if i == j {
				require.True(t,
					attributesEqual(
						attrs(kv(k1, values[i])),
						attrs(kv(k1, values[i])),
					))
				require.False(t,
					attributesEqual(
						attrs(kv(k1, values[i])),
						attrs(kv(k2, values[i])),
					))
				continue
			}
			require.False(t,
				attributesEqual(
					attrs(kv(k1, values[i])),
					attrs(kv(k1, values[j])),
				))
			require.False(t,
				attributesEqual(
					attrs(kv(k1, values[i]), kv(k2, values[j])),
					attrs(kv(k1, values[j])),
				))
		}
	}
}

const (
	// These constants create a fingerprint collision verified and
	// used in tests below.

	fpKey  = "k"
	fpInt1 = 7966407559257361274
	fpInt2 = 2645356950517223448
)

func TestFingerprintCollision(t *testing.T) {
	// To find this collision I ran the following program for
	// about 18 hours. It would OOM and start again at around 2^30
	// entries on an e2-highmem-16 machine.
	// https://gist.github.com/jmacd/610dda1db6bfd2cde86d5395ce17334e
	//
	// According to https://en.wikipedia.org/wiki/Birthday_problem#Approximations,
	// collisions would take place just before OOM with probability 3%.

	require.Equal(t,
		fingerprintAttributes([]attribute.KeyValue{attribute.Int64(fpKey, fpInt1)}),
		fingerprintAttributes([]attribute.KeyValue{attribute.Int64(fpKey, fpInt2)}),
	)
}

func TestDuplicateFingerprintSafety(t *testing.T) {
	ctx := context.Background()
	lib := instrumentation.Scope{
		Name: "testlib",
	}
	vcs := make([]*viewstate.Compiler, 2)
	vcs[0] = viewstate.New(lib, view.New(
		"test",
		safePerf,
		deltaSelector,
		view.WithClause(
			view.WithKeys([]attribute.Key{fpKey}),
		),
	))
	vcs[1] = viewstate.New(lib, view.New(
		"test",
		safePerf,
		deltaSelector,
		view.WithClause(
			view.WithKeys([]attribute.Key{}),
		),
	))

	desc := test.Descriptor("c", sdkinstrument.SyncCounter, number.Float64Kind)

	pipes := make(pipeline.Register[viewstate.Instrument], 2)
	pipes[0], _ = vcs[0].Compile(desc)
	pipes[1], _ = vcs[1].Compile(desc)

	require.NotNil(t, pipes[0])
	require.NotNil(t, pipes[1])

	inst := New(desc, safePerf, nil, pipes)
	require.NotNil(t, inst)

	attr1 := attribute.Int64(fpKey, fpInt1)
	attr2 := attribute.Int64(fpKey, fpInt2)

	inst.ObserveFloat64(ctx, 1, attrsConfig(attr1))
	inst.ObserveFloat64(ctx, 2, attrsConfig(attr2))

	// collect reader 0
	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
			test.Point(middleTime, endTime,
				sum.NewMonotonicFloat64(1),
				aggregation.DeltaTemporality,
				attr1,
			),
			test.Point(middleTime, endTime,
				sum.NewMonotonicFloat64(2),
				aggregation.DeltaTemporality,
				attr2,
			),
		),
	)

	// There are 2 entries in memory
	require.Equal(t, 2, vcs[0].Collectors()[0].InMemorySize())

	// collect reader 0 again
	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
		),
	)

	// There are 0 entries in memory
	require.Equal(t, 0, vcs[0].Collectors()[0].InMemorySize())

	// Use both again, collect reader 0 again
	inst.ObserveFloat64(ctx, 5, attrsConfig(attr1))
	inst.ObserveFloat64(ctx, 6, attrsConfig(attr2))

	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
			test.Point(middleTime, endTime,
				sum.NewMonotonicFloat64(5),
				aggregation.DeltaTemporality,
				attr1,
			),
			test.Point(middleTime, endTime,
				sum.NewMonotonicFloat64(6),
				aggregation.DeltaTemporality,
				attr2,
			),
		),
	)

	// There are 2 entries in memory again
	require.Equal(t, 2, vcs[0].Collectors()[0].InMemorySize())

	// Update attr1, collect reader 0
	inst.ObserveFloat64(ctx, 25, attrsConfig(attr1))

	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
			test.Point(middleTime, endTime,
				sum.NewMonotonicFloat64(25),
				aggregation.DeltaTemporality,
				attr1,
			),
		),
	)

	// Only 1 entry in memory
	require.Equal(t, 1, vcs[0].Collectors()[0].InMemorySize())

	// Update attr2, collect reader 0
	inst.ObserveFloat64(ctx, 32, attrsConfig(attr2))

	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
			test.Point(middleTime, endTime,
				sum.NewMonotonicFloat64(32),
				aggregation.DeltaTemporality,
				attr2,
			),
		),
	)

	// Only 1 entry in memory
	require.Equal(t, 1, vcs[0].Collectors()[0].InMemorySize())

	// No updates, clear memory
	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
		),
	)

	require.Equal(t, 0, vcs[0].Collectors()[0].InMemorySize())

	// collect reader 1
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[1].Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
			test.Point(middleTime, endTime,
				// Attributes removed. 1+2+5+6+25+32
				sum.NewMonotonicFloat64(71),
				aggregation.DeltaTemporality,
			),
		),
	)

}

func TestDuplicateFingerprintCollisionIgnored(t *testing.T) {
	ctx := context.Background()
	lib := instrumentation.Scope{
		Name: "testlib",
	}
	vcs := make([]*viewstate.Compiler, 1)
	vcs[0] = viewstate.New(lib, view.New(
		"test",
		safePerf,
		deltaSelector,
	))

	desc := test.Descriptor("c", sdkinstrument.SyncCounter, number.Float64Kind)

	pipes := make(pipeline.Register[viewstate.Instrument], 1)
	pipes[0], _ = vcs[0].Compile(desc)

	require.NotNil(t, pipes[0])

	inst := New(desc, sdkinstrument.Performance{
		// Do not check the collision.
		IgnoreCollisions:          true,
		InactiveCollectionPeriods: 1,
	}, nil, pipes)
	require.NotNil(t, inst)

	attr1 := attribute.Int64(fpKey, fpInt1)
	attr2 := attribute.Int64(fpKey, fpInt2)

	// Because of the duplicate, the first attribute set wins.
	inst.ObserveFloat64(ctx, 1, attrsConfig(attr1))
	inst.ObserveFloat64(ctx, 2, attrsConfig(attr2))

	// collect reader
	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
			test.Point(middleTime, endTime,
				sum.NewMonotonicFloat64(3), // combined values
				aggregation.DeltaTemporality,
				attr1, // first attribute set observed
			),
		),
	)

	// There is 1 entry in memory
	require.Equal(t, 1, vcs[0].Collectors()[0].InMemorySize())

	// collect reader again
	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
		),
	)

	// There are 0 entries in memory
	require.Equal(t, 0, vcs[0].Collectors()[0].InMemorySize())

	// Use both attribute sets in the opposite order, collect
	// reader again.
	inst.ObserveFloat64(ctx, 6, attrsConfig(attr2))
	inst.ObserveFloat64(ctx, 5, attrsConfig(attr1))

	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
			test.Point(middleTime, endTime,
				sum.NewMonotonicFloat64(11), // combined values
				aggregation.DeltaTemporality,
				attr2, // first attribute set observed
			),
		),
	)
}

func TestRecordInactivity(t *testing.T) {
	for inactive := uint32(1); inactive < 10; inactive++ {
		t.Run(fmt.Sprint(inactive), func(t *testing.T) {
			ctx := context.Background()
			lib := instrumentation.Scope{
				Name: "testlib",
			}
			vcs := make([]*viewstate.Compiler, 1)
			vcs[0] = viewstate.New(lib, view.New(
				"test",
				safePerf,
				deltaSelector,
			))

			desc := test.Descriptor("c", sdkinstrument.SyncCounter, number.Float64Kind)

			pipes := make(pipeline.Register[viewstate.Instrument], 1)
			pipes[0], _ = vcs[0].Compile(desc)

			require.NotNil(t, pipes[0])

			inst := New(desc, sdkinstrument.Performance{
				InactiveCollectionPeriods: inactive,
			}, nil, pipes)
			require.NotNil(t, inst)

			attr := attribute.Int64(fpKey, fpInt1)

			expectNothing := func() {
				test.RequireEqualMetrics(
					t,
					test.CollectScope(
						t,
						vcs[0].Collectors(),
						testSequence,
					),
					test.Instrument(
						desc,
					),
				)
			}
			expectSomething := func(value float64) {
				test.RequireEqualMetrics(
					t,
					test.CollectScope(
						t,
						vcs[0].Collectors(),
						testSequence,
					),
					test.Instrument(
						desc,
						test.Point(middleTime, endTime,
							sum.NewMonotonicFloat64(value),
							aggregation.DeltaTemporality,
							attr,
						),
					),
				)
			}

			inst.ObserveFloat64(ctx, 17, attrsConfig(attr))

			// collect reader
			inst.SnapshotAndProcess()
			expectSomething(17)

			// There is 1 entry in memory
			require.Equal(t, 1, vcs[0].Collectors()[0].InMemorySize())

			repeatedlyNothing := func() {
				// For 1 less than the inactive allowance,
				// expect nothing collected and one record
				// remaining in memory.
				for i := uint32(1); i < inactive; i++ {
					// collect reader again.
					inst.SnapshotAndProcess()
					expectNothing()

					// There is still 1 entry in memory.
					require.Equal(t, 1, vcs[0].Collectors()[0].InMemorySize())
				}
			}

			// Inactive for the up to the allowed inactivity.
			repeatedlyNothing()

			inst.ObserveFloat64(ctx, 23, attrsConfig(attr))

			// collect reader again, but an update came in.
			inst.SnapshotAndProcess()
			expectSomething(23)

			// There is still 1 entry in memory, the count has reset.
			require.Equal(t, 1, vcs[0].Collectors()[0].InMemorySize())

			// Inactive for the up to the allowed inactivity.
			repeatedlyNothing()

			// collect reader again
			inst.SnapshotAndProcess()
			expectNothing()

			// There are now 0 entries in memory.
			require.Equal(t, 0, vcs[0].Collectors()[0].InMemorySize())

		})
	}
}

func TestCardinalityOverflowCumulative(t *testing.T) {
	const total = 300
	const limit = 133

	ctx := context.Background()
	lib := instrumentation.Scope{
		Name: "testlib",
	}
	perf := sdkinstrument.Performance{
		InstrumentCardinalityLimit: limit,
	}
	vcs := make([]*viewstate.Compiler, 1)
	vcs[0] = viewstate.New(lib, view.New("test", perf))

	desc := test.Descriptor("c", sdkinstrument.SyncCounter, number.Float64Kind)

	pipes := make(pipeline.Register[viewstate.Instrument], 1)
	pipes[0], _ = vcs[0].Compile(desc)

	inst := New(desc, perf, nil, pipes)
	require.NotNil(t, inst)

	var expectPoints []data.Point
	var oflow int

	for i := 0; i < total; i++ {
		inst.ObserveFloat64(ctx, 1, attrsConfig(attribute.Int("i", i)))

		if i < int(perf.InstrumentCardinalityLimit)-1 {
			expectPoints = append(expectPoints, test.Point(
				startTime, endTime,
				sum.NewMonotonicFloat64(1),
				aggregation.CumulativeTemporality,
				attribute.Int("i", i),
			))
		} else {
			oflow++
		}
	}
	require.Equal(t, total-limit+1, oflow)
	expectPoints = append(expectPoints, test.Point(
		startTime, endTime,
		sum.NewMonotonicFloat64(float64(oflow)),
		aggregation.CumulativeTemporality,
		attribute.Bool("otel.metric.overflow", true),
	))

	inst.SnapshotAndProcess()

	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
			expectPoints...,
		),
	)

	// Repeat with all new attributes; since this is cumulative,
	// the original set is retained an this goes entirely to overflow.
	for i := 0; i < total; i++ {
		inst.ObserveFloat64(ctx, 1, attrsConfig(attribute.Int("i", total+i)))
		oflow++
	}
	// Replace the overflow value
	expectPoints = expectPoints[:len(expectPoints)-1]
	expectPoints = append(expectPoints, test.Point(
		startTime, endTime,
		sum.NewMonotonicFloat64(float64(oflow)),
		aggregation.CumulativeTemporality,
		attribute.Bool("otel.metric.overflow", true),
	))

	inst.SnapshotAndProcess()

	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
			expectPoints...,
		),
	)
}

func TestCardinalityOverflowAvoidedDelta(t *testing.T) {
	const limit = 100 // has to be even

	ctx := context.Background()
	lib := instrumentation.Scope{
		Name: "testlib",
	}
	perf := sdkinstrument.Performance{
		InstrumentCardinalityLimit: limit,
		InactiveCollectionPeriods:  1,
	}
	vcs := make([]*viewstate.Compiler, 1)
	vcs[0] = viewstate.New(lib, view.New("test", perf, deltaSelector))

	desc := test.Descriptor("c", sdkinstrument.SyncUpDownCounter, number.Int64Kind)

	pipes := make(pipeline.Register[viewstate.Instrument], 1)
	pipes[0], _ = vcs[0].Compile(desc)

	inst := New(desc, perf, nil, pipes)
	require.NotNil(t, inst)

	uniq := 0
	for reps := 0; reps < 4; reps++ {

		var expectPoints []data.Point

		// we expect to create half the limit less on every
		// iteration and never overflow.
		for i := 0; i < limit/2-1; i++ {
			// attr is unique on every iteration
			attr := attribute.String("s", fmt.Sprint(uniq))
			uniq++
			inst.ObserveInt64(ctx, 1, attrsConfig(attr))

			expectPoints = append(expectPoints, test.Point(
				middleTime, endTime,
				sum.NewNonMonotonicInt64(1),
				aggregation.DeltaTemporality,
				attr,
			))
		}

		inst.SnapshotAndProcess()

		test.RequireEqualMetrics(
			t,
			test.CollectScope(
				t,
				vcs[0].Collectors(),
				testSequence,
			),
			test.Instrument(
				desc,
				expectPoints...,
			),
		)
	}
}

func TestCardinalityOverflowOscillationDelta(t *testing.T) {
	// Note this tests a fairly bad behavior; it can be improved
	// upon, but at least this validates the current mechanism.
	// If more than half of the available aggregators are not
	// re-used an oscillation develops.
	//
	// TODO improve this logic and replace this test.
	const total = 8
	const limit = 4

	ctx := context.Background()
	lib := instrumentation.Scope{
		Name: "testlib",
	}
	perf := sdkinstrument.Performance{
		InstrumentCardinalityLimit: limit,
		InactiveCollectionPeriods:  1,
	}
	vcs := make([]*viewstate.Compiler, 1)
	vcs[0] = viewstate.New(lib, view.New("test", perf, deltaSelector))

	desc := test.Descriptor("c", sdkinstrument.SyncCounter, number.Float64Kind)

	pipes := make(pipeline.Register[viewstate.Instrument], 1)
	pipes[0], _ = vcs[0].Compile(desc)

	inst := New(desc, perf, nil, pipes)
	require.NotNil(t, inst)

	uniq := 0

	for reps := 0; reps < 10; reps++ {
		var expectPoints []data.Point
		var oflow int

		for i := 0; i < total; i++ {
			inst.ObserveFloat64(ctx, 1, attrsConfig(attribute.Int("uniq", uniq)))

			// Note: on even intervals, the overflow will
			// be as we expect, and in odd intervals, the
			// entire batch will overflow.
			if reps%2 == 0 {
				normal := int(perf.InstrumentCardinalityLimit) - 1
				if reps > 0 {
					// Except the first round there's one
					// overflow element already used so one
					// less normal "new" attribute set.
					normal--
				}
				if i < normal {
					expectPoints = append(expectPoints, test.Point(
						middleTime, endTime,
						sum.NewMonotonicFloat64(1),
						aggregation.DeltaTemporality,
						attribute.Int("uniq", uniq),
					))
				} else {
					oflow++
				}
			} else {
				oflow++
			}
			uniq++
		}
		expectPoints = append(expectPoints, test.Point(
			middleTime, endTime,
			sum.NewMonotonicFloat64(float64(oflow)),
			aggregation.DeltaTemporality,
			attribute.Bool("otel.metric.overflow", true),
		))

		inst.SnapshotAndProcess()

		test.RequireEqualMetrics(
			t,
			test.CollectScope(
				t,
				vcs[0].Collectors(),
				testSequence,
			),
			test.Instrument(
				desc,
				expectPoints...,
			),
		)
	}
}

func TestInputAttributeSliceRaceCondition(t *testing.T) {
	ctx := context.Background()
	lib := instrumentation.Scope{
		Name: "testlib",
	}
	perf := sdkinstrument.Performance{}
	vcs := make([]*viewstate.Compiler, 1)
	vcs[0] = viewstate.New(lib, view.New("test", perf))

	desc := test.Descriptor("c", sdkinstrument.SyncCounter, number.Float64Kind)

	pipes := make(pipeline.Register[viewstate.Instrument], 1)
	pipes[0], _ = vcs[0].Compile(desc)

	inst := New(desc, perf, nil, pipes)
	require.NotNil(t, inst)

	// This attribute slice is shared by multiple concurrent callers.
	attrs := []attribute.KeyValue{
		{
			Key:   "key1",
			Value: attribute.StringValue("val1"),
		},
		{
			Key:   "key2",
			Value: attribute.StringValue("val2"),
		},
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			inst.ObserveFloat64(ctx, 1, attrsConfig(attrs...))
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			inst.ObserveFloat64(ctx, 1, attrsConfig(attrs...))
		}
	}()

	wg.Wait()
}

func TestAttributeSizeLimit(t *testing.T) {
	ctx := context.Background()
	lib := instrumentation.Scope{
		Name: "testlib",
	}
	// This enters a default value
	perf := sdkinstrument.Performance{}.Validate()
	perf.AttributeSizeLimit = 100
	vcs := make([]*viewstate.Compiler, 2)
	vcs[0] = viewstate.New(lib, view.New(
		"test",
		perf,
		deltaSelector,
	))

	desc := test.Descriptor(
		"counter",
		sdkinstrument.SyncCounter,
		number.Float64Kind)

	pipes := make(pipeline.Register[viewstate.Instrument], 1)
	pipes[0], _ = vcs[0].Compile(desc)

	require.NotNil(t, pipes[0])

	inst := New(desc, perf, nil, pipes)
	require.NotNil(t, inst)

	inst.ObserveFloat64(ctx, 1, attrsConfig(attribute.String(strings.Repeat("X", 1<<20), strings.Repeat("Y", 1<<20))))
	inst.ObserveFloat64(ctx, 2, attrsConfig(attribute.String(strings.Repeat("X", 1<<20), strings.Repeat("Y", 1<<20))))
	inst.ObserveFloat64(ctx, 3, attrsConfig(attribute.String(strings.Repeat("X", 1<<20), strings.Repeat("Y", 1<<20))))

	limit := int(perf.AttributeSizeLimit)

	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs[0].Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
			test.Point(middleTime, endTime,
				sum.NewMonotonicFloat64(6),
				aggregation.DeltaTemporality,
				attribute.String(strings.Repeat("X", limit), strings.Repeat("Y", limit)),
			),
		),
	)
}

func TestSyncExemplars(t *testing.T) {
	lib := instrumentation.Scope{
		Name: "testlib",
	}
	perf := sdkinstrument.Performance{
		ExemplarsEnabled: 5,
	}
	vcs := viewstate.New(lib, view.New(
		"test",
		perf,
		deltaSelector,
		view.WithClause(
			view.WithKeys([]attribute.Key{"a"}),
		),
	))
	desc := test.Descriptor(
		"histo",
		sdkinstrument.SyncHistogram,
		number.Int64Kind)

	pipes := make(pipeline.Register[viewstate.Instrument], 1)
	pipes[0], _ = vcs.Compile(desc)

	require.NotNil(t, pipes[0])
	inst := New(desc, perf, nil, pipes)
	require.NotNil(t, inst)

	untracedCtx := context.Background()
	attrs1 := []attribute.KeyValue{attribute.String("a", "1"), attribute.String("b", "1")}
	attrs2 := []attribute.KeyValue{attribute.String("a", "1"), attribute.String("b", "2")}
	attrs3 := []attribute.KeyValue{attribute.String("a", "1"), attribute.String("b", "3")}

	for i := int64(0); i < 5; i++ {
		inst.ObserveInt64(
			trace.ContextWithSpan(context.Background(), test.FakeSpan(1, byte(i+1))),
			1+i,
			attrsConfig(attrs1...),
		)
		inst.ObserveInt64(untracedCtx, 1+i, attrsConfig(attrs2...))
		inst.ObserveInt64(untracedCtx, 1+i, attrsConfig(attrs3...))
	}

	inst.SnapshotAndProcess()
	test.RequireEqualMetrics(
		t,
		test.CollectScope(
			t,
			vcs.Collectors(),
			testSequence,
		),
		test.Instrument(
			desc,
			test.PointEx(
				middleTime, endTime,
				histogram.NewInt64(histostruct.NewConfig(),
					1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4, 5, 5, 5,
				),
				aggregation.DeltaTemporality,
				[]attribute.KeyValue{attribute.String("a", "1")},
				// Note: do not set timestamp below
				aggregator.WeightedExemplarBits{
					ExemplarBits: aggregator.ExemplarBits{
						Number:     number.FromInt64(1),
						Attributes: attrs1,
						Span:       test.FakeSpan(1, 1),
					},
					Weight: 1,
				},
				aggregator.WeightedExemplarBits{
					ExemplarBits: aggregator.ExemplarBits{
						Number:     number.FromInt64(2),
						Attributes: attrs1,
						Span:       test.FakeSpan(1, 2),
					},
					Weight: 1,
				},
				aggregator.WeightedExemplarBits{
					ExemplarBits: aggregator.ExemplarBits{
						Number:     number.FromInt64(3),
						Attributes: attrs1,
						Span:       test.FakeSpan(1, 3),
					},
					Weight: 1,
				},
				aggregator.WeightedExemplarBits{
					ExemplarBits: aggregator.ExemplarBits{
						Number:     number.FromInt64(4),
						Attributes: attrs1,
						Span:       test.FakeSpan(1, 4),
					},
					Weight: 1,
				},
				aggregator.WeightedExemplarBits{
					ExemplarBits: aggregator.ExemplarBits{
						Number:     number.FromInt64(5),
						Attributes: attrs1,
						Span:       test.FakeSpan(1, 5),
					},
					Weight: 1,
				},
			),
		),
	)
}
