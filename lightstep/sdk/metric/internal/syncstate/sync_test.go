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
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

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
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/instrumentation"
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
)

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
	// Note: prior to
	// https://github.com/lightstep/otel-launcher-go/pull/206 this
	// code was able to reproduce the race condition handled in
	// this code.  The race condition still exists, but the test
	// no longer covers the call to Gosched() in acquireRecord()
	// or the return-nil branch in acquireWrite().  This is
	// because with the RWMutex, the race is much less racey.
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

	lib := instrumentation.Library{
		Name: "testlib",
	}
	vcs := make([]*viewstate.Compiler, numReaders)
	for vci := range vcs {
		vcs[vci] = viewstate.New(lib, view.New("test", vopts...))
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

	inst := NewInstrument(desc, nil, pipes)
	require.NotNil(t, inst)

	cntr := NewCounter[N, Traits](inst)
	require.NotNil(t, cntr)

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

			for j := 0; j < numUpdates/numRoutines; j++ {
				cntr.Add(ctx, 1, attrs[rnd.Intn(len(attrs))])
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
	lib := instrumentation.Library{
		Name: "testlib",
	}
	vcs := make([]*viewstate.Compiler, 2)
	vcs[0] = viewstate.New(lib, view.New("dropper", vopts...))
	vcs[1] = viewstate.New(lib, view.New("keeper"))

	desc := test.Descriptor("dropme", sdkinstrument.SyncHistogram, number.Float64Kind)

	pipes := make(pipeline.Register[viewstate.Instrument], 2)
	pipes[0], _ = vcs[0].Compile(desc)
	pipes[1], _ = vcs[1].Compile(desc)

	require.Nil(t, pipes[0])
	require.NotNil(t, pipes[1])

	inst := NewInstrument(desc, nil, pipes)
	require.NotNil(t, inst)

	hist := NewHistogram[float64, number.Float64Traits](inst)
	require.NotNil(t, hist)

	hist.Record(ctx, 1)
	hist.Record(ctx, 2)
	hist.Record(ctx, 3)

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
	lib := instrumentation.Library{
		Name: "testlib",
	}
	vcs := make([]*viewstate.Compiler, 2)
	vcs[0] = viewstate.New(lib, view.New("dropper", vopts...))
	vcs[1] = viewstate.New(lib, view.New("keeper", vopts...))

	desc := test.Descriptor("dropme", sdkinstrument.SyncHistogram, number.Float64Kind)

	pipes := make(pipeline.Register[viewstate.Instrument], 2)
	pipes[0], _ = vcs[0].Compile(desc)
	pipes[1], _ = vcs[1].Compile(desc)

	require.Nil(t, pipes[0])
	require.Nil(t, pipes[1])

	inst := NewInstrument(desc, nil, pipes)
	require.Nil(t, inst)

	hist := NewHistogram[float64, number.Float64Traits](inst)
	require.NotNil(t, hist)

	hist.Record(ctx, 1)
	hist.Record(ctx, 2)
	hist.Record(ctx, 3)

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
		lib := instrumentation.Library{
			Name: "testlib",
		}
		vcs := make([]*viewstate.Compiler, 1)
		vcs[0] = viewstate.New(lib, view.New("test"))

		pipes := make(pipeline.Register[viewstate.Instrument], 1)
		pipes[0], _ = vcs[0].Compile(desc)

		inst := NewInstrument(desc, nil, pipes)
		require.NotNil(t, inst)

		var negOne aggregation.Aggregation

		if desc.NumberKind == number.Float64Kind {
			cntr := NewCounter[float64, number.Float64Traits](inst)

			cntr.Add(ctx, -1)
			cntr.Add(ctx, math.NaN())
			cntr.Add(ctx, math.Inf(+1))
			cntr.Add(ctx, math.Inf(-1))
			negOne = sum.NewNonMonotonicFloat64(-1)
		} else {
			cntr := NewCounter[int64, number.Int64Traits](inst)

			cntr.Add(ctx, -1)
			negOne = sum.NewNonMonotonicInt64(-1)
		}

		inst.SnapshotAndProcess()

		var expectPoints []data.Point

		if desc.Kind == sdkinstrument.SyncUpDownCounter {
			expectPoints = append(expectPoints, test.Point(
				startTime, endTime,
				negOne,
				aggregation.CumulativeTemporality,
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
	lib := instrumentation.Library{
		Name: "testlib",
	}
	vcs := make([]*viewstate.Compiler, 2)
	vcs[0] = viewstate.New(lib, view.New(
		"test",
		deltaSelector,
		view.WithClause(
			view.WithKeys([]attribute.Key{"A", "C"}),
		),
	))

	indesc := test.Descriptor(
		"syncgauge",
		sdkinstrument.SyncUpDownCounter,
		number.Float64Kind,
		instrument.WithDescription(`{
  "aggregation": "gauge",
  "description": "incredible"
}`))

	outdesc := test.Descriptor(
		"syncgauge",
		sdkinstrument.SyncUpDownCounter,
		number.Float64Kind,
		instrument.WithDescription("incredible"),
	)

	pipes := make(pipeline.Register[viewstate.Instrument], 1)
	pipes[0], _ = vcs[0].Compile(indesc)

	require.NotNil(t, pipes[0])

	inst := NewInstrument(indesc, nil, pipes)
	require.NotNil(t, inst)

	sg := NewCounter[float64, number.Float64Traits](inst)
	require.NotNil(t, sg)

	sg.Add(ctx, 1)
	sg.Add(ctx, 2)
	sg.Add(ctx, 3)

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
	sg.Add(ctx, 172)
	sg.Add(ctx, 175)

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
	sg.Add(ctx, 1333, attribute.String("A", "B"))
	sg.Add(ctx, 1337, attribute.String("C", "D"))

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
		sg.Add(ctx, float64(i), attribute.Int("ignored", i), attribute.String("A", "B"))
	}
	for i := 1000; i > 0; i-- {
		sg.Add(ctx, float64(i), attribute.Int("ignored", i), attribute.String("C", "D"))
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

func TestDuplicateFingerprint(t *testing.T) {
	ctx := context.Background()
	lib := instrumentation.Library{
		Name: "testlib",
	}
	vcs := make([]*viewstate.Compiler, 2)
	vcs[0] = viewstate.New(lib, view.New(
		"test",
		deltaSelector,
		view.WithClause(
			view.WithKeys([]attribute.Key{fpKey}),
		),
	))
	vcs[1] = viewstate.New(lib, view.New(
		"test",
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

	inst := NewInstrument(desc, nil, pipes)
	require.NotNil(t, inst)

	sg := NewCounter[float64, number.Float64Traits](inst)
	require.NotNil(t, sg)

	attr1 := attribute.Int64(fpKey, fpInt1)
	attr2 := attribute.Int64(fpKey, fpInt2)

	sg.Add(ctx, 1, attr1)
	sg.Add(ctx, 2, attr2)

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
	require.Equal(t, 2, vcs[0].Collectors()[0].Size())

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
	require.Equal(t, 0, vcs[0].Collectors()[0].Size())

	// Use both again, collect reader 0 again
	sg.Add(ctx, 5, attr1)
	sg.Add(ctx, 6, attr2)

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
	require.Equal(t, 2, vcs[0].Collectors()[0].Size())

	// Update attr1, collect reader 0
	sg.Add(ctx, 25, attr1)

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
	require.Equal(t, 1, vcs[0].Collectors()[0].Size())

	// Update attr2, collect reader 0
	sg.Add(ctx, 32, attr2)

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
	require.Equal(t, 1, vcs[0].Collectors()[0].Size())

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

	require.Equal(t, 0, vcs[0].Collectors()[0].Size())

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
