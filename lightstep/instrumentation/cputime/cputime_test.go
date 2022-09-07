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

package cputime

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/metrictest"
)

func getMetric(exp *metrictest.Exporter, name string, lbl attribute.KeyValue) float64 {
	for _, r := range exp.GetRecords() {
		if r.InstrumentName != name {
			continue
		}

		if lbl.Key != "" {
			foundAttribute := false
			for _, haveLabel := range r.Attributes {
				if haveLabel != lbl {
					continue
				}
				foundAttribute = true
				break
			}
			if !foundAttribute {
				continue
			}
		}

		switch r.AggregationKind {
		case aggregation.SumKind, aggregation.HistogramKind:
			return r.Sum.CoerceToFloat64(r.NumberKind)
		case aggregation.LastValueKind:
			return r.LastValue.CoerceToFloat64(r.NumberKind)
		default:
			panic(fmt.Sprintf("invalid aggregation type: %v", r.AggregationKind))
		}
	}
	panic("Could not locate a metric in test output")
}

func TestProcessCPU(t *testing.T) {
	provider, exp := metrictest.NewTestMeterProvider()
	err := Start(
		WithMeterProvider(provider),
	)
	require.NoError(t, err)

	ctx := context.Background()

	// This is a second copy of the same source of information.
	// We ultimately have to trust the information source, the
	// test here is to be sure the information is correctly
	// translated into metrics.
	c, err := newCputime(config{
		MeterProvider: provider,
	})
	require.NoError(t, err)

	start := time.Now()
	for time.Since(start) < time.Second {
		// This has a mix of user and system time, so serves
		// the purpose of advancing both process and host,
		// user and system CPU usage.
		_, _, _, _ = c.getProcessTimes(ctx)
	}

	beforeUser, beforeSystem, _, _ := c.getProcessTimes(ctx)

	require.NoError(t, exp.Collect(ctx))

	processUser := getMetric(exp, "process.cpu.time", AttributeCPUTimeUser[0])
	processSystem := getMetric(exp, "process.cpu.time", AttributeCPUTimeSystem[0])

	afterUser, afterSystem, _, _ := c.getProcessTimes(ctx)

	// Validate process times:
	// User times are in range
	require.LessOrEqual(t, beforeUser, processUser)
	require.GreaterOrEqual(t, afterUser, processUser)
	// System times are in range
	require.LessOrEqual(t, beforeSystem, processSystem)
	require.GreaterOrEqual(t, afterSystem, processSystem)
}

func TestProcessUptime(t *testing.T) {
	ctx := context.Background()
	y2k, err := time.Parse(time.RFC3339, "2000-01-01T00:00:00Z")
	require.NoError(t, err)
	expectUptime := time.Since(y2k).Seconds()

	var save time.Time
	processStartTime, save = y2k, processStartTime
	defer func() {
		processStartTime = save
	}()

	provider, exp := metrictest.NewTestMeterProvider()
	c, err := newCputime(config{MeterProvider: provider})
	require.NoError(t, err)
	require.NoError(t, c.register())

	require.NoError(t, exp.Collect(ctx))
	procUptime := getMetric(exp, "process.uptime", attribute.KeyValue{})

	require.LessOrEqual(t, expectUptime, procUptime)
}

func TestProcessGCCPUTime(t *testing.T) {
	ctx := context.Background()

	provider, exp := metrictest.NewTestMeterProvider()
	c, err := newCputime(config{
		MeterProvider: provider,
	})
	require.NoError(t, err)
	require.NoError(t, c.register())

	require.NoError(t, exp.Collect(ctx))
	initialUtime := getMetric(exp, "process.cpu.time", AttributeCPUTimeUser[0])
	initialStime := getMetric(exp, "process.cpu.time", AttributeCPUTimeSystem[0])
	initialGCtime := getMetric(exp, "process.runtime.go.gc.cpu.time", attribute.KeyValue{})

	// Make garbage
	for i := 0; i < 2; i++ {
		var garbage []struct{}
		for start := time.Now(); time.Since(start) < time.Second/16; {
			garbage = append(garbage, struct{}{})
		}
		require.Less(t, 0, len(garbage))
		runtime.GC()

		require.NoError(t, exp.Collect(ctx))
		utime := -initialUtime + getMetric(exp, "process.cpu.time", AttributeCPUTimeUser[0])
		stime := -initialStime + getMetric(exp, "process.cpu.time", AttributeCPUTimeSystem[0])
		gctime := -initialGCtime + getMetric(exp, "process.runtime.go.gc.cpu.time", attribute.KeyValue{})

		require.LessOrEqual(t, gctime, utime+stime)
	}
}
