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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/encoding/prototext"

	"github.com/lightstep/otel-launcher-go/pipelines/test"
	"go.opentelemetry.io/otel/attribute"
	metricglobal "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func newTLSConfig() *tls.Config {
	certPool := x509.NewCertPool()

	ok := certPool.AppendCertsFromPEM([]byte(test.TestCARootCertificate))

	if !ok {
		panic("could not parse certificate authority certificate")
	}
	return &tls.Config{
		RootCAs:    certPool,
		ServerName: test.ServerName,
	}
}

func TestInsecureMetrics(t *testing.T) {
	server := test.NewServer(t)
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
		ReportingPeriod: "24h",
	})
	assert.NoError(t, err)

	meter := metricglobal.Meter("test-library")
	counter, err := meter.SyncFloat64().Counter("test-counter")
	assert.NoError(t, err)
	counter.Add(context.Background(), 1)

	shutdown()

	require.Equal(t, 0, len(server.TraceRequests()))
	require.Equal(t, 1, len(server.MetricsRequests()))
	txt, err := prototext.Marshal(server.MetricsRequests()[0])
	require.NoError(t, err)
	require.Contains(t, string(txt), "test-counter")
	require.Contains(t, string(txt), "test-r1")
	require.Contains(t, string(txt), "test-v1")
	require.Contains(t, string(txt), "test-library")

	require.Equal(t, []string{"test-value"}, server.MetricsMDs()[0]["test-header"])
}

func TestSecureMetrics(t *testing.T) {
	server := test.NewServer(t)
	defer server.Stop()

	shutdown, err := NewMetricsPipeline(PipelineConfig{
		Endpoint: fmt.Sprintf("%s:%d", test.ServerName, server.SecureMetricsPort),
		Headers: map[string]string{
			"test-header": "test-value",
		},
		Resource: resource.NewWithAttributes(
			semconv.SchemaURL,
			attribute.String("test-r1", "test-v1"),
		),
		ReportingPeriod: "24h",
		Credentials:     credentials.NewTLS(newTLSConfig()),
	})
	assert.NoError(t, err)

	meter := metricglobal.Meter("test-library")
	counter, err := meter.SyncFloat64().Counter("test-counter")
	assert.NoError(t, err)
	counter.Add(context.Background(), 1)

	shutdown()

	require.Equal(t, 0, len(server.TraceRequests()))
	require.Equal(t, 1, len(server.MetricsRequests()))
	txt, err := prototext.Marshal(server.MetricsRequests()[0])
	require.NoError(t, err)
	require.Contains(t, string(txt), "test-counter")
	require.Contains(t, string(txt), "test-r1")
	require.Contains(t, string(txt), "test-v1")
	require.Contains(t, string(txt), "test-library")

	require.Equal(t, []string{"test-value"}, server.MetricsMDs()[0]["test-header"])
}
