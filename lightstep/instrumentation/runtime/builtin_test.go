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
	"fmt"
	"runtime/metrics"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// TestMetricTranslation1 validates the translation logic using
// synthetic metric names and values.
func TestMetricTranslation1(t *testing.T) {
	testMetricTranslation(t, makeTestCase1)
}

// TestMetricTranslation2 is a more complex test than the first.
func TestMetricTranslation2(t *testing.T) {
	testMetricTranslation(t, makeTestCase2)
}

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

func makeTestCase1() (allFunc, readFunc, *builtinDescriptor, testExpectation) {
	allInts := makeAllInts()

	af := func() []metrics.Description {
		return []metrics.Description{
			{
				Name:       "/cntr/things:things",
				Kind:       metrics.KindUint64,
				Cumulative: true,
			},
			{
				Name:       "/updowncntr/things:things",
				Kind:       metrics.KindUint64,
				Cumulative: false,
			},
			{
				Name:       "/process/count:objects",
				Kind:       metrics.KindUint64,
				Cumulative: true,
			},
			{
				Name:       "/process/count:bytes",
				Kind:       metrics.KindUint64,
				Cumulative: true,
			},
			{
				Name:       "/waste/cycles/ocean:gc-cycles",
				Kind:       metrics.KindUint64,
				Cumulative: true,
			},
			{
				Name:       "/waste/cycles/sea:gc-cycles",
				Kind:       metrics.KindUint64,
				Cumulative: true,
			},
			{
				Name:       "/waste/cycles/lake:gc-cycles",
				Kind:       metrics.KindUint64,
				Cumulative: true,
			},
			{
				Name:       "/waste/cycles/pond:gc-cycles",
				Kind:       metrics.KindUint64,
				Cumulative: true,
			},
			{
				Name:       "/waste/cycles/puddle:gc-cycles",
				Kind:       metrics.KindUint64,
				Cumulative: true,
			},
			{
				Name:       "/waste/cycles/total:gc-cycles",
				Kind:       metrics.KindUint64,
				Cumulative: true,
			},
		}
	}
	mapping := testMapping{
		"/cntr/things:things":            allInts[0],
		"/updowncntr/things:things":      allInts[1],
		"/process/count:objects":         allInts[2],
		"/process/count:bytes":           allInts[3],
		"/waste/cycles/ocean:gc-cycles":  allInts[4],
		"/waste/cycles/sea:gc-cycles":    allInts[5],
		"/waste/cycles/lake:gc-cycles":   allInts[6],
		"/waste/cycles/pond:gc-cycles":   allInts[7],
		"/waste/cycles/puddle:gc-cycles": allInts[8],
		"/waste/cycles/total:gc-cycles":  allInts[9],
	}
	bd := newBuiltinDescriptor()
	bd.singleCounter("/cntr/things:things")
	bd.singleUpDownCounter("/updowncntr/things:things")
	bd.objectBytesCounter("/process/count:*")
	bd.classesCounter("/waste/cycles/*:gc-cycles",
		attribute.NewSet(classKey.String("ocean")),
		attribute.NewSet(classKey.String("sea")),
		attribute.NewSet(classKey.String("lake")),
		attribute.NewSet(classKey.String("pond")),
		attribute.NewSet(classKey.String("puddle")),
	)
	return af, mapping.read, bd, testExpectation{
		"cntr.things": testExpectMetric{
			unit: "{things}",
			desc: "/cntr/things:things from runtime/metrics",
			vals: map[attribute.Set]metrics.Value{
				emptySet: allInts[0],
			},
		},
		"updowncntr.things": testExpectMetric{
			unit: "{things}",
			desc: "/updowncntr/things:things from runtime/metrics",
			vals: map[attribute.Set]metrics.Value{
				emptySet: allInts[1],
			},
		},
		"process.count.objects": testExpectMetric{
			unit: "",
			desc: "/process/count:objects from runtime/metrics",
			vals: map[attribute.Set]metrics.Value{
				emptySet: allInts[2],
			},
		},
		"process.count": testExpectMetric{
			unit: unit.Bytes,
			desc: "/process/count:bytes from runtime/metrics",
			vals: map[attribute.Set]metrics.Value{
				emptySet: allInts[3],
			},
		},
		"waste.cycles": testExpectMetric{
			unit: "{gc-cycles}",
			desc: "/waste/cycles/*:gc-cycles from runtime/metrics",
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

func makeTestCase2() (allFunc, readFunc, *builtinDescriptor, testExpectation) {
	allInts := makeAllInts()

	af := func() []metrics.Description {
		return []metrics.Description{
			{
				Name:       "/objsize/classes/presos:bytes",
				Kind:       metrics.KindUint64,
				Cumulative: true,
			},
			{
				Name:       "/objsize/classes/sheets:bytes",
				Kind:       metrics.KindUint64,
				Cumulative: true,
			},
			{
				Name:       "/objsize/classes/docs/word:bytes",
				Kind:       metrics.KindUint64,
				Cumulative: true,
			},
			{
				Name:       "/objsize/classes/docs/pdf:bytes",
				Kind:       metrics.KindUint64,
				Cumulative: true,
			},
			{
				Name:       "/objsize/classes/docs/total:bytes",
				Kind:       metrics.KindUint64,
				Cumulative: true,
			},
			{
				Name:       "/objsize/classes/total:bytes",
				Kind:       metrics.KindUint64,
				Cumulative: true,
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
	bd := newBuiltinDescriptor()
	bd.classesUpDownCounter("/objsize/classes/*:bytes",
		attribute.NewSet(classKey.String("presos")),
		attribute.NewSet(classKey.String("sheets")),
		attribute.NewSet(classKey.String("docs"), subclassKey.String("word")),
		attribute.NewSet(classKey.String("docs"), subclassKey.String("pdf")),
	)
	return af, mapping.read, bd, testExpectation{
		"objsize.usage": testExpectMetric{
			unit: unit.Bytes,
			desc: "/objsize/classes/*:bytes from runtime/metrics",
			vals: map[attribute.Set]metrics.Value{
				attribute.NewSet(classKey.String("presos")):                           allInts[0],
				attribute.NewSet(classKey.String("sheets")):                           allInts[1],
				attribute.NewSet(classKey.String("docs"), subclassKey.String("word")): allInts[2],
				attribute.NewSet(classKey.String("docs"), subclassKey.String("pdf")):  allInts[3],
			},
		},
	}
}

func testMetricTranslation(t *testing.T, makeTestCase func() (allFunc, readFunc, *builtinDescriptor, testExpectation)) {
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	af, rf, desc, mapping := makeTestCase()
	br := newBuiltinRuntime(provider.Meter("test"), af, rf)
	err := br.register(desc)
	require.NoError(t, err)

	data, err := reader.Collect(context.Background())
	require.NoError(t, err)

	require.Equal(t, 1, len(data.ScopeMetrics))

	require.Equal(t, 1, len(data.ScopeMetrics))
	require.Equal(t, "test", data.ScopeMetrics[0].Scope.Name)
	for _, m := range data.ScopeMetrics[0].Metrics {
		fmt.Println(m)
	}
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
