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
	"sync"
	"testing"
	"time"

	"github.com/lightstep/otel-launcher-go/pipelines/test"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	expectedAccessTokenLengthError  = "invalid configuration: access token length incorrect. Ensure token is set correctly"
	expectedAccessTokenMissingError = "invalid configuration: access token missing, must be set when reporting to ingest.lightstep.com"
	expectedTracingDisabledMessage  = "tracing is disabled by configuration: no endpoint set"
	expectedMetricsDisabledMessage  = "metrics are disabled by configuration: no endpoint set"
)

type testSuite struct {
	suite.Suite

	*test.Server

	testLogger
	testErrorHandler
}

func (suite *testSuite) SetupSuite() {
	suite.Server = test.NewServer(suite.T())
}

func (suite *testSuite) SetupTest() {
	suite.testLogger.reset()
}

func (suite *testSuite) bothInsecureEndpointOptions() []Option {
	return []Option{
		WithMetricExporterEndpoint(fmt.Sprintf(":%d", suite.Server.InsecureMetricsPort)),
		WithSpanExporterEndpoint(fmt.Sprintf(":%d", suite.Server.InsecureTracePort)),
		WithSpanExporterInsecure(true),
		WithMetricExporterInsecure(true),
	}
}

func (suite *testSuite) insecureTraceEndpointOptions() []Option {
	return []Option{
		WithSpanExporterEndpoint(fmt.Sprintf(":%d", suite.Server.InsecureTracePort)),
		WithSpanExporterInsecure(true),
	}
}

func (suite *testSuite) insecureMetricsEndpointOptions() []Option {
	return []Option{
		WithSpanExporterEndpoint(fmt.Sprintf(":%d", suite.Server.InsecureMetricsPort)),
		WithSpanExporterInsecure(true),
	}
}

func (suite *testSuite) TearDownTest() {
	unsetEnvironment()
	suite.testLogger.reset()
}

func (suite *testSuite) TearDownSuite() {
	suite.Server.Stop()
}

func TestLauncherSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}

type testLogger struct {
	lock   sync.Mutex
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

func (suite *testSuite) getOutput() []string {
	suite.testLogger.lock.Lock()
	defer suite.testLogger.lock.Unlock()
	return suite.testLogger.output
}

func (suite *testSuite) requireLogContains(expected string) {
	suite.T().Helper()

	for _, output := range suite.getOutput() {
		if strings.Contains(output, expected) {
			return
		}
	}

	suite.T().Errorf("\nString unexpectedly not found: %v\nIn: %v", expected, suite.getOutput())
}

func (suite *testSuite) requireLogNotContains(expected string) {
	suite.T().Helper()

	for _, output := range suite.getOutput() {
		if strings.Contains(output, expected) {
			suite.T().Errorf("\nString unexpectedly found: %v\nIn: %v", expected, suite.getOutput())
			return
		}
	}
}

func (logger *testLogger) reset() {
	logger.lock.Lock()
	defer logger.lock.Unlock()
	logger.output = nil
}

type testErrorHandler struct {
	lock sync.Mutex
	errs []error
}

func (t *testErrorHandler) Handle(err error) {
	fmt.Printf("test error handler handled error: %v\n", err)

	t.lock.Lock()
	defer t.lock.Unlock()
	t.errs = append(t.errs, err)
}

func fakeAccessToken() string {
	return strings.Repeat("1", 32)
}

func (suite *testSuite) TestInvalidServiceName() {
	lsOtel := ConfigureOpentelemetry(WithLogger(&suite.testLogger))
	defer lsOtel.Shutdown()

	expected := "invalid configuration: service name missing"
	suite.requireLogContains(expected)
}

func (suite *testSuite) testInvalidMissingAccessToken(opts ...Option) {
	lsOtel := ConfigureOpentelemetry(
		append(opts,
			WithLogger(&suite.testLogger),
			WithServiceName("test-service"),
		)...,
	)
	defer lsOtel.Shutdown()

	suite.requireLogContains(expectedAccessTokenMissingError)
}

func (suite *testSuite) TestInvalidMissingDefaultAccessToken() {
	suite.testInvalidMissingAccessToken(
		WithAccessToken(""),
	)
}

func (suite *testSuite) TestInvalidTraceDefaultAccessToken() {
	suite.testInvalidMissingAccessToken(
		append(suite.insecureMetricsEndpointOptions(),
			WithAccessToken(""),
			WithSpanExporterEndpoint(DefaultSpanExporterEndpoint),
		)...,
	)
}

func (suite *testSuite) TestInvalidMetricDefaultAccessToken() {
	suite.testInvalidMissingAccessToken(
		append(suite.insecureTraceEndpointOptions(),
			WithAccessToken(""),
			WithMetricExporterEndpoint(DefaultMetricExporterEndpoint),
		)...,
	)
}

func (suite *testSuite) testInvalidAccessToken(opts ...Option) {
	lsOtel := ConfigureOpentelemetry(
		append(opts,
			WithLogger(&suite.testLogger),
			WithServiceName("test-service"),
		)...,
	)
	defer lsOtel.Shutdown()

	suite.requireLogContains(expectedAccessTokenLengthError)
}

func (suite *testSuite) TestInvalidTraceAccessTokenLength() {
	suite.testInvalidAccessToken(
		append(suite.insecureTraceEndpointOptions(),
			WithAccessToken("1234"),
		)...,
	)
}

func (suite *testSuite) TestInvalidMetricAccessTokenLength() {
	suite.testInvalidAccessToken(
		append(suite.bothInsecureEndpointOptions(),
			WithAccessToken("1234"),
		)...,
	)
}

func (suite *testSuite) testEndpointDisabled(expected string, opts ...Option) {
	lsOtel := ConfigureOpentelemetry(
		append(opts,
			WithLogger(&suite.testLogger),
			WithServiceName("test-service"),
			WithMetricsEnabled(false),
		)...,
	)
	defer lsOtel.Shutdown()

	suite.requireLogNotContains(expectedAccessTokenMissingError)
	suite.requireLogContains(expected)
}

func (suite *testSuite) TestTraceEndpointDisabled() {
	suite.testEndpointDisabled(
		expectedTracingDisabledMessage,
		WithAccessToken(fakeAccessToken()),
		WithSpanExporterEndpoint(""),
	)
}

func (suite *testSuite) TestMetricEndpointDisabled() {
	suite.testEndpointDisabled(
		expectedMetricsDisabledMessage,
		WithAccessToken(fakeAccessToken()),
		WithMetricExporterEndpoint(""),
	)
}

func (suite *testSuite) TestValidConfig() {
	lsOtel := ConfigureOpentelemetry(
		WithLogger(&suite.testLogger),
		WithServiceName("test-service"),
		WithAccessToken(fakeAccessToken()),
		WithErrorHandler(&suite.testErrorHandler),
	)
	defer lsOtel.Shutdown()

	lsOtel = ConfigureOpentelemetry(
		append(suite.bothInsecureEndpointOptions(),
			WithLogger(&suite.testLogger),
			WithServiceName("test-service"),
		)...,
	)
	defer lsOtel.Shutdown()

	if len(suite.getOutput()) > 0 {
		suite.T().Errorf("\nExpected: no logs\ngot: %v", suite.getOutput())
	}
}

func (suite *testSuite) TestInvalidEnvironment() {
	os.Setenv("OTEL_EXPORTER_OTLP_METRIC_INSECURE", "bleargh")

	lsOtel := ConfigureOpentelemetry(
		WithLogger(&suite.testLogger),
		WithServiceName("test-service"),
	)
	defer lsOtel.Shutdown()

	suite.requireLogContains("environment error")
}

func (suite *testSuite) TestInvalidMetricsPushIntervalEnv() {
	os.Setenv("OTEL_EXPORTER_OTLP_METRIC_PERIOD", "300million")

	lsOtel := ConfigureOpentelemetry(
		WithLogger(&suite.testLogger),
		WithServiceName("test-service"),
		WithSpanExporterEndpoint("127.0.0.1:4000"),
		WithMetricExporterEndpoint("127.0.0.1:4000"),
	)
	defer lsOtel.Shutdown()

	suite.requireLogContains("setup error: invalid metric reporting period")
}

func (suite *testSuite) TestInvalidMetricsPushIntervalConfig() {
	lsOtel := ConfigureOpentelemetry(
		WithLogger(&suite.testLogger),
		WithServiceName("test-service"),
		WithSpanExporterEndpoint("127.0.0.1:4000"),
		WithMetricExporterEndpoint("127.0.0.1:4000"),
		WithMetricReportingPeriod(-time.Second),
	)
	defer lsOtel.Shutdown()

	suite.requireLogContains("setup error: invalid metric reporting period")
}

func (suite *testSuite) TestDebugEnabled() {
	lsOtel := ConfigureOpentelemetry(
		WithLogger(&suite.testLogger),
		WithServiceName("test-service"),
		WithAccessToken("access-token-123-123456789abcdef"),
		WithSpanExporterEndpoint("localhost:443"),
		WithLogLevel("debug"),
		WithResourceAttributes(map[string]string{
			"attr1":     "val1",
			"host.name": "host456",
		}),
	)
	defer lsOtel.Shutdown()
	output := strings.Join(suite.getOutput()[:], ",")
	assert := suite.Assert()
	assert.Contains(output, "debug logging enabled")
	assert.Contains(output, "test-service")
	assert.Contains(output, "access-token-123")
	assert.Contains(output, "localhost:443")
	assert.Contains(output, "attr1")
	assert.Contains(output, "val1")
	assert.Contains(output, "host.name")
	assert.Contains(output, "host456")
}

func (suite *testSuite) TestDefaultConfig() {
	assert := suite.Assert()
	config := newConfig(
		WithLogger(&suite.testLogger),
		WithErrorHandler(&suite.testErrorHandler),
	)

	attributes := []attribute.KeyValue{
		attribute.String("host.name", host()),
		attribute.String("service.version", "unknown"),
		attribute.String("telemetry.sdk.name", "launcher"),
		attribute.String("telemetry.sdk.language", "go"),
		attribute.String("telemetry.sdk.version", version),
	}

	expected := Config{
		ServiceName:                    "",
		ServiceVersion:                 "unknown",
		SpanExporterEndpoint:           "ingest.lightstep.com:443",
		SpanExporterEndpointInsecure:   false,
		MetricExporterEndpoint:         "ingest.lightstep.com:443",
		MetricExporterEndpointInsecure: false,
		MetricReportingPeriod:          "30s",
		MetricsEnabled:                 true,
		MetricTemporalityPreference:    "cumulative",
		LogLevel:                       "info",
		Propagators:                    []string{"b3"},
		Resource:                       resource.NewWithAttributes(semconv.SchemaURL, attributes...),
		logger:                         &suite.testLogger,
		errorHandler:                   &suite.testErrorHandler,
	}
	assert.Equal(expected, config)
}

func (suite *testSuite) TestEnvironmentVariables() {
	assert := suite.Assert()

	setEnvironment()

	config := newConfig(
		WithLogger(&suite.testLogger),
		WithErrorHandler(&suite.testErrorHandler),
	)

	attributes := []attribute.KeyValue{
		attribute.String("host.name", host()),
		attribute.String("service.name", "test-service-name"),
		attribute.String("service.version", "test-service-version"),
		attribute.String("telemetry.sdk.name", "launcher"),
		attribute.String("telemetry.sdk.language", "go"),
		attribute.String("telemetry.sdk.version", version),
	}

	expected := Config{
		ServiceName:                    "test-service-name",
		ServiceVersion:                 "test-service-version",
		SpanExporterEndpoint:           "satellite-url",
		SpanExporterEndpointInsecure:   true,
		MetricExporterEndpoint:         "metrics-url",
		MetricExporterEndpointInsecure: true,
		MetricReportingPeriod:          "30s",
		MetricTemporalityPreference:    "delta",
		LogLevel:                       "debug",
		Propagators:                    []string{"b3", "w3c"},
		Resource:                       resource.NewWithAttributes(semconv.SchemaURL, attributes...),
		logger:                         &suite.testLogger,
		errorHandler:                   &suite.testErrorHandler,
	}
	assert.Equal(expected, config)

}

func (suite *testSuite) TestConfigurationOverrides() {
	assert := suite.Assert()

	setEnvironment()

	config := newConfig(
		WithServiceName("override-service-name"),
		WithServiceVersion("override-service-version"),
		WithAccessToken("override-access-token"),
		WithSpanExporterEndpoint("override-satellite-url"),
		WithSpanExporterInsecure(false),
		WithMetricExporterEndpoint("override-metrics-url"),
		WithMetricExporterInsecure(false),
		WithMetricTemporalityPreference("stateless"),
		WithLogLevel("info"),
		WithLogger(&suite.testLogger),
		WithErrorHandler(&suite.testErrorHandler),
		WithPropagators([]string{"b3"}),
	)

	attributes := []attribute.KeyValue{
		attribute.String("host.name", host()),
		attribute.String("service.name", "override-service-name"),
		attribute.String("service.version", "override-service-version"),
		attribute.String("telemetry.sdk.name", "launcher"),
		attribute.String("telemetry.sdk.language", "go"),
		attribute.String("telemetry.sdk.version", version),
	}

	expected := Config{
		ServiceName:                    "override-service-name",
		ServiceVersion:                 "override-service-version",
		SpanExporterEndpoint:           "override-satellite-url",
		SpanExporterEndpointInsecure:   false,
		MetricExporterEndpoint:         "override-metrics-url",
		MetricExporterEndpointInsecure: false,
		MetricReportingPeriod:          "30s",
		MetricTemporalityPreference:    "stateless",
		Headers:                        map[string]string{"lightstep-access-token": "override-access-token"},
		LogLevel:                       "info",
		Propagators:                    []string{"b3"},
		Resource:                       resource.NewWithAttributes(semconv.SchemaURL, attributes...),
		logger:                         &suite.testLogger,
		errorHandler:                   &suite.testErrorHandler,
	}
	assert.Equal(expected, config)
}

type TestCarrier struct {
	values map[string]string
}

func (t TestCarrier) Keys() []string {
	keys := make([]string, 0, len(t.values))
	for k := range t.values {
		keys = append(keys, k)
	}
	return keys
}

func (t TestCarrier) Get(key string) string {
	return t.values[key]
}

func (t TestCarrier) Set(key string, value string) {
	t.values[key] = value
}

func (suite *testSuite) TestConfigurePropagators() {
	assert := suite.Assert()
	mem1, _ := baggage.NewMember("keyone", "foo1")
	mem2, _ := baggage.NewMember("keytwo", "bar1")
	bag, _ := baggage.New(mem1, mem2)

	ctx := baggage.ContextWithBaggage(context.Background(), bag)

	lsOtel := ConfigureOpentelemetry(
		append(suite.insecureTraceEndpointOptions(),
			WithLogger(&suite.testLogger),
			WithServiceName("test-service"),
		)...,
	)
	defer lsOtel.Shutdown()
	ctx, finish := otel.Tracer("ex.com/basic").Start(ctx, "foo")
	defer finish.End()
	carrier := TestCarrier{values: map[string]string{}}
	prop := otel.GetTextMapPropagator()
	prop.Inject(ctx, carrier)
	assert.Greater(len(carrier.Get("x-b3-traceid")), 0)
	assert.Equal("", carrier.Get("baggage"))
	assert.Equal(len(carrier.Get("traceparent")), 0)

	lsOtel = ConfigureOpentelemetry(
		append(suite.insecureTraceEndpointOptions(),
			WithLogger(&suite.testLogger),
			WithServiceName("test-service"),
			WithPropagators([]string{"b3", "baggage", "tracecontext"}),
		)...,
	)
	defer lsOtel.Shutdown()
	carrier = TestCarrier{values: map[string]string{}}
	prop = otel.GetTextMapPropagator()
	prop.Inject(ctx, carrier)
	assert.Greater(len(carrier.Get("x-b3-traceid")), 0)
	assert.Contains(carrier.Get("baggage"), "keytwo=bar1")
	assert.Contains(carrier.Get("baggage"), "keyone=foo1")
	assert.Greater(len(carrier.Get("traceparent")), 0)

	lsOtel = ConfigureOpentelemetry(
		append(suite.bothInsecureEndpointOptions(),
			WithLogger(&suite.testLogger),
			WithServiceName("test-service"),
			WithPropagators([]string{"invalid"}),
		)...,
	)
	defer lsOtel.Shutdown()

	expected := "setup error: invalid configuration: unsupported propagators. Supported options: b3,baggage,tracecontext,ottrace"
	if !assert.Contains(suite.getOutput(), expected) {
		suite.T().Errorf("\nString not found: %v\nIn: %v", expected, suite.getOutput())
	}
}

func host() string {
	host, _ := os.Hostname()
	return host
}

func (suite *testSuite) TestConfigureResourcesAttributes() {
	assert := suite.Assert()
	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "label1=value1,label2=value2")
	config := Config{
		ServiceName:    "test-service",
		ServiceVersion: "test-version",
	}
	resource := newResource(&config)
	expected := []attribute.KeyValue{
		attribute.String("host.name", host()),
		attribute.String("label1", "value1"),
		attribute.String("label2", "value2"),
		attribute.String("service.name", "test-service"),
		attribute.String("service.version", "test-version"),
		attribute.String("telemetry.sdk.language", "go"),
		attribute.String("telemetry.sdk.name", "launcher"),
		attribute.String("telemetry.sdk.version", version),
	}
	assert.Equal(expected, resource.Attributes())

	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "telemetry.sdk.language=test-language")
	config = Config{
		ServiceName:    "test-service",
		ServiceVersion: "test-version",
	}
	resource = newResource(&config)
	expected = []attribute.KeyValue{
		attribute.String("host.name", host()),
		attribute.String("service.name", "test-service"),
		attribute.String("service.version", "test-version"),
		attribute.String("telemetry.sdk.language", "go"),
		attribute.String("telemetry.sdk.name", "launcher"),
		attribute.String("telemetry.sdk.version", version),
	}
	assert.Equal(expected, resource.Attributes())

	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.name=test-service-b,host.name=host123")
	config = Config{
		ServiceName:    "test-service-b",
		ServiceVersion: "test-version",
	}
	resource = newResource(&config)
	expected = []attribute.KeyValue{
		attribute.String("host.name", "host123"),
		attribute.String("service.name", "test-service-b"),
		attribute.String("service.version", "test-version"),
		attribute.String("telemetry.sdk.language", "go"),
		attribute.String("telemetry.sdk.name", "launcher"),
		attribute.String("telemetry.sdk.version", version),
	}
	assert.Equal(expected, resource.Attributes())
}

func (suite *testSuite) TestServiceNameViaResourceAttributes() {
	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.name=test-service-b")
	lsOtel := ConfigureOpentelemetry(WithLogger(&suite.testLogger))
	defer lsOtel.Shutdown()

	expected := "invalid configuration: service name missing"
	if strings.Contains(suite.getOutput()[0], expected) {
		suite.T().Errorf("\nString found: %v\nIn: %v", expected, suite.getOutput()[0])
	}
}

func (suite *testSuite) TestEmptyHostnameDefaultsToOsHostname() {
	assert := suite.Assert()
	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "host.name=")
	lsOtel := ConfigureOpentelemetry(
		append(suite.bothInsecureEndpointOptions(),
			WithServiceName("test-service"),
			WithAccessToken(fakeAccessToken()),
			WithResourceAttributes(map[string]string{
				"attr1":     "val1",
				"host.name": "",
			}),
		)...,
	)
	defer lsOtel.Shutdown()

	attrs := attribute.NewSet(lsOtel.config.Resource.Attributes()...)
	v, ok := attrs.Value("host.name")
	assert.Equal(host(), v.AsString())
	assert.True(ok)
}

func (suite *testSuite) TestConfigWithResourceAttributes() {
	assert := suite.Assert()
	lsOtel := ConfigureOpentelemetry(
		append(suite.bothInsecureEndpointOptions(),
			WithServiceName("test-service"),
			WithAccessToken(fakeAccessToken()),
			WithResourceAttributes(map[string]string{
				"attr1": "val1",
				"attr2": "val2",
			}),
		)...,
	)
	defer lsOtel.Shutdown()
	attrs := attribute.NewSet(lsOtel.config.Resource.Attributes()...)
	v, ok := attrs.Value("attr1")
	assert.Equal("val1", v.AsString())
	assert.True(ok)

	v, ok = attrs.Value("attr2")
	assert.Equal("val2", v.AsString())
	assert.True(ok)
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
	os.Setenv("OTEL_EXPORTER_OTLP_METRIC_TEMPORALITY_PREFERENCE", "delta")
	os.Setenv("LS_METRICS_ENABLED", "false")
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
		"OTEL_EXPORTER_OTLP_METRIC_TEMPORALITY_PREFERENCE",
		"LS_METRICS_ENABLED",
	}
	for _, envvar := range vars {
		os.Unsetenv(envvar)
	}
}

func TestMain(m *testing.M) {
	unsetEnvironment()
	os.Exit(m.Run())
}
