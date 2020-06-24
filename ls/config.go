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

package ls

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/sethvargo/go-envconfig/pkg/envconfig"
	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/propagation"
	apitrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

type Option func(*LightstepConfig)

func WithAccessToken(accessToken string) Option {
	return func(c *LightstepConfig) {
		c.AccessToken = accessToken
	}
}

func WithSatelliteURL(url string) Option {
	return func(c *LightstepConfig) {
		c.SatelliteURL = url
	}
}

func WithServiceName(name string) Option {
	return func(c *LightstepConfig) {
		c.ServiceName = name
	}
}

func WithDebug(debug bool) Option {
	return func(c *LightstepConfig) {
		c.Debug = debug
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

var (
	defaultSatelliteURL = "ingest.lightstep.com:443"
	Version             = "0.0.1" // overridden at build time
)

type LightstepConfig struct {
	ServiceName    string `env:"LS_SERVICE_NAME"`
	ServiceVersion string `env:"LS_SERVICE_VERSION,default=unknown"`
	SatelliteURL   string `env:"LS_SATELLITE_URL,default=ingest.lightstep.com:443"`
	MetricsURL     string `env:"LS_METRICS_URL,default=ingest.lightstep.com:443/metrics"`
	AccessToken    string `env:"LS_ACCESS_TOKEN"`
	Debug          bool   `env:"LS_DEBUG,default=false"`
	Insecure       bool   `env:"LS_INSECURE,default=false"`
	logger         Logger
}

func validateConfiguration(c LightstepConfig) error {
	if len(c.ServiceName) == 0 {
		return errors.New("invalid configuration: service name missing. Set LS_SERVICE_NAME or configure WithServiceName")
	}

	if len(c.AccessToken) == 0 && c.SatelliteURL == defaultSatelliteURL {
		return fmt.Errorf("invalid configuration: access token missing, must be set when reporting to %s. Set LS_ACCESS_TOKEN or configure WithAccessToken", c.SatelliteURL)
	}
	return nil
}

func newConfig(opts ...Option) LightstepConfig {
	var c LightstepConfig
	if err := envconfig.Process(context.Background(), &c); err != nil {
		log.Fatal(err)
	}
	c.logger = &DefaultLogger{}
	var defaultOpts []Option

	for _, opt := range append(defaultOpts, opts...) {
		opt(&c)
	}

	return c
}

type LightstepOpentelemetry struct {
	spanExporter *otlp.Exporter
}

func configurePropagators() {
	tcPropagator := apitrace.TraceContext{}
	b3Propagator := apitrace.B3{}
	ccPropagator := correlation.CorrelationContext{}
	global.SetPropagators(propagation.New(
		propagation.WithExtractors(tcPropagator, b3Propagator, ccPropagator),
		propagation.WithInjectors(tcPropagator, b3Propagator, ccPropagator),
	))
}

func ConfigureOpentelemetry(opts ...Option) LightstepOpentelemetry {
	c := newConfig(opts...)

	err := validateConfiguration(c)
	if err != nil {
		c.logger.Fatalf(err.Error())
	}

	if c.Debug {
		c.logger.Debugf("debug logging enabled")
		c.logger.Debugf("configuration")
		s, _ := json.MarshalIndent(c, "", "\t")
		c.logger.Debugf(string(s)) // TODO: pretty print
	}

	headers := map[string]string{
		"lightstep-access-token": c.AccessToken,
	}

	secureOption := otlp.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if c.Insecure {
		secureOption = otlp.WithInsecure()
	}
	exporter, err := otlp.NewExporter(
		secureOption,
		otlp.WithAddress(c.SatelliteURL),
		otlp.WithHeaders(headers),
	)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}
	resources := resource.New(
		// TODO: use keys from the semantic convention definition
		kv.String("service.name", c.ServiceName),
		kv.String("service.version", c.ServiceVersion),
		kv.String("telemetry.sdk.language", "go"),
		kv.String("telemetry.sdk.version", Version),
	)
	tp, err := trace.NewProvider(
		trace.WithConfig(trace.Config{DefaultSampler: trace.AlwaysSample()}),
		trace.WithSyncer(exporter),
		trace.WithResource(resources),
	)
	if err != nil {
		log.Fatal(err)
	}

	configurePropagators()

	global.SetTraceProvider(tp)
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
