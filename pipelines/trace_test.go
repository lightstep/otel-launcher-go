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
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/encoding/prototext"

	"github.com/lightstep/otel-launcher-go/pipelines/test"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func TestInsecureTrace(t *testing.T) {
	server := test.NewServer(t)
	defer server.Stop()

	shutdown, err := NewTracePipeline(PipelineConfig{
		Endpoint: fmt.Sprintf("%s:%d", test.ServerName, server.InsecureTracePort),
		Insecure: true,
		Headers: map[string]string{
			"test-header": "test-value",
		},
		Resource: resource.NewWithAttributes(
			semconv.SchemaURL,
			attribute.String("test-r1", "test-v1"),
		),
		ReportingPeriod: "24h",
		Propagators:     []string{"tracecontext", "baggage"},
	})
	assert.NoError(t, err)

	tracer := otel.Tracer("test-library")
	_, span := tracer.Start(context.Background(), "test-span")
	span.End()

	require.NoError(t, shutdown())

	require.Equal(t, 0, len(server.MetricsRequests()))
	require.Equal(t, 1, len(server.TraceRequests()))
	txt, err := prototext.Marshal(server.TraceRequests()[0])
	require.NoError(t, err)
	require.Contains(t, string(txt), "test-span")
	require.Contains(t, string(txt), "test-r1")
	require.Contains(t, string(txt), "test-v1")
	require.Contains(t, string(txt), "test-library")

	require.Equal(t, []string{"test-value"}, server.TraceMDs()[0]["test-header"])
}

func TestSecureTrace(t *testing.T) {
	server := test.NewServer(t)
	defer server.Stop()

	shutdown, err := NewTracePipeline(PipelineConfig{
		Endpoint: fmt.Sprintf("%s:%d", test.ServerName, server.SecureTracePort),
		Headers: map[string]string{
			"test-header": "test-value",
		},
		Resource: resource.NewWithAttributes(
			semconv.SchemaURL,
			attribute.String("test-r1", "test-v1"),
		),
		ReportingPeriod: "24h",
		Credentials:     credentials.NewTLS(newTLSConfig()),
		Propagators:     []string{"tracecontext", "baggage"},
	})
	assert.NoError(t, err)

	tracer := otel.Tracer("test-library")
	_, span := tracer.Start(context.Background(), "test-span")
	span.End()

	require.NoError(t, shutdown())

	require.Equal(t, 0, len(server.MetricsRequests()))
	require.Equal(t, 1, len(server.TraceRequests()))
	txt, err := prototext.Marshal(server.TraceRequests()[0])
	require.NoError(t, err)
	require.Contains(t, string(txt), "test-span")
	require.Contains(t, string(txt), "test-r1")
	require.Contains(t, string(txt), "test-v1")
	require.Contains(t, string(txt), "test-library")

	require.Equal(t, []string{"test-value"}, server.TraceMDs()[0]["test-header"])
}
