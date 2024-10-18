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

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/gauge"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exemplar"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func Descriptor(name string, ik sdkinstrument.Kind, nk number.Kind) sdkinstrument.Descriptor {
	return sdkinstrument.NewDescriptor(name, ik, nk, "", "")
}

func DescriptorDescUnit(name string, ik sdkinstrument.Kind, nk number.Kind, desc, unit string) sdkinstrument.Descriptor {
	return sdkinstrument.NewDescriptor(name, ik, nk, desc, unit)
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

func PointEx(start, end time.Time, agg aggregation.Aggregation, tempo aggregation.Temporality, kvs []attribute.KeyValue, exs ...aggregator.WeightedExemplarBits) data.Point {
	attrs := attribute.NewSet(kvs...)
	return data.Point{
		Start:       start,
		End:         end,
		Attributes:  attrs,
		Aggregation: agg,
		Temporality: tempo,
		Exemplars:   exs,
	}
}

func Instrument(desc sdkinstrument.Descriptor, points ...data.Point) data.Instrument {
	return data.Instrument{
		Descriptor: desc,
		Points:     points,
	}
}

func Library(name string, opts ...metric.MeterOption) instrumentation.Scope {
	cfg := metric.NewMeterConfig(opts...)
	return instrumentation.Scope{
		Name:      name,
		Version:   cfg.InstrumentationVersion(),
		SchemaURL: cfg.SchemaURL(),
	}
}

func Scope(library instrumentation.Scope, insts ...data.Instrument) data.Scope {
	return data.Scope{
		Library:     library,
		Instruments: insts,
	}
}

func Metrics(res *resource.Resource, scopes ...data.Scope) data.Metrics {
	return data.Metrics{
		Resource: res,
		Scopes:   scopes,
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

	require.Equal(t, len(expected), len(output), "points have different length")

	cpy := make([]data.Point, len(expected))
	copy(cpy, expected)

	// If the expectations have zero timestamps, the output
	// timestamps are zeroed so they will match exactly.  Gauge
	// sequence numbers are set to match test conditions.
	for idx := range cpy {
		exp := &cpy[idx]
		out := &output[idx]

		if exp.Start.IsZero() {
			out.Start = exp.Start
		}

		if exp.End.IsZero() {
			out.End = exp.End
		}

		// Special case for testing exemplars: the exemplar aggregator is full of
		// intermediate state, so we replace the aggregation with the
		// underlying aggregation for the purposes of testing its correct
		// value w/o testing sampler state.
		if unwr, ok := out.Aggregation.(exemplar.Unwrapper); ok {
			out.Aggregation = unwr.Unwrap()
		}

		if outig, ok := out.Aggregation.(*gauge.Int64); ok {
			outig.SetSequenceForTesting()
		}
		if outfg, ok := out.Aggregation.(*gauge.Float64); ok {
			outfg.SetSequenceForTesting()
		}
	}

	var cpyEx [][]aggregator.WeightedExemplarBits
	var outEx [][]aggregator.WeightedExemplarBits

	for i := range cpy {
		if len(cpy[i].Exemplars) != 0 {
			cpyEx = append(cpyEx, cpy[i].Exemplars)
			cpy[i].Exemplars = nil
		}
	}
	for i := range output {
		if len(output[i].Exemplars) != 0 {
			outEx = append(outEx, output[i].Exemplars)
			output[i].Exemplars = nil
		}
	}

	require.ElementsMatch(t, cpy, output)

	require.Equal(t, len(cpyEx), len(outEx))

	for i := range cpyEx {
		if len(cpyEx[i]) != len(outEx[i]) {
			require.ElementsMatch(t, cpyEx, outEx)
			continue
		}

		// As with the above, we zero timestamps if the expectation
		// is zero, to skip the timestamp test.
		for j := range cpyEx[i] {
			if cpyEx[i][j].Time.IsZero() {
				outEx[i][j].Time = time.Time{}
			}
		}

		require.ElementsMatch(t, cpyEx, outEx)
	}
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

type FakeSpanData struct {
	t, s byte
	noop.Span
}

func FakeSpan(t, s byte) trace.Span {
	return &FakeSpanData{
		t: t,
		s: s,
	}
}

func (f *FakeSpanData) SpanContext() trace.SpanContext {
	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    [16]byte{f.t, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		SpanID:     [8]byte{f.s, 0, 0, 0, 0, 0, 0, 0},
		TraceFlags: 0x1,
	})
}
