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
package host

import (
	"context"
	"fmt"
	gonet "net"
	"os"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/stretchr/testify/assert"
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
func TestHostCPU(t *testing.T) {
	ctx := context.Background()
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	err := Start(WithMeterProvider(provider))
	require.NoError(t, err)

	// Note: we use a different library
	// ("github.com/shirou/gopsutil/v3/process") to verify process
	// CPU times computed from syscall.Getrusage().
	proc, err := process.NewProcess(int32(os.Getpid()))
	require.NoError(t, err)

	hostBefore, err := cpu.TimesWithContext(ctx, false)
	require.NoError(t, err)

	start := time.Now()
	for time.Since(start) < time.Second {
		// This has a mix of user and system time, so serves
		// the purpose of advancing both process and host,
		// user and system CPU usage.
		_, err = proc.TimesWithContext(ctx)
		require.NoError(t, err)
	}

	var data metricdata.ResourceMetrics
	err = reader.Collect(ctx, &data)
	require.NoError(t, err)
	require.Equal(t, 1, len(data.ScopeMetrics))

	hostUser := getMetric(data.ScopeMetrics[0].Metrics, "system.cpu.time", attrSlice(AttributeCPUTimeUser)[0])
	hostSystem := getMetric(data.ScopeMetrics[0].Metrics, "system.cpu.time", attrSlice(AttributeCPUTimeSystem)[0])

	hostAfter, err := cpu.TimesWithContext(ctx, false)
	require.NoError(t, err)

	// Validate host times:
	// Correct assumptions:
	require.Equal(t, 1, len(hostBefore))
	require.Equal(t, 1, len(hostAfter))
	// User times are in range
	require.LessOrEqual(t, hostBefore[0].User, hostUser)
	require.GreaterOrEqual(t, hostAfter[0].User, hostUser)
	// System times are in range
	require.LessOrEqual(t, hostBefore[0].System, hostSystem)
	require.GreaterOrEqual(t, hostAfter[0].System, hostSystem)
	// Ranges are not empty
	require.NotEqual(t, hostAfter[0].System, hostBefore[0].System)
	require.NotEqual(t, hostAfter[0].User, hostBefore[0].User)
	// TODO: We are not testing host "Other" nor "Idle" and
	// generally the specification hasn't been finalized, so
	// there's more to do.  Moreover, "Other" is not portable and
	// "Idle" may not advance on a fully loaded machine => both
	// are difficult to test.
}

func TestHostMemory(t *testing.T) {
	ctx := context.Background()
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	err := Start(WithMeterProvider(provider))
	require.NoError(t, err)

	vMem, err := mem.VirtualMemoryWithContext(ctx)
	require.NoError(t, err)

	var data metricdata.ResourceMetrics
	err = reader.Collect(ctx, &data)
	require.NoError(t, err)
	require.Equal(t, 1, len(data.ScopeMetrics))

	hostUsed := getMetric(data.ScopeMetrics[0].Metrics, "system.memory.usage", attrSlice(AttributeMemoryUsed)[0])
	assert.Greater(t, hostUsed, 0.0)
	assert.LessOrEqual(t, hostUsed, float64(vMem.Total))

	hostAvailable := getMetric(data.ScopeMetrics[0].Metrics, "system.memory.usage", attrSlice(AttributeMemoryAvailable)[0])
	assert.GreaterOrEqual(t, hostAvailable, 0.0)
	assert.Less(t, hostAvailable, float64(vMem.Total))

	hostUsedUtil := getMetric(data.ScopeMetrics[0].Metrics, "system.memory.utilization", attrSlice(AttributeMemoryUsed)[0])
	assert.Greater(t, hostUsedUtil, 0.0)
	assert.LessOrEqual(t, hostUsedUtil, 1.0)

	hostAvailableUtil := getMetric(data.ScopeMetrics[0].Metrics, "system.memory.utilization", attrSlice(AttributeMemoryAvailable)[0])
	assert.GreaterOrEqual(t, hostAvailableUtil, 0.0)
	assert.Less(t, hostAvailableUtil, 1.0)

	if hostUsed > hostAvailable {
		assert.Greater(t, hostUsedUtil, hostAvailableUtil)
	} else {
		assert.Less(t, hostUsedUtil, hostAvailableUtil)
	}
}

func sendBytes(t *testing.T, count int) error {
	conn1, err := gonet.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer conn1.Close()

	conn2, err := gonet.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer conn2.Close()

	data1 := make([]byte, 1000)
	data2 := make([]byte, 1000)
	for i := range data1 {
		data1[i] = byte(i)
	}

	for ; count > 0; count -= len(data1) {
		_, err = conn1.WriteTo(data1, conn2.LocalAddr())
		if err != nil {
			return err
		}
		_, readAddr, err := conn2.ReadFrom(data2)
		if err != nil {
			return err
		}

		require.Equal(t, "udp", readAddr.Network())
		require.Equal(t, conn1.LocalAddr().String(), readAddr.String())
	}

	return nil
}

func TestHostNetwork(t *testing.T) {
	ctx := context.Background()
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	err := Start(WithMeterProvider(provider))
	require.NoError(t, err)

	hostBefore, err := net.IOCountersWithContext(ctx, false)
	require.NoError(t, err)

	const howMuch = 10000
	err = sendBytes(t, howMuch)
	require.NoError(t, err)

	// As we are going to read the /proc file system for this info, sleep a while:
	require.Eventually(t, func() bool {
		hostAfter, err := net.IOCountersWithContext(ctx, false)
		require.NoError(t, err)

		return uint64(howMuch) <= hostAfter[0].BytesSent-hostBefore[0].BytesSent &&
			uint64(howMuch) <= hostAfter[0].BytesRecv-hostBefore[0].BytesRecv
	}, 30*time.Second, time.Second/2)

	var data metricdata.ResourceMetrics
	err = reader.Collect(ctx, &data)
	require.NoError(t, err)
	require.Equal(t, 1, len(data.ScopeMetrics))
	hostTransmit := getMetric(data.ScopeMetrics[0].Metrics, "system.network.io", attrSlice(AttributeNetworkTransmit)[0])
	hostReceive := getMetric(data.ScopeMetrics[0].Metrics, "system.network.io", attrSlice(AttributeNetworkReceive)[0])

	// Check that the recorded measurements reflect the same change:
	require.LessOrEqual(t, uint64(howMuch), uint64(hostTransmit)-hostBefore[0].BytesSent)
	require.LessOrEqual(t, uint64(howMuch), uint64(hostReceive)-hostBefore[0].BytesRecv)
}
