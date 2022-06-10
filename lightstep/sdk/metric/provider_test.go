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

package metric // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/sum"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/test"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

// Tests that calls to Produce() will re-use the underlying memory up to their capacity.
func TestOutputReuse(t *testing.T) {
	ctx := context.Background()

	rdr := NewManualReader("test")
	res := resource.Empty()
	provider := NewMeterProvider(WithReader(rdr), WithResource(res))

	cntr := must(provider.Meter("test").SyncInt64().Counter("hello"))

	var reuse data.Metrics
	var notime time.Time
	const cumulative = aggregation.CumulativeTemporality
	attr := attribute.Int("K", 1)
	attr2 := attribute.Int("L", 1)

	cntr.Add(ctx, 1, attr)

	// With a single point, check correct initial result.
	reuse = rdr.Produce(&reuse)

	test.RequireEqualResourceMetrics(
		t, reuse, res,
		test.Scope(
			test.Library("test"),
			test.Instrument(
				test.Descriptor("hello", sdkinstrument.SyncCounter, number.Int64Kind),
				test.Point(notime, notime, sum.NewMonotonicInt64(1), cumulative, attr),
			),
		),
	)

	// Now ensure that repeated Produce gives the same point
	// address, new result.
	pointAddrBefore := &reuse.Scopes[0].Instruments[0].Points[0]

	cntr.Add(ctx, 1, attribute.Int("K", 1))

	reuse = rdr.Produce(&reuse)

	test.RequireEqualResourceMetrics(
		t, reuse, res,
		test.Scope(
			test.Library("test"),
			test.Instrument(
				test.Descriptor("hello", sdkinstrument.SyncCounter, number.Int64Kind),
				test.Point(notime, notime, sum.NewMonotonicInt64(2), cumulative, attr),
			),
		),
	)

	pointAddrAfter := &reuse.Scopes[0].Instruments[0].Points[0]

	require.Equal(t, pointAddrBefore, pointAddrAfter)

	// Give the points list a cap of 17 and make a second point, ensure
	reuse.Scopes[0].Instruments[0].Points = make([]data.Point, 0, 17)

	cntr.Add(ctx, 1, attr2)

	reuse = rdr.Produce(&reuse)

	require.Equal(t, 17, cap(reuse.Scopes[0].Instruments[0].Points))

	test.RequireEqualResourceMetrics(
		t, reuse, res,
		test.Scope(
			test.Library("test"),
			test.Instrument(
				test.Descriptor("hello", sdkinstrument.SyncCounter, number.Int64Kind),
				test.Point(notime, notime, sum.NewMonotonicInt64(2), cumulative, attr),
				test.Point(notime, notime, sum.NewMonotonicInt64(1), cumulative, attr2),
			),
		),
	)

	// Make 20 more points, ensure capacity is greater than 17
	for i := 0; i < 20; i++ {
		cntr.Add(ctx, 1, attribute.Int("I", i))
	}

	reuse = rdr.Produce(&reuse)

	require.Less(t, 17, cap(reuse.Scopes[0].Instruments[0].Points))

	// Get the new address of scope-0, instrument-0, point-0:
	pointAddrBefore = &reuse.Scopes[0].Instruments[0].Points[0]

	// Register more meters; expecting a capacity of 1 before and
	// >1 after (but still the same point addresses).
	require.Equal(t, 1, cap(reuse.Scopes))

	fcntr := must(provider.Meter("real").SyncFloat64().Counter("goodbye"))

	fcntr.Add(ctx, 2, attr)

	reuse = rdr.Produce(&reuse)

	pointAddrAfter = &reuse.Scopes[0].Instruments[0].Points[0]

	require.Equal(t, pointAddrBefore, pointAddrAfter)
	require.Less(t, 1, cap(reuse.Scopes))

	require.Equal(t, "test", reuse.Scopes[0].Library.Name)
	require.Equal(t, "real", reuse.Scopes[1].Library.Name)
}

type testReader struct {
	flushes   int
	shutdowns int
	retval    error
}

func (t *testReader) String() string {
	return "testreader"
}

func (t *testReader) Register(_ Producer) {
}

func (t *testReader) ForceFlush(_ context.Context) error {
	t.flushes++
	return t.retval
}

func (t *testReader) Shutdown(_ context.Context) error {
	t.shutdowns++
	return t.retval
}

func TestForceFlush(t *testing.T) {
	// Test with 0 through 2 readers, ForceFlush() with and without errors.
	ctx := context.Background()
	provider := NewMeterProvider()

	require.NoError(t, provider.ForceFlush(ctx))

	rdr1 := &testReader{}
	provider = NewMeterProvider(WithReader(rdr1))

	require.NoError(t, provider.ForceFlush(ctx))
	require.Equal(t, 1, rdr1.flushes)

	rdr1.retval = fmt.Errorf("flush fail")
	err := provider.ForceFlush(ctx)
	require.Error(t, err)
	require.Equal(t, "flush fail", err.Error())
	require.Equal(t, 2, rdr1.flushes)

	rdr2 := &testReader{}
	provider = NewMeterProvider(
		WithReader(rdr2),
		WithReader(rdr1),
		WithReader(NewManualReader("also_tested")), // ManualReader.ForceFlush cannot error
	)

	err = provider.ForceFlush(ctx)
	require.Error(t, err)
	require.Equal(t, "flush fail", err.Error())
	require.Equal(t, 3, rdr1.flushes)
	require.Equal(t, 1, rdr2.flushes)

	rdr1.retval = nil

	err = provider.ForceFlush(ctx)
	require.NoError(t, err)
	require.Equal(t, 4, rdr1.flushes)
	require.Equal(t, 2, rdr2.flushes)

	require.Equal(t, 0, rdr1.shutdowns)
	require.Equal(t, 0, rdr2.shutdowns)
}

func TestShutdown(t *testing.T) {
	ctx := context.Background()

	// Shutdown with 0 meters; repeat shutdown causes error.
	provider := NewMeterProvider()

	require.NoError(t, provider.Shutdown(ctx))
	err := provider.Shutdown(ctx)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrAlreadyShutdown))

	// Shutdown with 1 meters
	rdr1 := &testReader{}
	provider = NewMeterProvider(WithReader(rdr1))

	require.NoError(t, provider.Shutdown(ctx))
	require.Equal(t, 1, rdr1.shutdowns)
	require.Equal(t, 0, rdr1.flushes)

	err = provider.Shutdown(ctx)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrAlreadyShutdown))

	// Shutdown with 3 meters, 2 errors
	rdr1.retval = fmt.Errorf("first error")

	rdr2 := &testReader{}
	rdr2.retval = fmt.Errorf("second error")

	provider = NewMeterProvider(
		WithReader(rdr1),
		WithReader(rdr2),
		WithReader(NewManualReader("also_tested")), // ManualReader.Shutdown cannot error
	)

	err = provider.Shutdown(ctx)

	require.Error(t, err)
	require.Contains(t, err.Error(), "first error")
	require.Contains(t, err.Error(), "second error")
	require.Equal(t, 2, rdr1.shutdowns)
	require.Equal(t, 1, rdr2.shutdowns)
	require.Equal(t, 0, rdr1.flushes)
	require.Equal(t, 0, rdr2.flushes)
}

func TestManualReaderRegisteredTwice(t *testing.T) {
	rdr := NewManualReader("tester")
	errs := test.OTelErrors()

	_ = NewMeterProvider(WithReader(rdr))
	_ = NewMeterProvider(WithReader(rdr))

	require.Equal(t, 1, len(*errs))
	require.True(t, errors.Is((*errs)[0], ErrMultipleReaderRegistration))
}

func TestDuplicateInstrumentConflict(t *testing.T) {
	rdr := NewManualReader("test")
	res := resource.Empty()
	errs := test.OTelErrors()

	provider := NewMeterProvider(WithReader(rdr), WithResource(res))

	// int64/float64 conflicting registration
	icntr := must(provider.Meter("test").SyncInt64().Counter("counter"))
	fcntr, err := provider.Meter("test").SyncFloat64().Counter("counter")

	require.NotNil(t, icntr)
	require.NotNil(t, fcntr)
	require.NotEqual(t, icntr, fcntr)
	require.Error(t, err)

	expected := "SyncCounter-Int64-MonotonicSum, SyncCounter-Float64-MonotonicSum"

	require.Contains(t, err.Error(), expected)

	// re-register the first instrument, get a failure
	icntr2, err := provider.Meter("test").SyncInt64().Counter("counter")
	require.Error(t, err)
	require.Contains(t, err.Error(), expected)
	require.Equal(t, icntr, icntr2)

	// there is only one conflict passed to otel.Handle()--only
	// the first error is handled, even though it happened twice.
	require.Equal(t, 1, len(*errs))
	require.True(t, errors.Is((*errs)[0], viewstate.ViewConflictsError{}))
}
