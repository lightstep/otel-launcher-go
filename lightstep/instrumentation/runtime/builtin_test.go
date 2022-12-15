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
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// prefix is mandatory for this library, however the "go." part is not.
const expectPrefix = "process.runtime.go."

var expectScope = instrumentation.Scope{
	Name: "otel-launcher-go/runtime",
}

// TestBuiltinRuntimeMetrics tests the real output of the library to
// ensure expected prefix, instrumentation scope, and empty
// attributes.
func TestBuiltinRuntimeMetrics(t *testing.T) {
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	err := Start(WithMeterProvider(provider))
	require.NoError(t, err)

	data, err := reader.Collect(context.Background())
	require.NoError(t, err)

	require.Equal(t, 1, len(data.ScopeMetrics))
	require.Equal(t, expectScope, data.ScopeMetrics[0].Scope)

	expect := expectRuntimeMetrics
	allNames := map[string]int{}

	// Note: metrictest library lacks a way to distinguish
	// monotonic vs not or to test the unit. This will be fixed in
	// the new SDK, all the pieces untested here.
	for _, inst := range data.ScopeMetrics[0].Metrics {
		require.True(t, strings.HasPrefix(inst.Name, expectPrefix), "%s", inst.Name)
		name := inst.Name[len(expectPrefix):]
		var attrs attribute.Set
		switch dt := inst.Data.(type) {
		case metricdata.Gauge[int64]:
			require.Equal(t, 1, len(dt.DataPoints))
			attrs = dt.DataPoints[0].Attributes
		case metricdata.Gauge[float64]:
			require.Equal(t, 1, len(dt.DataPoints))
			attrs = dt.DataPoints[0].Attributes
		case metricdata.Sum[int64]:
			require.Equal(t, 1, len(dt.DataPoints))
			attrs = dt.DataPoints[0].Attributes
		case metricdata.Sum[float64]:
			require.Equal(t, 1, len(dt.DataPoints))
			attrs = dt.DataPoints[0].Attributes
		}

		if expect[name] > 1 {
			require.Equal(t, 1, attrs.Len())
		} else {
			require.Equal(t, 1, expect[name], "for %v", inst.Name)
			require.Equal(t, 0, attrs.Len())
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
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	af, rf, mapping := makeTestCase()
	br := newBuiltinRuntime(provider.Meter("test"), af, rf)
	err := br.register()
	require.NoError(t, err)

	expectRecords := 0
	for _, values := range mapping {
		expectRecords += len(values)
		if len(values) > 1 {
			// Counts the total
			expectRecords++
		}
	}

	data, err := reader.Collect(context.Background())
	require.NoError(t, err)
	require.Equal(t, 10, expectRecords)

	require.Equal(t, 1, len(data.ScopeMetrics))
	require.Equal(t, "test", data.ScopeMetrics[0].Scope.Name)

	for _, inst := range data.ScopeMetrics[0].Metrics {
		// Test the special cases are present always:

		require.True(t, strings.HasPrefix(inst.Name, expectPrefix), "%s", inst.Name)
		name := inst.Name[len(expectPrefix):]

		require.Equal(t, 1, len(inst.Data.(metricdata.Sum[int64]).DataPoints))

		sum := inst.Data.(metricdata.Sum[int64]).DataPoints[0].Value
		attrs := inst.Data.(metricdata.Sum[int64]).DataPoints[0].Attributes

		// Note: only int64 is tested, we have no way to
		// generate Float64 values and Float64Hist values are
		// not implemented for testing.
		m := mapping[name]
		if len(m) == 1 {
			require.Equal(t, mapping[name][""].Uint64(), uint64(sum))

			// no attributes
			require.Equal(t, 0, attrs.Len())
		} else {
			require.Equal(t, 5, len(m))
			require.Equal(t, 1, attrs.Len())
			require.Equal(t, attrs.ToSlice()[0].Key, "class")
			feature := attrs.ToSlice()[0].Value.AsString()
			require.Equal(t, mapping[name][feature].Uint64(), uint64(sum))
		}
	}
}
