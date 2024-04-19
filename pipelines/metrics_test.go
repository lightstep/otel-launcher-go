// Copyright Lightstep Authors
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

package pipelines

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/prototext"

	"github.com/lightstep/otel-launcher-go/pipelines/test"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func newTLSClientSetting() *configtls.ClientConfig {
	return &configtls.ClientConfig{
		Config: configtls.Config{
			CAFile:   "test/testdata/caroot.crt",
			CertFile: "test/testdata/testserver.crt",
			KeyFile:  "test/testdata/testserver.key",
		},
		ServerName: test.ServerName,
	}
}

func testInsecureMetrics(t *testing.T, lightstepSDK, builtins bool) {
	var errors []error
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		errors = append(errors, err)
	}))

	server := test.NewServer()
	defer server.Stop()

	shutdown, err := NewMetricsPipeline(PipelineConfig{
		Endpoint: fmt.Sprintf("%s:%d", test.ServerName, server.InsecureMetricsPort),
		Insecure: true,
		Headers: map[string]string{
			"test-header": "test-value",
		},
		Resource: resource.NewWithAttributes(
			semconv.SchemaURL,
			attribute.String("test-r1", "test-v1"),
		),
		ReportingPeriod:         "24h",
		MetricsBuiltinsEnabled:  builtins,
		MetricsBuiltinLibraries: []string{"cputime:stable"},
	})
	assert.NoError(t, err)

	meter := otel.Meter("test-library")
	counter, err := meter.Float64Counter("test-counter")
	assert.NoError(t, err)
	counter.Add(context.Background(), 1)

	require.NoError(t, shutdown())

	require.Equal(t, 0, len(server.TraceRequests()))
	require.Equal(t, 1, len(server.MetricsRequests()))
	txt, err := prototext.Marshal(server.MetricsRequests()[0])
	require.NoError(t, err)
	require.Contains(t, string(txt), "test-counter")
	require.Contains(t, string(txt), "test-r1")
	require.Contains(t, string(txt), "test-v1")
	require.Contains(t, string(txt), "test-library")

	require.Equal(t, []string{"test-value"}, server.MetricsMDs()[0]["test-header"])

	if builtins {
		require.Contains(t, string(txt), "process.uptime")
	} else {
		require.NotContains(t, string(txt), "process.uptime")
	}

	// There should be no partial errors reported.
	require.Equal(t, 0, len(errors))
}

func testSecureMetrics(t *testing.T, lightstepSDK, builtins bool) {
	server := test.NewServer()
	defer server.Stop()

	cfg := PipelineConfig{
		Endpoint: fmt.Sprintf("%s:%d", test.ServerName, server.SecureMetricsPort),
		Headers: map[string]string{
			"test-header": "test-value",
		},
		Resource: resource.NewWithAttributes(
			semconv.SchemaURL,
			attribute.String("test-r1", "test-v1"),
		),
		ReportingPeriod:         "24h",
		MetricsBuiltinsEnabled:  builtins,
		MetricsBuiltinLibraries: []string{"cputime:stable"},
	}
	cfg.TLSSetting = newTLSClientSetting()

	shutdown, err := NewMetricsPipeline(cfg)
	require.NoError(t, err)

	meter := otel.Meter("test-library")
	counter, err := meter.Float64Counter("test-counter")
	assert.NoError(t, err)
	counter.Add(context.Background(), 1)

	require.NoError(t, shutdown())

	require.Equal(t, 0, len(server.TraceRequests()))
	require.Equal(t, 1, len(server.MetricsRequests()))
	txt, err := prototext.Marshal(server.MetricsRequests()[0])
	require.NoError(t, err)
	require.Contains(t, string(txt), "test-counter")
	require.Contains(t, string(txt), "test-r1")
	require.Contains(t, string(txt), "test-v1")
	require.Contains(t, string(txt), "test-library")

	if builtins {
		require.Contains(t, string(txt), "process.uptime")
	} else {
		require.NotContains(t, string(txt), "process.uptime")
	}

	require.Equal(t, []string{"test-value"}, server.MetricsMDs()[0]["test-header"])
}

// TODO: Fix the secure test for the TLSClientSettings-based setup.
// Failure is transport: authentication handshake failed: tls: failed to verify certificate: x509: “TestServer” certificate is not standards compliant\"
func TestSecureMetricsAltSDK(t *testing.T) {
	testSecureMetrics(t, true, true)
}

func TestSecureMetricsOldSDK(t *testing.T) {
	testSecureMetrics(t, false, true)
}

func TestInsecureMetricsAltSDK(t *testing.T) {
	testInsecureMetrics(t, true, true)
}

func TestInsecureMetricsOldSDK(t *testing.T) {
	testInsecureMetrics(t, false, true)
}

// TODO: Fix the secure test for the TLSClientSettings-based setup.
// Failure is transport: authentication handshake failed: tls: failed to verify certificate: x509: “TestServer” certificate is not standards compliant\"
func TestSecureMetricsAltSDKNoBuiltins(t *testing.T) {
	testSecureMetrics(t, true, false)
}

func TestSecureMetricsOldSDKNoBuiltins(t *testing.T) {
	testSecureMetrics(t, false, false)
}

func TestInsecureMetricsAltSDKNoBuiltins(t *testing.T) {
	testInsecureMetrics(t, true, false)
}

func TestInsecureMetricsOldSDKNoBuiltins(t *testing.T) {
	testInsecureMetrics(t, false, false)
}

func testBuiltinMetrics(t *testing.T, builtins []string, expectMetric string) {
	server := test.NewServer()
	defer server.Stop()

	shutdown, err := NewMetricsPipeline(PipelineConfig{
		Endpoint:                fmt.Sprintf("%s:%d", test.ServerName, server.InsecureMetricsPort),
		Insecure:                true,
		Headers:                 nil,
		Resource:                resource.Empty(),
		MetricsBuiltinsEnabled:  true,
		MetricsBuiltinLibraries: builtins,
	})
	assert.NoError(t, err)

	require.NoError(t, shutdown())

	require.Equal(t, 0, len(server.TraceRequests()))

	// Expect one request, even with no metrics reporting.
	require.Equal(t, 1, len(server.MetricsRequests()))

	txt, err := prototext.Marshal(server.MetricsRequests()[0])
	require.NoError(t, err)

	if expectMetric == "" {
		require.Equal(t, "resource_metrics:{resource:{}}", string(txt))
	} else {
		require.Contains(t, string(txt), expectMetric)
	}
}

func TestBuiltins(t *testing.T) {
	for _, test := range []struct {
		Builtins []string
		Expect   string
	}{
		// invalid entry, valid entry still starts
		{[]string{"invalid", "cputime"}, "process.uptime"},
		{[]string{"cputime", "invalid"}, "process.uptime"},
		{[]string{"cputime:stable"}, "process.uptime"},

		// empty version: success
		{[]string{"cputime:"}, "process.uptime"},

		// New runtime library
		{[]string{"runtime:stable"}, "process.runtime.go.gc.heap.frees.objects"},

		// New host library
		{[]string{"host:stable"}, "system.cpu.time"},

		// All libraries
		{[]string{"all:stable"}, "system.cpu.time"},
		{[]string{"all:stable"}, "process.uptime"},
		{[]string{"all:stable"}, "process.runtime.go.gc.heap.frees.objects"},
	} {
		testBuiltinMetrics(t, test.Builtins, test.Expect)
	}
}
