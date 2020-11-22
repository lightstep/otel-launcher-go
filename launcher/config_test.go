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
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/sdk/resource"
)

const (
	expectedAccessTokenLengthError  = "invalid configuration: access token length incorrect. Ensure token is set correctly"
	expectedAccessTokenMissingError = "invalid configuration: access token missing, must be set when reporting to ingest.lightstep.com"
	expectedTracingDisabledMessage  = "tracing is disabled by configuration: no endpoint set"
	expectedMetricsDisabledMessage  = "metrics are disabled by configuration: no endpoint set"
)

type testLogger struct {
	output []string
}

func (logger *testLogger) addOutput(output string) {
	logger.output = append(logger.output, output)
}

func (logger *testLogger) Fatalf(format string, v ...interface{}) {
	logger.addOutput(fmt.Sprintf(format, v...))
}

func (logger *testLogger) Debugf(format string, v ...interface{}) {
	logger.addOutput(fmt.Sprintf(format, v...))
}

func (logger *testLogger) requireContains(t *testing.T, expected string) {
	t.Helper()
	for _, output := range logger.output {
		if strings.Contains(output, expected) {
			return
		}
	}

	t.Errorf("\nString unexpectedly not found: %v\nIn: %v", expected, logger.output)
}

func (logger *testLogger) requireNotContains(t *testing.T, expected string) {
	t.Helper()
	for _, output := range logger.output {
		if strings.Contains(output, expected) {
			t.Errorf("\nString unexpectedly found: %v\nIn: %v", expected, logger.output)
			return
		}
	}
}

func (logger *testLogger) reset() {
	logger.output = nil
}

type testErrorHandler struct {
}

func (t *testErrorHandler) Handle(err error) {
	fmt.Printf("test error handler handled error: %v\n", err)
}

func fakeAccessToken() string {
	return strings.Repeat("1", 32)
}

func TestInvalidServiceName(t *testing.T) {
	logger := &testLogger{}
	lsOtel := ConfigureOpentelemetry(WithLogger(logger))
	defer lsOtel.Shutdown()

	expected := "invalid configuration: service name missing"
	logger.requireContains(t, expected)
}

func testInvalidMissingAccessToken(t *testing.T, opts ...Option) {
	logger := &testLogger{}
	lsOtel := ConfigureOpentelemetry(
		append(opts,
			WithLogger(logger),
			WithServiceName("test-service"),
		)...,
	)
	defer lsOtel.Shutdown()

	logger.requireContains(t, expectedAccessTokenMissingError)
}

func TestInvalidMissingDefaultAccessToken(t *testing.T) {
	testInvalidMissingAccessToken(
		t,
		WithAccessToken(""),
	)
}

func TestInvalidTraceDefaultAccessToken(t *testing.T) {
	testInvalidMissingAccessToken(t,
		WithAccessToken(""),
		WithSpanExporterEndpoint(DefaultSpanExporterEndpoint),
		WithMetricExporterEndpoint("127.0.0.1:4000"),
	)
}

func TestInvalidMetricDefaultAccessToken(t *testing.T) {
	testInvalidMissingAccessToken(t,
		WithAccessToken(""),
		WithSpanExporterEndpoint("127.0.0.1:4000"),
		WithMetricExporterEndpoint(DefaultMetricExporterEndpoint))
}

func testInvalidAccessToken(t *testing.T, opts ...Option) {
	logger := &testLogger{}
	lsOtel := ConfigureOpentelemetry(
		append(opts,
			WithLogger(logger),
			WithServiceName("test-service"),
		)...,
	)
	defer lsOtel.Shutdown()

	logger.requireContains(t, expectedAccessTokenLengthError)
}

func TestInvalidTraceAccessTokenLength(t *testing.T) {
	testInvalidAccessToken(t,
		WithSpanExporterEndpoint("127.0.0.1:4000"),
		WithAccessToken("1234"),
	)
}

func TestInvalidMetricAccessTokenLength(t *testing.T) {
	testInvalidAccessToken(t,
		WithSpanExporterEndpoint(""),
		WithMetricExporterEndpoint("127.0.0.1:4000"),
		WithAccessToken("1234"),
	)
}

func testEndpointDisabled(t *testing.T, expected string, opts ...Option) {
	logger := &testLogger{}
	lsOtel := ConfigureOpentelemetry(
		append(opts,
			WithLogger(logger),
			WithServiceName("test-service"),
		)...,
	)
	defer lsOtel.Shutdown()

	logger.requireNotContains(t, expectedAccessTokenMissingError)
	logger.requireContains(t, expected)
}

func TestTraceEndpointDisabled(t *testing.T) {
	testEndpointDisabled(
		t,
		expectedTracingDisabledMessage,
		WithAccessToken(fakeAccessToken()),
		WithSpanExporterEndpoint(""),
	)
}

func TestMetricEndpointDisabled(t *testing.T) {
	testEndpointDisabled(
		t,
		expectedMetricsDisabledMessage,
		WithAccessToken(fakeAccessToken()),
		WithMetricExporterEndpoint(""),
	)
}

func TestValidConfig(t *testing.T) {
	logger := &testLogger{}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithAccessToken(fakeAccessToken()),
		WithErrorHandler(&testErrorHandler{}),
	)
	defer lsOtel.Shutdown()

	logger.requireContains(t, expectedMetricsDisabledMessage)
	logger.reset()

	lsOtel = ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithMetricExporterEndpoint("localhost:443"),
		WithSpanExporterEndpoint("localhost:443"),
	)
	defer lsOtel.Shutdown()

	if len(logger.output) > 0 {
		t.Errorf("\nExpected: no logs\ngot: %v", logger.output)
	}
}

func TestInvalidEnvironment(t *testing.T) {
	os.Setenv("OTEL_EXPORTER_OTLP_METRIC_INSECURE", "bleargh")

	logger := &testLogger{}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
	)
	defer lsOtel.Shutdown()

	logger.requireContains(t, "environment error")
	unsetEnvironment()
}

func TestInvalidMetricsPushIntervalEnv(t *testing.T) {
	os.Setenv("OTEL_EXPORTER_OTLP_METRIC_PERIOD", "300million")

	logger := &testLogger{}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithSpanExporterEndpoint("127.0.0.1:4000"),
		WithMetricExporterEndpoint("127.0.0.1:4000"),
	)
	defer lsOtel.Shutdown()

	logger.requireContains(t, "setup error: invalid metric reporting period")
	unsetEnvironment()
}

func TestInvalidMetricsPushIntervalConfig(t *testing.T) {
	logger := &testLogger{}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithSpanExporterEndpoint("127.0.0.1:4000"),
		WithMetricExporterEndpoint("127.0.0.1:4000"),
		WithMetricReportingPeriod(-time.Second),
	)
	defer lsOtel.Shutdown()

	logger.requireContains(t, "setup error: invalid metric reporting period")
	unsetEnvironment()
}

func TestDebugEnabled(t *testing.T) {
	logger := &testLogger{}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithAccessToken("access-token-123"),
		WithSpanExporterEndpoint("localhost:443"),
		WithLogLevel("debug"),
		WithResourceAttributes(map[string]string{
			"attr1":     "val1",
			"host.name": "host456",
		}),
	)
	defer lsOtel.Shutdown()
	output := strings.Join(logger.output[:], ",")
	assert.Contains(t, output, "debug logging enabled")
	assert.Contains(t, output, "test-service")
	assert.Contains(t, output, "access-token-123")
	assert.Contains(t, output, "localhost:443")
	assert.Contains(t, output, "attr1")
	assert.Contains(t, output, "val1")
	assert.Contains(t, output, "host.name")
	assert.Contains(t, output, "host456")
}

func TestDefaultConfig(t *testing.T) {
	logger := &testLogger{}
	handler := &testErrorHandler{}
	config := newConfig(
		WithLogger(logger),
		WithErrorHandler(handler),
	)

	attributes := []label.KeyValue{
		label.String("host.name", host()),
		label.String("service.version", "unknown"),
		label.String("telemetry.sdk.name", "launcher"),
		label.String("telemetry.sdk.language", "go"),
		label.String("telemetry.sdk.version", version),
	}

	expected := Config{
		ServiceName:                    "",
		ServiceVersion:                 "unknown",
		SpanExporterEndpoint:           "ingest.lightstep.com:443",
		SpanExporterEndpointInsecure:   false,
		MetricExporterEndpoint:         "",
		MetricExporterEndpointInsecure: false,
		MetricReportingPeriod:          "30s",
		AccessToken:                    "",
		LogLevel:                       "info",
		Propagators:                    []string{"b3"},
		Resource:                       resource.NewWithAttributes(attributes...),
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

	attributes := []label.KeyValue{
		label.String("host.name", host()),
		label.String("service.name", "test-service-name"),
		label.String("service.version", "test-service-version"),
		label.String("telemetry.sdk.name", "launcher"),
		label.String("telemetry.sdk.language", "go"),
		label.String("telemetry.sdk.version", version),
	}

	expected := Config{
		ServiceName:                    "test-service-name",
		ServiceVersion:                 "test-service-version",
		SpanExporterEndpoint:           "satellite-url",
		SpanExporterEndpointInsecure:   true,
		MetricExporterEndpoint:         "metrics-url",
		MetricExporterEndpointInsecure: true,
		MetricReportingPeriod:          "30s",
		AccessToken:                    "token",
		LogLevel:                       "debug",
		Propagators:                    []string{"b3", "w3c"},
		Resource:                       resource.NewWithAttributes(attributes...),
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

	attributes := []label.KeyValue{
		label.String("host.name", host()),
		label.String("service.name", "override-service-name"),
		label.String("service.version", "override-service-version"),
		label.String("telemetry.sdk.name", "launcher"),
		label.String("telemetry.sdk.language", "go"),
		label.String("telemetry.sdk.version", version),
	}

	expected := Config{
		ServiceName:                    "override-service-name",
		ServiceVersion:                 "override-service-version",
		SpanExporterEndpoint:           "override-satellite-url",
		SpanExporterEndpointInsecure:   false,
		MetricExporterEndpoint:         "override-metrics-url",
		MetricExporterEndpointInsecure: false,
		MetricReportingPeriod:          "30s",
		AccessToken:                    "override-access-token",
		LogLevel:                       "info",
		Propagators:                    []string{"b3"},
		Resource:                       resource.NewWithAttributes(attributes...),
		logger:                         logger,
		errorHandler:                   handler,
	}
	assert.Equal(t, expected, config)
}

type TestCarrier struct {
	values map[string]string
}

func (t TestCarrier) Get(key string) string {
	return t.values[key]
}

func (t TestCarrier) Set(key string, value string) {
	t.values[key] = value
}

func TestConfigurePropagators(t *testing.T) {
	ctx := baggage.ContextWithValues(context.Background(),
		label.String("keyone", "foo1"),
		label.String("keytwo", "bar1"),
	)
	unsetEnvironment()
	logger := &testLogger{}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithSpanExporterEndpoint("localhost:443"),
	)
	defer lsOtel.Shutdown()
	ctx, finish := otel.Tracer("ex.com/basic").Start(ctx, "foo")
	defer finish.End()
	carrier := TestCarrier{values: map[string]string{}}
	prop := otel.GetTextMapPropagator()
	prop.Inject(ctx, carrier)
	assert.Greater(t, len(carrier.Get("x-b3-traceid")), 0)
	assert.Equal(t, "", carrier.Get("baggage"))
	assert.Equal(t, len(carrier.Get("traceparent")), 0)

	lsOtel = ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithSpanExporterEndpoint("localhost:443"),
		WithPropagators([]string{"b3", "baggage", "tracecontext"}),
	)
	defer lsOtel.Shutdown()
	carrier = TestCarrier{values: map[string]string{}}
	prop = otel.GetTextMapPropagator()
	prop.Inject(ctx, carrier)
	assert.Greater(t, len(carrier.Get("x-b3-traceid")), 0)
	assert.Contains(t, carrier.Get("baggage"), "keytwo=bar1")
	assert.Contains(t, carrier.Get("baggage"), "keyone=foo1")
	assert.Greater(t, len(carrier.Get("traceparent")), 0)

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

func host() string {
	host, _ := os.Hostname()
	return host
}

func TestConfigureResourcesAttributes(t *testing.T) {
	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "label1=value1,label2=value2")
	config := Config{
		ServiceName:    "test-service",
		ServiceVersion: "test-version",
	}
	resource := newResource(&config)
	expected := []label.KeyValue{
		label.String("host.name", host()),
		label.String("label1", "value1"),
		label.String("label2", "value2"),
		label.String("service.name", "test-service"),
		label.String("service.version", "test-version"),
		label.String("telemetry.sdk.language", "go"),
		label.String("telemetry.sdk.name", "launcher"),
		label.String("telemetry.sdk.version", version),
	}
	assert.Equal(t, expected, resource.Attributes())

	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "telemetry.sdk.language=test-language")
	config = Config{
		ServiceName:    "test-service",
		ServiceVersion: "test-version",
	}
	resource = newResource(&config)
	expected = []label.KeyValue{
		label.String("host.name", host()),
		label.String("service.name", "test-service"),
		label.String("service.version", "test-version"),
		label.String("telemetry.sdk.language", "go"),
		label.String("telemetry.sdk.name", "launcher"),
		label.String("telemetry.sdk.version", version),
	}
	assert.Equal(t, expected, resource.Attributes())

	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.name=test-service-b,host.name=host123")
	config = Config{
		ServiceName:    "test-service-b",
		ServiceVersion: "test-version",
	}
	resource = newResource(&config)
	expected = []label.KeyValue{
		label.String("host.name", "host123"),
		label.String("service.name", "test-service-b"),
		label.String("service.version", "test-version"),
		label.String("telemetry.sdk.language", "go"),
		label.String("telemetry.sdk.name", "launcher"),
		label.String("telemetry.sdk.version", version),
	}
	assert.Equal(t, expected, resource.Attributes())
}

func TestServiceNameViaResourceAttributes(t *testing.T) {
	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.name=test-service-b")
	logger := &testLogger{}
	lsOtel := ConfigureOpentelemetry(WithLogger(logger))
	defer lsOtel.Shutdown()

	expected := "invalid configuration: service name missing"
	if strings.Contains(logger.output[0], expected) {
		t.Errorf("\nString found: %v\nIn: %v", expected, logger.output[0])
	}
}

func TestEmptyHostnameDefaultsToOsHostname(t *testing.T) {
	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "host.name=")
	logger := &testLogger{}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(logger),
		WithServiceName("test-service"),
		WithSpanExporterEndpoint("localhost:443"),
		WithLogLevel("debug"),
		WithResourceAttributes(map[string]string{
			"attr1":     "val1",
			"host.name": "",
		}),
	)
	defer lsOtel.Shutdown()
	output := strings.Join(logger.output[:], ",")
	assert.Contains(t, output, "host.name")
	assert.Contains(t, output, host())
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
	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.name=test-service-name-b")
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
		"OTEL_RESOURCE_ATTRIBUTES",
		"OTEL_EXPORTER_OTLP_METRIC_PERIOD",
	}
	for _, envvar := range vars {
		os.Unsetenv(envvar)
	}
}

func TestMain(m *testing.M) {
	unsetEnvironment()
	os.Exit(m.Run())
}
