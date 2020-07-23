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

package launcher

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	apitrace "go.opentelemetry.io/otel/api/trace"
)

type testLogger struct {
	output []string
}

func (t *testLogger) addOutput(output string) {
	t.output = append(t.output, output)
}

func (t *testLogger) Fatalf(format string, v ...interface{}) {
	t.addOutput(fmt.Sprintf(format, v...))
}

func (t *testLogger) Debugf(format string, v ...interface{}) {
	t.addOutput(fmt.Sprintf(format, v...))
}

type testErrorHandler struct {
}

func (t *testErrorHandler) Handle(err error) {
	fmt.Printf("test error handler handled error: %v\n", err)
}

func TestInvalidServiceName(t *testing.T) {
	logger := &testLogger{output: []string{}}
	lsOtel := ConfigureOpentelemetry(WithLogger(logger))
	defer lsOtel.Shutdown()

	expected := "invalid configuration: service name missing"
	if !strings.Contains(logger.output[0], expected) {
		t.Errorf("\nString not found: %v\nIn: %v", expected, logger.output[0])
	}
}

func TestInvalidMissingAccessToken(t *testing.T) {
	logger := &testLogger{output: []string{}}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
	)
	defer lsOtel.Shutdown()

	expected := "invalid configuration: access token missing, must be set when reporting to ingest.lightstep.com:443"
	if !strings.Contains(logger.output[0], expected) {
		t.Errorf("\nString not found: %v\nIn: %v", expected, logger.output[0])
	}
}

func TestInvalidAccessToken(t *testing.T) {
	logger := &testLogger{output: []string{}}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithSpanExporterEndpoint("test123"),
		WithAccessToken("1234"),
	)
	defer lsOtel.Shutdown()

	expected := "invalid configuration: access token length incorrect. Ensure token is set correctly"
	if !strings.Contains(logger.output[0], expected) {
		t.Errorf("\nString not found: %v\nIn: %v", expected, logger.output[0])
	}
}

func TestValidConfig(t *testing.T) {
	logger := &testLogger{output: []string{}}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithAccessToken(strings.Repeat("1", 32)),
		WithErrorHandler(&testErrorHandler{}),
	)
	defer lsOtel.Shutdown()
	expected := 0
	if len(logger.output) > expected {
		t.Errorf("\nExpected: %v\ngot: %v", expected, len(logger.output))
	}

	lsOtel = ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithSpanExporterEndpoint("localhost:443"),
	)
	defer lsOtel.Shutdown()
	expected = 0
	if len(logger.output) > expected {
		t.Errorf("\nExpected: %v\ngot: %v", expected, len(logger.output))
	}
}

func TestDebugEnabled(t *testing.T) {
	logger := &testLogger{output: []string{}}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithAccessToken("access-token-123"),
		WithSpanExporterEndpoint("localhost:443"),
		WithLogLevel("debug"),
	)
	defer lsOtel.Shutdown()
	expected := "debug logging enabled"
	if expected != logger.output[0] {
		t.Errorf("\nExpected: %v\ngot: %v", expected, logger.output[0])
	}
}

func TestDefaultConfig(t *testing.T) {
	logger := &testLogger{}
	handler := &testErrorHandler{}
	config := newConfig(
		WithLogger(logger),
		WithErrorHandler(handler),
	)

	expected := LightstepConfig{
		ServiceName:                    "",
		ServiceVersion:                 "unknown",
		SpanExporterEndpoint:           "ingest.lightstep.com:443",
		SpanExporterEndpointInsecure:   false,
		MetricExporterEndpoint:         "ingest.lightstep.com:443/metrics",
		MetricExporterEndpointInsecure: false,
		AccessToken:                    "",
		LogLevel:                       "info",
		Propagators:                    []string{"b3"},
		logger:                         logger,
		errorHandler:                   handler,
	}
	assert.Equal(t, expected, config)
}

func TestEnvironmentVariables(t *testing.T) {
	setEnvironment()
	logger := &testLogger{}
	handler := &testErrorHandler{}
	config := newConfig(
		WithLogger(logger),
		WithErrorHandler(handler),
	)

	expected := LightstepConfig{
		ServiceName:                    "test-service-name",
		ServiceVersion:                 "test-service-version",
		SpanExporterEndpoint:           "satellite-url",
		SpanExporterEndpointInsecure:   true,
		MetricExporterEndpoint:         "metrics-url",
		MetricExporterEndpointInsecure: true,
		AccessToken:                    "token",
		LogLevel:                       "debug",
		Propagators:                    []string{"b3", "w3c"},
		logger:                         logger,
		errorHandler:                   handler,
	}
	unsetEnvironment()
	assert.Equal(t, expected, config)

}

func TestConfigurationOverrides(t *testing.T) {
	setEnvironment()
	logger := &testLogger{}
	handler := &testErrorHandler{}
	config := newConfig(
		WithServiceName("override-service-name"),
		WithServiceVersion("override-service-version"),
		WithAccessToken("override-access-token"),
		WithSpanExporterEndpoint("override-satellite-url"),
		WithSpanExporterInsecure(false),
		WithMetricExporterEndpoint("override-metrics-url"),
		WithMetricExporterInsecure(false),
		WithLogLevel("info"),
		WithLogger(logger),
		WithErrorHandler(handler),
		WithPropagators([]string{"b3"}),
	)

	expected := LightstepConfig{
		ServiceName:                    "override-service-name",
		ServiceVersion:                 "override-service-version",
		SpanExporterEndpoint:           "override-satellite-url",
		SpanExporterEndpointInsecure:   false,
		MetricExporterEndpoint:         "override-metrics-url",
		MetricExporterEndpointInsecure: false,
		AccessToken:                    "override-access-token",
		LogLevel:                       "info",
		Propagators:                    []string{"b3"},
		logger:                         logger,
		errorHandler:                   handler,
	}
	assert.Equal(t, expected, config)
}

func TestConfigurePropagators(t *testing.T) {
	unsetEnvironment()
	logger := &testLogger{}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithSpanExporterEndpoint("localhost:443"),
	)
	defer lsOtel.Shutdown()
	extractors := global.Propagators().HTTPExtractors()
	injectors := global.Propagators().HTTPInjectors()
	assert.Len(t, extractors, 1)
	assert.IsType(t, apitrace.B3{}, extractors[0])
	assert.Len(t, injectors, 1)
	assert.IsType(t, apitrace.B3{}, injectors[0])

	lsOtel = ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithSpanExporterEndpoint("localhost:443"),
		WithPropagators([]string{"b3", "cc"}),
	)
	defer lsOtel.Shutdown()
	extractors = global.Propagators().HTTPExtractors()
	injectors = global.Propagators().HTTPInjectors()
	assert.Len(t, extractors, 2)
	assert.IsType(t, apitrace.B3{}, extractors[0])
	assert.IsType(t, correlation.CorrelationContext{}, extractors[1])
	assert.Len(t, injectors, 2)
	assert.IsType(t, apitrace.B3{}, injectors[0])
	assert.IsType(t, correlation.CorrelationContext{}, injectors[1])

	logger = &testLogger{}
	lsOtel = ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithSpanExporterEndpoint("localhost:443"),
		WithPropagators([]string{"invalid"}),
	)
	defer lsOtel.Shutdown()

	expected := "invalid configuration: unsupported propagators. Supported options: b3,cc"
	if !strings.Contains(logger.output[0], expected) {
		t.Errorf("\nString not found: %v\nIn: %v", expected, logger.output[0])
	}
}

func TestConfigureResourcesLabels(t *testing.T) {
	config := LightstepConfig{
		ServiceName:    "test-service",
		ServiceVersion: "test-version",
		ResourceLabels: map[string]string{"label1": "value1", "label2": "value2"},
	}
	resource := newResource(&config)
	expected := []kv.KeyValue{
		kv.String("label1", "value1"),
		kv.String("label2", "value2"),
		kv.String("service.name", "test-service"),
		kv.String("service.version", "test-version"),
		kv.String("telemetry.sdk.language", "go"),
		kv.String("telemetry.sdk.name", "launcher"),
		kv.String("telemetry.sdk.version", "0.0.1"),
	}
	assert.Equal(t, expected, resource.Attributes())

	config = LightstepConfig{
		ServiceName:    "test-service",
		ServiceVersion: "test-version",
		ResourceLabels: map[string]string{"telemetry.sdk.language": "test-language"},
	}
	resource = newResource(&config)
	expected = []kv.KeyValue{
		kv.String("service.name", "test-service"),
		kv.String("service.version", "test-version"),
		kv.String("telemetry.sdk.language", "test-language"),
		kv.String("telemetry.sdk.name", "launcher"),
		kv.String("telemetry.sdk.version", "0.0.1"),
	}
	assert.Equal(t, expected, resource.Attributes())

	logger := &testLogger{}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithSpanExporterEndpoint("localhost:443"),
		WithSpanExporterInsecure(true),
		WithResourceLabels(map[string]string{"label1": "value1"}),
	)
	defer lsOtel.Shutdown()
}

func setEnvironment() {
	os.Setenv("LS_SERVICE_NAME", "test-service-name")
	os.Setenv("LS_SERVICE_VERSION", "test-service-version")
	os.Setenv("LS_ACCESS_TOKEN", "token")
	os.Setenv("OTEL_EXPORTER_OTLP_SPAN_ENDPOINT", "satellite-url")
	os.Setenv("OTEL_EXPORTER_OTLP_SPAN_INSECURE", "true")
	os.Setenv("OTEL_EXPORTER_OTLP_METRIC_ENDPOINT", "metrics-url")
	os.Setenv("OTEL_EXPORTER_OTLP_METRIC_INSECURE", "true")
	os.Setenv("OTEL_LOG_LEVEL", "debug")
	os.Setenv("OTEL_PROPAGATORS", "b3,w3c")
}

func unsetEnvironment() {
	vars := []string{
		"LS_SERVICE_NAME",
		"LS_SERVICE_VERSION",
		"LS_ACCESS_TOKEN",
		"OTEL_EXPORTER_OTLP_SPAN_ENDPOINT",
		"OTEL_EXPORTER_OTLP_SPAN_INSECURE",
		"OTEL_EXPORTER_OTLP_METRIC_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRIC_INSECURE",
		"OTEL_LOG_LEVEL",
		"OTEL_PROPAGATORS",
	}
	for _, envvar := range vars {
		os.Unsetenv(envvar)
	}
}

func TestMain(m *testing.M) {
	unsetEnvironment()
	os.Exit(m.Run())
}
