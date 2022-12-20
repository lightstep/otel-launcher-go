// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func makeAllInts() []metrics.Value {
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
	return allInts
}

type testMapping map[string]metrics.Value

func (m testMapping) read(samples []metrics.Sample) {
	for i := range samples {
		v, ok := m[samples[i].Name]
		if ok {
			samples[i].Value = v
		} else {
			panic("outcome uncertain")
		}
	}
}

type testExpectation map[string]testExpectMetric

type testExpectMetric struct {
	desc string
	unit unit.Unit
	vals map[attribute.Set]metrics.Value
}

func makeTestCase1() (allFunc, readFunc, testExpectation) {
	allInts := makeAllInts()

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
			{
				Name:        "/waste/cycles/ocean:cycles",
				Description: "blah blah",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
			{
				Name:        "/waste/cycles/sea:cycles",
				Description: "blah blah",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
			{
				Name:        "/waste/cycles/lake:cycles",
				Description: "blah blah",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
			{
				Name:        "/waste/cycles/pond:cycles",
				Description: "blah blah",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
			{
				Name:        "/waste/cycles/puddle:cycles",
				Description: "blah blah",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
			{
				Name:        "/waste/cycles/total:cycles",
				Description: "blah blah",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
		}
	}
	mapping := testMapping{
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
	return af, mapping.read, testExpectation{
		"cntr.things": testExpectMetric{
			unit: "{things}",
			desc: "runtime/metrics: /cntr/things:things",
			vals: map[attribute.Set]metrics.Value{
				emptySet: allInts[0],
			},
		},
		"updowncntr.things": testExpectMetric{
			unit: "{things}",
			desc: "runtime/metrics: /updowncntr/things:things",
			vals: map[attribute.Set]metrics.Value{
				emptySet: allInts[1],
			},
		},
		"process.count.objects": testExpectMetric{
			unit: "{objects}",
			desc: "runtime/metrics: /process/count:objects",
			vals: map[attribute.Set]metrics.Value{
				emptySet: allInts[2],
			},
		},
		"process.count": testExpectMetric{
			unit: unit.Bytes,
			desc: "runtime/metrics: /process/count:bytes",
			vals: map[attribute.Set]metrics.Value{
				emptySet: allInts[3],
			},
		},
		"waste.cycles": testExpectMetric{
			unit: "{cycles}",
			desc: "runtime/metrics: /waste/cycles/*:cycles",
			vals: map[attribute.Set]metrics.Value{
				attribute.NewSet(classKey.String("ocean")):  allInts[4],
				attribute.NewSet(classKey.String("sea")):    allInts[5],
				attribute.NewSet(classKey.String("lake")):   allInts[6],
				attribute.NewSet(classKey.String("pond")):   allInts[7],
				attribute.NewSet(classKey.String("puddle")): allInts[8],
			},
		},
	}
}

func makeTestCase2() (allFunc, readFunc, testExpectation) {
	allInts := makeAllInts()

	af := func() []metrics.Description {
		return []metrics.Description{
			{
				Name:        "/objsize/classes/presos:bytes",
				Description: "a counter of presos bytes",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
			{
				Name:        "/objsize/classes/sheets:bytes",
				Description: "a counter of sheets bytes",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
			{
				Name:        "/objsize/classes/docs/word:bytes",
				Description: "a counter of word doc bytes",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
			{
				Name:        "/objsize/classes/docs/pdf:bytes",
				Description: "a counter of word docs",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
			{
				Name:        "/objsize/classes/docs/total:bytes",
				Description: "a counter of all docs bytes",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
			{
				Name:        "/objsize/classes/total:bytes",
				Description: "a counter of all kinds things bytes",
				Kind:        metrics.KindUint64,
				Cumulative:  true,
			},
		}
	}
	mapping := testMapping{
		"/objsize/classes/presos:bytes":     allInts[0],
		"/objsize/classes/sheets:bytes":     allInts[1],
		"/objsize/classes/docs/word:bytes":  allInts[2],
		"/objsize/classes/docs/pdf:bytes":   allInts[3],
		"/objsize/classes/docs/total:bytes": allInts[4],
		"/objsize/classes/total:bytes":      allInts[5],
	}
	return af, mapping.read, testExpectation{
		"objsize.usage": testExpectMetric{
			unit: unit.Bytes,
			desc: "runtime/metrics: /objsize/classes/*:bytes",
			vals: map[attribute.Set]metrics.Value{
				attribute.NewSet(classKey.String("presos")):                           allInts[0],
				attribute.NewSet(classKey.String("sheets")):                           allInts[1],
				attribute.NewSet(classKey.String("docs"), subclassKey.String("word")): allInts[2],
				attribute.NewSet(classKey.String("docs"), subclassKey.String("pdf")):  allInts[3],
			},
		},
	}
}

// TestMetricTranslation1 validates the translation logic using
// synthetic metric names and values.
func TestMetricTranslation1(t *testing.T) {
	testMetricTranslation(t, makeTestCase1)
}

// TestMetricTranslation2 is a more complex test than the first.
func TestMetricTranslation2(t *testing.T) {
	testMetricTranslation(t, makeTestCase2)
}

func testMetricTranslation(t *testing.T, makeTestCase func() (allFunc, readFunc, testExpectation)) {
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	af, rf, mapping := makeTestCase()
	br := newBuiltinRuntime(provider.Meter("test"), af, rf)
	err := br.register()
	require.NoError(t, err)

	data, err := reader.Collect(context.Background())
	require.NoError(t, err)

	require.Equal(t, 1, len(data.ScopeMetrics))

	require.Equal(t, 1, len(data.ScopeMetrics))
	require.Equal(t, "test", data.ScopeMetrics[0].Scope.Name)
	require.Equal(t, len(mapping), len(data.ScopeMetrics[0].Metrics), "metrics count: %v", data.ScopeMetrics[0].Metrics)

	for _, inst := range data.ScopeMetrics[0].Metrics {
		// Test the special cases are present always:
		require.True(t, strings.HasPrefix(inst.Name, namePrefix+"."), "%s", inst.Name)
		name := inst.Name[len(namePrefix)+1:]

		// Note: only int64 values are tested, as we have no
		// way to generate float64 values w/o an
		// runtime/metrics supporting a metric to generate
		// these values.
		exm := mapping[name]

		require.Equal(t, exm.desc, inst.Description)
		require.Equal(t, exm.unit, inst.Unit)

		require.Equal(t, len(exm.vals), len(inst.Data.(metricdata.Sum[int64]).DataPoints), "points count: %v != %v", inst.Data.(metricdata.Sum[int64]).DataPoints, exm.vals)

		for _, point := range inst.Data.(metricdata.Sum[int64]).DataPoints {
			lookup, ok := exm.vals[point.Attributes]
			require.True(t, ok, "lookup failed: %v", exm.vals, point.Attributes)
			require.Equal(t, lookup.Uint64(), uint64(point.Value))
		}
	}
}
