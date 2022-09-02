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

package runtime

import (
	"context"
	"runtime/metrics"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metrictest"
)

// prefix is mandatory for this library, however the "go." part is not.
const expectPrefix = "process.runtime.go."

var expectLib = metrictest.Scope{
	InstrumentationName:    "otel-launcher-go/runtime",
	InstrumentationVersion: "",
	SchemaURL:              "",
}

// TestBuiltinRuntimeMetrics tests the real output of the library to
// ensure expected prefix, instrumentation scope, and empty
// attributes.
func TestBuiltinRuntimeMetrics(t *testing.T) {
	provider, exp := metrictest.NewTestMeterProvider()

	err := Start(WithMeterProvider(provider))

	require.NoError(t, err)

	require.NoError(t, exp.Collect(context.Background()))

	// Counts are >1 for metrics that are totalized.
	expect := expectRuntimeMetrics
	allNames := map[string]int{}

	// Note: metrictest library lacks a way to distinguish
	// monotonic vs not or to test the unit. This will be fixed in
	// the new SDK, all the pieces untested here.
	for _, rec := range exp.Records {
		require.True(t, strings.HasPrefix(rec.InstrumentName, expectPrefix), "%s", rec.InstrumentName)
		name := rec.InstrumentName[len(expectPrefix):]

		require.Equal(t, expectLib, rec.InstrumentationLibrary)

		if expect[name] > 1 {
			require.Equal(t, 1, len(rec.Attributes))
		} else {
			require.Equal(t, 1, expect[name])
			require.Equal(t, []attribute.KeyValue(nil), rec.Attributes)
		}
		allNames[name]++
	}

	require.Equal(t, expect, allNames)
}

func makeTestCase() (allFunc, readFunc, map[string]map[string]metrics.Value) {
	// Note: the library provides no way to generate values, so use the
	// builtin library to get some.  Since we can't generate a Float64 value
	// we can't even test the Gauge logic in this package.
	ints := map[metrics.Value]bool{}

	real := metrics.All()
	realSamples := make([]metrics.Sample, len(real))
	for i := range real {
		realSamples[i].Name = real[i].Name
	}
	metrics.Read(realSamples)
	for i, rs := range realSamples {
		switch real[i].Kind {
		case metrics.KindUint64:
			ints[rs.Value] = true
		default:
			// Histograms and Floats are not tested.
			// The 1.19 runtime generates no Floats and
			// exports no test constructors.
		}
	}

	var allInts []metrics.Value

	for iv := range ints {
		allInts = append(allInts, iv)
	}

	af := func() []metrics.Description {
		return []metrics.Description{
			{
				Name:        "/cntr/things:things",
				Description: "a counter of things",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
			{
				Name:        "/updowncntr/things:things",
				Description: "an updowncounter of things",
				Kind:        metrics.KindUint64,
				Cumulative:  false,
			},
			{
				Name:        "/process/count:objects",
				Description: "a process counter of objects",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
			{
				Name:        "/process/count:bytes",
				Description: "a process counter of bytes",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
		}
	}
	mapping := map[string]metrics.Value{
		"/cntr/things:things":         allInts[0],
		"/updowncntr/things:things":   allInts[1],
		"/process/count:objects":      allInts[2],
		"/process/count:bytes":        allInts[3],
		"/waste/cycles/ocean:cycles":  allInts[4],
		"/waste/cycles/sea:cycles":    allInts[5],
		"/waste/cycles/lake:cycles":   allInts[6],
		"/waste/cycles/pond:cycles":   allInts[7],
		"/waste/cycles/puddle:cycles": allInts[8],
		"/waste/cycles/total:cycles":  allInts[9],
	}
	rf := func(samples []metrics.Sample) {
		for i := range samples {
			v, ok := mapping[samples[i].Name]
			if ok {
				samples[i].Value = v
			} else {
				panic("outcome uncertain")
			}
		}
	}
	return af, rf, map[string]map[string]metrics.Value{
		"cntr.things":           {"": allInts[0]},
		"updowncntr.things":     {"": allInts[1]},
		"process.count.objects": {"": allInts[2]},
		"process.count":         {"": allInts[3]},

		// This uses "cycles", one of the two known
		// multi-variate metrics as of go-1.19.
		"waste.cycles": {
			"ocean":  allInts[4],
			"sea":    allInts[5],
			"lake":   allInts[6],
			"pond":   allInts[7],
			"puddle": allInts[8],
		},
	}
}

// TestMetricTranslation validates the translation logic using
// synthetic metric names and values.
func TestMetricTranslation(t *testing.T) {
	provider, exp := metrictest.NewTestMeterProvider()

	af, rf, mapping := makeTestCase()
	br := newBuiltinRuntime(provider.Meter("test"), af, rf)
	br.register()

	expectRecords := 0
	for _, values := range mapping {
		expectRecords += len(values)
		if len(values) > 1 {
			// Counts the total
			expectRecords++
		}
	}

	require.NoError(t, exp.Collect(context.Background()))
	require.Equal(t, 10, expectRecords)

	for _, rec := range exp.Records {
		// Test the special cases are present always:

		require.True(t, strings.HasPrefix(rec.InstrumentName, expectPrefix), "%s", rec.InstrumentName)
		name := rec.InstrumentName[len(expectPrefix):]

		// Note: only int64 is tested, we have no way to
		// generate Float64 values and Float64Hist values are
		// not implemented for testing.
		m := mapping[name]
		if len(m) == 1 {
			require.Equal(t, mapping[name][""].Uint64(), uint64(rec.Sum.AsInt64()))

			// no attributes
			require.Equal(t, []attribute.KeyValue(nil), rec.Attributes)
		} else {
			require.Equal(t, 5, len(m))
			require.Equal(t, 1, len(rec.Attributes))
			require.Equal(t, rec.Attributes[0].Key, "class")
			feature := rec.Attributes[0].Value.AsString()
			require.Equal(t, mapping[name][feature].Uint64(), uint64(rec.Sum.AsInt64()))
		}
	}
}
