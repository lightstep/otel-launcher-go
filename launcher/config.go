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
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/sethvargo/go-envconfig/pkg/envconfig"
	"go.opentelemetry.io/collector/translator/conventions"
	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/oterror"
	"go.opentelemetry.io/otel/api/propagation"
	apitrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

type Option func(*LightstepConfig)

// WithAccessToken configures the lightstep access token
func WithAccessToken(accessToken string) Option {
	return func(c *LightstepConfig) {
		c.AccessToken = accessToken
	}
}

// WithMetricExporterEndpoint configures the endpoint for sending metrics via OTLP
func WithMetricExporterEndpoint(url string) Option {
	return func(c *LightstepConfig) {
		c.MetricExporterEndpoint = url
	}
}

// WithSpanExporterEndpoint configures the endpoint for sending traces via OTLP
func WithSpanExporterEndpoint(url string) Option {
	return func(c *LightstepConfig) {
		c.SpanExporterEndpoint = url
	}
}

// WithServiceName configures a "service.name" resource label
func WithServiceName(name string) Option {
	return func(c *LightstepConfig) {
		c.ServiceName = name
	}
}

// WithServiceVersion configures a "service.version" resource label
func WithServiceVersion(version string) Option {
	return func(c *LightstepConfig) {
		c.ServiceVersion = version
	}
}

// WithLogLevel configures the logging level for OpenTelemetry
func WithLogLevel(loglevel string) Option {
	return func(c *LightstepConfig) {
		c.LogLevel = loglevel
	}
}

// WithSpanExporterInsecure permits connecting to the
// trace endpoint without a certificate
func WithSpanExporterInsecure(insecure bool) Option {
	return func(c *LightstepConfig) {
		c.SpanExporterEndpointInsecure = insecure
	}
}

// WithMetricExporterInsecure permits connecting to the
// metric endpoint without a certificate
func WithMetricExporterInsecure(insecure bool) Option {
	return func(c *LightstepConfig) {
		c.MetricExporterEndpointInsecure = insecure
	}
}

// WithResourceLabels configures attributes on the resource
func WithResourceLabels(labels map[string]string) Option {
	return func(c *LightstepConfig) {
		c.ResourceLabels = labels
	}
}

// WithPropagators configures propagators
func WithPropagators(propagators []string) Option {
	return func(c *LightstepConfig) {
		c.Propagators = propagators
	}
}

// Configures a global error handler to be used throughout an OpenTelemetry instrumented project.
// See "go.opentelemetry.io/otel/api/global"
func WithErrorHandler(handler oterror.Handler) Option {
	return func(c *LightstepConfig) {
		c.errorHandler = handler
	}
}

type Logger interface {
	Fatalf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}

func WithLogger(logger Logger) Option {
	return func(c *LightstepConfig) {
		c.logger = logger
	}
}

type DefaultLogger struct {
}

func (l *DefaultLogger) Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

func (l *DefaultLogger) Debugf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

type defaultHandler struct {
	logger Logger
}

func (l *defaultHandler) Handle(err error) {
	l.logger.Debugf("error: %v\n", err)
}

var (
	defaultSpanExporterEndpoint = "ingest.lightstep.com:443"
)

type LightstepConfig struct {
	SpanExporterEndpoint           string            `env:"OTEL_EXPORTER_OTLP_SPAN_ENDPOINT,default=ingest.lightstep.com:443"`
	SpanExporterEndpointInsecure   bool              `env:"OTEL_EXPORTER_OTLP_SPAN_INSECURE,default=false"`
	ServiceName                    string            `env:"LS_SERVICE_NAME"`
	ServiceVersion                 string            `env:"LS_SERVICE_VERSION,default=unknown"`
	MetricExporterEndpoint         string            `env:"OTEL_EXPORTER_OTLP_METRIC_ENDPOINT,default=ingest.lightstep.com:443/metrics"`
	MetricExporterEndpointInsecure bool              `env:"OTEL_EXPORTER_OTLP_METRIC_INSECURE,default=false"`
	AccessToken                    string            `env:"LS_ACCESS_TOKEN"`
	LogLevel                       string            `env:"OTEL_LOG_LEVEL,default=info"`
	Propagators                    []string          `env:"OTEL_PROPAGATORS,default=b3"`
	ResourceLabels                 map[string]string `env:"OTEL_RESOURCE_LABELS"`
	logger                         Logger
	errorHandler                   oterror.Handler
}

func validateConfiguration(c LightstepConfig) error {
	if len(c.ServiceName) == 0 {
		return errors.New("invalid configuration: service name missing. Set LS_SERVICE_NAME env var or configure WithServiceName in code")
	}

	accessTokenLen := len(c.AccessToken)
	if accessTokenLen == 0 && c.SpanExporterEndpoint == defaultSpanExporterEndpoint {
		return fmt.Errorf("invalid configuration: access token missing, must be set when reporting to %s. Set LS_ACCESS_TOKEN env var or configure WithAccessToken in code", c.SpanExporterEndpoint)
	}

	if accessTokenLen > 0 && (accessTokenLen != 32 && accessTokenLen != 84 && accessTokenLen != 104) {
		return fmt.Errorf("invalid configuration: access token length incorrect. Ensure token is set correctly")
	}
	return nil
}

func newConfig(opts ...Option) LightstepConfig {
	var c LightstepConfig
	if err := envconfig.Process(context.Background(), &c); err != nil {
		log.Fatal(err)
	}
	c.logger = &DefaultLogger{}
	c.errorHandler = &defaultHandler{logger: c.logger}
	var defaultOpts []Option

	for _, opt := range append(defaultOpts, opts...) {
		opt(&c)
	}

	return c
}

type LightstepOpentelemetry struct {
	spanExporter *otlp.Exporter
}

// configurePropagators configures B3 propagation by default
func configurePropagators(c *LightstepConfig) error {
	propagatorsMap := map[string]propagation.HTTPPropagator{
		"b3": apitrace.B3{},
		"cc": correlation.CorrelationContext{},
	}
	var extractors []propagation.HTTPExtractor
	var injectors []propagation.HTTPInjector
	for _, key := range c.Propagators {
		prop := propagatorsMap[key]
		if prop != nil {
			extractors = append(extractors, prop)
			injectors = append(injectors, prop)
		}
	}
	if len(extractors) == 0 || len(injectors) == 0 {
		return fmt.Errorf("invalid configuration: unsupported propagators. Supported options: b3,cc")
	}
	global.SetPropagators(propagation.New(
		propagation.WithExtractors(extractors...),
		propagation.WithInjectors(injectors...),
	))
	return nil
}

func newResource(c *LightstepConfig) *resource.Resource {
	attributes := []kv.KeyValue{
		kv.String(conventions.AttributeServiceName, c.ServiceName),
		kv.String(conventions.AttributeServiceVersion, c.ServiceVersion),
		kv.String(conventions.AttributeTelemetrySDKName, "launcher"),
		kv.String(conventions.AttributeTelemetrySDKLanguage, "go"),
		kv.String(conventions.AttributeTelemetrySDKVersion, version),
	}
	for key, attribute := range c.ResourceLabels {
		attributes = append(attributes, kv.String(key, attribute))
	}
	return resource.New(attributes...)
}

func ConfigureOpentelemetry(opts ...Option) LightstepOpentelemetry {
	c := newConfig(opts...)

	if c.LogLevel == "debug" {
		c.logger.Debugf("debug logging enabled")
		c.logger.Debugf("configuration")
		s, _ := json.MarshalIndent(c, "", "\t")
		c.logger.Debugf(string(s))
	}

	err := validateConfiguration(c)
	if err != nil {
		c.logger.Fatalf(err.Error())
	}

	headers := map[string]string{
		"lightstep-access-token": c.AccessToken,
	}

	secureOption := otlp.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if c.SpanExporterEndpointInsecure {
		secureOption = otlp.WithInsecure()
	}
	exporter, err := otlp.NewExporter(
		secureOption,
		otlp.WithAddress(c.SpanExporterEndpoint),
		otlp.WithHeaders(headers),
	)
	if err != nil {
		c.logger.Fatalf("failed to create exporter: %v", err)
	}

	tp, err := trace.NewProvider(
		trace.WithConfig(trace.Config{DefaultSampler: trace.AlwaysSample()}),
		trace.WithSyncer(exporter),
		trace.WithResource(newResource(&c)),
	)
	if err != nil {
		c.logger.Fatalf(err.Error())
	}

	err = configurePropagators(&c)
	if err != nil {
		c.logger.Fatalf(err.Error())
	}

	global.SetTraceProvider(tp)

	if c.errorHandler != nil {
		global.SetHandler(c.errorHandler)
	}
	return LightstepOpentelemetry{
		spanExporter: exporter,
	}
}

func (ls *LightstepOpentelemetry) Shutdown() {
	err := ls.spanExporter.Stop()
	if err != nil {
		log.Fatalf("failed to stop exporter: %v", err)
	}
}
