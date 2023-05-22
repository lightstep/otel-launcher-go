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
package cputime

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel/attribute"
	apimetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func attrSlice(opts []apimetric.ObserveOption) []attribute.KeyValue {
	cfg := apimetric.NewObserveConfig(opts)
	attrs := cfg.Attributes()
	return attrs.ToSlice()
}

func getMetric(metrics []metricdata.Metrics, name string, lbl attribute.KeyValue) float64 {
	for _, m := range metrics {
		fmt.Println(m.Name)
		if m.Name != name {
			continue
		}

		switch dt := m.Data.(type) {
		case metricdata.Gauge[int64]:
			if !lbl.Valid() {
				return float64(dt.DataPoints[0].Value)
			}
			for _, p := range dt.DataPoints {
				if val, ok := p.Attributes.Value(lbl.Key); ok && val.Emit() == lbl.Value.Emit() {
					return float64(p.Value)
				}
			}
		case metricdata.Gauge[float64]:
			if !lbl.Valid() {
				return dt.DataPoints[0].Value
			}
			for _, p := range dt.DataPoints {
				if val, ok := p.Attributes.Value(lbl.Key); ok && val.Emit() == lbl.Value.Emit() {
					return p.Value
				}
			}
		case metricdata.Sum[int64]:
			if !lbl.Valid() {
				return float64(dt.DataPoints[0].Value)
			}
			for _, p := range dt.DataPoints {
				if val, ok := p.Attributes.Value(lbl.Key); ok && val.Emit() == lbl.Value.Emit() {
					return float64(p.Value)
				}
			}
		case metricdata.Sum[float64]:
			if !lbl.Valid() {
				return dt.DataPoints[0].Value
			}
			for _, p := range dt.DataPoints {
				if val, ok := p.Attributes.Value(lbl.Key); ok && val.Emit() == lbl.Value.Emit() {
					return p.Value
				}
			}
		default:
			panic(fmt.Sprintf("invalid aggregation type: %v", dt))
		}
	}
	panic(fmt.Sprintf("Could not locate a metric in test output, name: %s, keyValue: %v", name, lbl))
}

func TestProcessCPU(t *testing.T) {
	ctx := context.Background()
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	err := Start(WithMeterProvider(provider))
	require.NoError(t, err)

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
		_, _, _ = c.getProcessTimes(ctx)
	}

	beforeUser, beforeSystem, _ := c.getProcessTimes(ctx)

	var data metricdata.ResourceMetrics
	err = reader.Collect(ctx, &data)
	require.NoError(t, err)
	require.Equal(t, 1, len(data.ScopeMetrics))

	processUser := getMetric(data.ScopeMetrics[0].Metrics, "process.cpu.time", attrSlice(AttributeCPUTimeUser)[0])
	processSystem := getMetric(data.ScopeMetrics[0].Metrics, "process.cpu.time", attrSlice(AttributeCPUTimeSystem)[0])

	afterUser, afterSystem, _ := c.getProcessTimes(ctx)

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

	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	c, err := newCputime(config{MeterProvider: provider})
	require.NoError(t, err)
	require.NoError(t, c.register())

	var data metricdata.ResourceMetrics
	err = reader.Collect(ctx, &data)
	require.NoError(t, err)
	require.Equal(t, 1, len(data.ScopeMetrics))

	procUptime := getMetric(data.ScopeMetrics[0].Metrics, "process.uptime", attribute.KeyValue{})

	require.LessOrEqual(t, expectUptime, procUptime)
}
