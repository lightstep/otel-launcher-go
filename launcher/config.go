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
	"os"
	"time"

	"github.com/sethvargo/go-envconfig"
	"go.opentelemetry.io/collector/translator/conventions"
	hostMetrics "go.opentelemetry.io/contrib/instrumentation/host"
	runtimeMetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/propagation"
	apitrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/label"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/push"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

type Option func(*Config)

// WithAccessToken configures the lightstep access token
func WithAccessToken(accessToken string) Option {
	return func(c *Config) {
		c.AccessToken = accessToken
	}
}

// WithMetricExporterEndpoint configures the endpoint for sending metrics via OTLP
func WithMetricExporterEndpoint(url string) Option {
	return func(c *Config) {
		c.MetricExporterEndpoint = url
	}
}

// WithSpanExporterEndpoint configures the endpoint for sending traces via OTLP
func WithSpanExporterEndpoint(url string) Option {
	return func(c *Config) {
		c.SpanExporterEndpoint = url
	}
}

// WithServiceName configures a "service.name" resource label
func WithServiceName(name string) Option {
	return func(c *Config) {
		c.ServiceName = name
	}
}

// WithServiceVersion configures a "service.version" resource label
func WithServiceVersion(version string) Option {
	return func(c *Config) {
		c.ServiceVersion = version
	}
}

// WithLogLevel configures the logging level for OpenTelemetry
func WithLogLevel(loglevel string) Option {
	return func(c *Config) {
		c.LogLevel = loglevel
	}
}

// WithSpanExporterInsecure permits connecting to the
// trace endpoint without a certificate
func WithSpanExporterInsecure(insecure bool) Option {
	return func(c *Config) {
		c.SpanExporterEndpointInsecure = insecure
	}
}

// WithMetricExporterInsecure permits connecting to the
// metric endpoint without a certificate
func WithMetricExporterInsecure(insecure bool) Option {
	return func(c *Config) {
		c.MetricExporterEndpointInsecure = insecure
	}
}

// WithResourceAttributes configures attributes on the resource
func WithResourceAttributes(attributes map[string]string) Option {
	return func(c *Config) {
		c.resourceAttributes = attributes
	}
}

// WithPropagators configures propagators
func WithPropagators(propagators []string) Option {
	return func(c *Config) {
		c.Propagators = propagators
	}
}

// Configures a global error handler to be used throughout an OpenTelemetry instrumented project.
// See "go.opentelemetry.io/otel/api/global"
func WithErrorHandler(handler otel.ErrorHandler) Option {
	return func(c *Config) {
		c.errorHandler = handler
	}
}

// WithMetricReportingPeriod configures the metric reporting period,
// how often the controller collects and exports metric data.
func WithMetricReportingPeriod(p time.Duration) Option {
	return func(c *Config) {
		c.MetricReportingPeriod = fmt.Sprint(p)
	}
}

type Logger interface {
	Fatalf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}

func WithLogger(logger Logger) Option {
	return func(c *Config) {
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

const (
	// Note: these values should match the defaults used in `env` tags for Config fields.
	// Note: the MetricExporterEndpoint currently defaults to "".  When LS is ready for OTLP metrics
	// we'll set this to `DefaultMetricExporterEndpoint`.

	DefaultSpanExporterEndpoint   = "ingest.lightstep.com:443"
	DefaultMetricExporterEndpoint = "ingest.lightstep.com:443"
)

type Config struct {
	SpanExporterEndpoint           string   `env:"OTEL_EXPORTER_OTLP_SPAN_ENDPOINT,default=ingest.lightstep.com:443"`
	SpanExporterEndpointInsecure   bool     `env:"OTEL_EXPORTER_OTLP_SPAN_INSECURE,default=false"`
	ServiceName                    string   `env:"LS_SERVICE_NAME"`
	ServiceVersion                 string   `env:"LS_SERVICE_VERSION,default=unknown"`
	MetricExporterEndpoint         string   `env:"OTEL_EXPORTER_OTLP_METRIC_ENDPOINT"`
	MetricExporterEndpointInsecure bool     `env:"OTEL_EXPORTER_OTLP_METRIC_INSECURE,default=false"`
	AccessToken                    string   `env:"LS_ACCESS_TOKEN"`
	LogLevel                       string   `env:"OTEL_LOG_LEVEL,default=info"`
	Propagators                    []string `env:"OTEL_PROPAGATORS,default=b3"`
	MetricReportingPeriod          string   `env:"OTEL_EXPORTER_OTLP_METRIC_PERIOD,default=30s"`
	resourceAttributes             map[string]string
	Resource                       *resource.Resource
	logger                         Logger
	errorHandler                   otel.ErrorHandler
}

func checkEndpointDefault(value, defValue string) error {
	if value == "" {
		// The endpoint is disabled.
		return nil
	}
	if value == defValue {
		return fmt.Errorf("invalid configuration: access token missing, must be set when reporting to %s. Set LS_ACCESS_TOKEN env var or configure WithAccessToken in code", value)
	}
	return nil
}

func validateConfiguration(c Config) error {
	if len(c.ServiceName) == 0 {
		serviceNameSet := false
		for _, kv := range c.Resource.Attributes() {
			if kv.Key == conventions.AttributeServiceName {
				if len(kv.Value.AsString()) > 0 {
					serviceNameSet = true
				}
				break
			}
		}
		if !serviceNameSet {
			return errors.New("invalid configuration: service name missing. Set LS_SERVICE_NAME env var or configure WithServiceName in code")
		}
	}

	accessTokenLen := len(c.AccessToken)
	if accessTokenLen == 0 {
		if err := checkEndpointDefault(c.SpanExporterEndpoint, DefaultSpanExporterEndpoint); err != nil {
			return err
		}

		if err := checkEndpointDefault(c.MetricExporterEndpoint, DefaultMetricExporterEndpoint); err != nil {
			return err
		}
	}

	if accessTokenLen > 0 && (accessTokenLen != 32 && accessTokenLen != 84 && accessTokenLen != 104) {
		return fmt.Errorf("invalid configuration: access token length incorrect. Ensure token is set correctly")
	}
	return nil
}

func newConfig(opts ...Option) Config {
	var c Config
	envError := envconfig.Process(context.Background(), &c)
	c.logger = &DefaultLogger{}
	c.errorHandler = &defaultHandler{logger: c.logger}
	var defaultOpts []Option

	for _, opt := range append(defaultOpts, opts...) {
		opt(&c)
	}
	c.Resource = newResource(&c)

	if envError != nil {
		c.logger.Fatalf("environment error: %v", envError)
	}

	return c
}

type Launcher struct {
	config        Config
	shutdownFuncs []func() error
}

// configurePropagators configures B3 propagation by default
func configurePropagators(c *Config) error {
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

func newResource(c *Config) *resource.Resource {
	// workaround until the following change is released
	// https://github.com/open-telemetry/opentelemetry-go/pull/1042
	reset := false
	if len(os.Getenv("OTEL_RESOURCE_LABELS")) == 0 {
		reset = true
		os.Setenv("OTEL_RESOURCE_LABELS", os.Getenv("OTEL_RESOURCE_ATTRIBUTES"))
	}
	detector := resource.FromEnv{}
	r, _ := detector.Detect(context.Background())
	if reset {
		os.Unsetenv("OTEL_RESOURCE_LABELS")
	}
	attributes := []label.KeyValue{
		label.String(conventions.AttributeTelemetrySDKName, "launcher"),
		label.String(conventions.AttributeTelemetrySDKLanguage, "go"),
		label.String(conventions.AttributeTelemetrySDKVersion, version),
	}

	if len(c.ServiceName) > 0 {
		attributes = append(attributes, label.String(conventions.AttributeServiceName, c.ServiceName))
	}

	if len(c.ServiceVersion) > 0 {
		attributes = append(attributes, label.String(conventions.AttributeServiceVersion, c.ServiceVersion))
	}

	for key, value := range c.resourceAttributes {
		if len(value) > 0 {
			attributes = append(attributes, label.String(key, value))
		}
	}

	attributes = append(r.Attributes(), attributes...)
	return resource.New(attributes...)
}

func newExporter(accessToken, endpoint string, insecure bool) (*otlp.Exporter, error) {
	headers := map[string]string{
		"lightstep-access-token": accessToken,
	}

	secureOption := otlp.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if insecure {
		secureOption = otlp.WithInsecure()
	}
	return otlp.NewExporter(
		secureOption,
		otlp.WithAddress(endpoint),
		otlp.WithHeaders(headers),
	)
}

func setupTracing(c Config) (func() error, error) {
	if c.SpanExporterEndpoint == "" {
		c.logger.Debugf("tracing is disabled by configuration: no endpoint set")
		return nil, nil
	}
	spanExporter, err := newExporter(c.AccessToken, c.SpanExporterEndpoint, c.SpanExporterEndpointInsecure)
	if err != nil {
		return nil, fmt.Errorf("failed to create span exporter: %v", err)
	}

	tp, err := trace.NewProvider(
		trace.WithConfig(trace.Config{DefaultSampler: trace.AlwaysSample()}),
		trace.WithSyncer(spanExporter),
		trace.WithResource(c.Resource),
	)
	if err != nil {
		return nil, err
	}

	if err = configurePropagators(&c); err != nil {
		return nil, err
	}

	global.SetTraceProvider(tp)

	return func() error {
		return spanExporter.Stop()
	}, nil
}

type setupFunc func(Config) (func() error, error)

func setupMetrics(c Config) (func() error, error) {
	if c.MetricExporterEndpoint == "" {
		c.logger.Debugf("metrics are disabled by configuration: no endpoint set")
		return nil, nil
	}
	metricExporter, err := newExporter(c.AccessToken, c.MetricExporterEndpoint, c.MetricExporterEndpointInsecure)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %v", err)
	}

	period := controller.DefaultPushPeriod
	if c.MetricReportingPeriod != "" {
		period, err = time.ParseDuration(c.MetricReportingPeriod)
		if err != nil {
			return nil, fmt.Errorf("invalid metric reporting period: %v", err)
		}
		if period <= 0 {

			return nil, fmt.Errorf("invalid metric reporting period: %v", c.MetricReportingPeriod)
		}
	}
	pusher := controller.New(
		processor.New(
			selector.NewWithInexpensiveDistribution(),
			metricExporter,
		),
		metricExporter,
		controller.WithResource(c.Resource),
		controller.WithPeriod(period),
	)

	pusher.Start()

	provider := pusher.Provider()

	if err = runtimeMetrics.Start(runtimeMetrics.WithMeterProvider(provider)); err != nil {
		return nil, fmt.Errorf("failed to start runtime metrics: %v", err)
	}

	if err = hostMetrics.Start(hostMetrics.WithMeterProvider(provider)); err != nil {
		return nil, fmt.Errorf("failed to start host metrics: %v", err)
	}

	global.SetMeterProvider(provider)
	return func() error {
		pusher.Stop()
		return metricExporter.Stop()
	}, nil
}

func ConfigureOpentelemetry(opts ...Option) Launcher {
	c := newConfig(opts...)

	if c.LogLevel == "debug" {
		c.logger.Debugf("debug logging enabled")
		c.logger.Debugf("configuration")
		s, _ := json.MarshalIndent(c, "", "\t")
		c.logger.Debugf(string(s))
	}

	err := validateConfiguration(c)
	if err != nil {
		c.logger.Fatalf("configuration error: %v", err)
	}

	if c.errorHandler != nil {
		global.SetErrorHandler(c.errorHandler)
	}

	ls := Launcher{
		config: c,
	}
	for _, setup := range []setupFunc{setupTracing, setupMetrics} {
		shutdown, err := setup(c)
		if err != nil {
			c.logger.Fatalf("setup error: %v", err)
			continue
		}
		if shutdown != nil {
			ls.shutdownFuncs = append(ls.shutdownFuncs, shutdown)
		}
	}
	return ls
}

func (ls Launcher) Shutdown() {
	for _, shutdown := range ls.shutdownFuncs {
		if err := shutdown(); err != nil {
			ls.config.logger.Fatalf("failed to stop exporter: %v", err)
		}
	}
}
