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

package test // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/test"

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
)

func Descriptor(name string, ik sdkinstrument.Kind, nk number.Kind, opts ...instrument.Option) sdkinstrument.Descriptor {
	cfg := instrument.NewConfig(opts...)
	return sdkinstrument.NewDescriptor(name, ik, nk, cfg.Description(), cfg.Unit())
}

func Point(start, end time.Time, agg aggregation.Aggregation, tempo aggregation.Temporality, kvs ...attribute.KeyValue) data.Point {
	attrs := attribute.NewSet(kvs...)
	return data.Point{
		Start:       start,
		End:         end,
		Attributes:  attrs,
		Aggregation: agg,
		Temporality: tempo,
	}
}

func Instrument(desc sdkinstrument.Descriptor, points ...data.Point) data.Instrument {
	return data.Instrument{
		Descriptor: desc,
		Points:     points,
	}
}

func Library(name string, opts ...metric.MeterOption) instrumentation.Library {
	cfg := metric.NewMeterConfig(opts...)
	return instrumentation.Library{
		Name:      name,
		Version:   cfg.InstrumentationVersion(),
		SchemaURL: cfg.SchemaURL(),
	}
}

func Scope(library instrumentation.Library, insts ...data.Instrument) data.Scope {
	return data.Scope{
		Library:     library,
		Instruments: insts,
	}
}

func CollectScope(t *testing.T, collectors []data.Collector, seq data.Sequence) []data.Instrument {
	t.Helper()
	var output data.Scope
	return CollectScopeReuse(t, collectors, seq, &output)
}

func CollectScopeReuse(t *testing.T, collectors []data.Collector, seq data.Sequence, output *data.Scope) []data.Instrument {
	t.Helper()
	output.Reset()

	for _, coll := range collectors {
		coll.Collect(seq, &output.Instruments)
	}
	return output.Instruments
}

func RequireEqualPoints(t *testing.T, output []data.Point, expected ...data.Point) {
	t.Helper()

	require.Equal(t, len(output), len(expected))

	cpy := make([]data.Point, len(expected))
	copy(cpy, expected)

	// Zero timestamps are skipped so as to bypass clock-based
	// testing in typical testing scenarios.
	for idx := range cpy {
		exp := &cpy[idx]
		out := &output[idx]

		if exp.Start.IsZero() {
			exp.Start = out.Start
		}

		if exp.End.IsZero() {
			exp.End = out.End
		}
	}

	require.ElementsMatch(t, cpy, output)
}

// RequireEqualMetrics checks that an output equals the expected
// output, where the points are taken to be unordered.  Instrument
// order is expected to match because the compiler preserves the order
// of views and instruments as they are compiled.
func RequireEqualMetrics(
	t *testing.T,
	output []data.Instrument,
	expected ...data.Instrument) {
	t.Helper()

	require.Equal(t, len(expected), len(output))

	for idx := range output {
		require.Equal(t, expected[idx].Descriptor, output[idx].Descriptor)

		RequireEqualPoints(t, output[idx].Points, expected[idx].Points...)
	}
}

func RequireEqualResourceMetrics(t *testing.T, output data.Metrics, expectRes *resource.Resource, expectScopes ...data.Scope) {
	t.Helper()
	require.Equal(t, expectRes, output.Resource)

	require.Equal(t, len(expectScopes), len(output.Scopes))

	for i, got := range output.Scopes {
		expect := expectScopes[i]

		require.Equal(t, expect.Library, got.Library)

		RequireEqualMetrics(t, got.Instruments, expect.Instruments...)
	}
}

func OTelErrors() *[]error {
	errors := new([]error)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		*errors = append(*errors, err)
	}))
	return errors
}
