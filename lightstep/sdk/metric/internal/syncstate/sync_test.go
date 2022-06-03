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
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
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
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

func deltaUpdate(old, new int64) int64 {
	return old + new
}

func cumulativeUpdate(_, new int64) int64 {
	return new
}

func TestSyncStateDeltaConcurrency1(t *testing.T) {
	testSyncStateConcurrency(t, 1, aggregation.DeltaTemporality, deltaUpdate)
}

func TestSyncStateDeltaConcurrency2(t *testing.T) {
	testSyncStateConcurrency(t, 2, aggregation.DeltaTemporality, deltaUpdate)
}

func TestSyncStateCumulativeConcurrency1(t *testing.T) {
	testSyncStateConcurrency(t, 1, aggregation.CumulativeTemporality, cumulativeUpdate)
}

func TestSyncStateCumulativeConcurrency2(t *testing.T) {
	testSyncStateConcurrency(t, 2, aggregation.CumulativeTemporality, cumulativeUpdate)
}

// TODO: Add float64, Histogram tests.
func testSyncStateConcurrency(t *testing.T, numReaders int, tempo aggregation.Temporality, update func(old, new int64) int64) {
	const (
		numRoutines = 10
		numAttrs    = 10
		numUpdates  = 1e6
	)

	var writers sync.WaitGroup
	var readers sync.WaitGroup
	readers.Add(numReaders)
	writers.Add(numRoutines)

	lib := instrumentation.Library{
		Name: "testlib",
	}
	vopts := []view.Option{
		view.WithDefaultAggregationTemporalitySelector(func(_ sdkinstrument.Kind) aggregation.Temporality {
			return tempo
		}),
	}
	vcs := make([]*viewstate.Compiler, numReaders)
	for vci := range vcs {
		vcs[vci] = viewstate.New(lib, view.New("test", vopts...))
	}
	attrs := make([]attribute.KeyValue, numAttrs)
	for i := range attrs {
		attrs[i] = attribute.Int("i", i)
	}

	desc := test.Descriptor("tester", sdkinstrument.CounterKind, number.Int64Kind)

	pipes := make(pipeline.Register[viewstate.Instrument], numReaders)
	for vci := range vcs {
		pipes[vci], _ = vcs[vci].Compile(desc)
	}

	inst := NewInstrument(desc, nil, pipes)
	require.NotNil(t, inst)

	cntr := NewCounter[int64, number.Int64Traits](inst)
	require.NotNil(t, cntr)

	ctx, cancel := context.WithCancel(context.Background())

	partialCounts := make([]map[attribute.Set]int64, numReaders)
	for vci := range vcs {
		partialCounts[vci] = map[attribute.Set]int64{}
	}

	// Reader loops
	for vci := range vcs {
		go func(vci int, partial map[attribute.Set]int64, vc *viewstate.Compiler) {
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
					partial[pt.Attributes] = update(partial[pt.Attributes], pt.Aggregation.(*sum.MonotonicInt64).Sum().AsInt64())
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
		var sum int64
		for _, count := range partialCounts[vci] {
			sum += count
		}
		require.Equal(t, int64(numUpdates), sum, "vci==%d", vci)
	}
}

func TestSyncStateNoopInstrument(t *testing.T) {
	// TODO: test with disabled instrument
}
