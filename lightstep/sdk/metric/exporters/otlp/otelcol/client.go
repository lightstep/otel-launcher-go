// Copyright The OpenTelemetry Authors
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

package otelcol

import (
	"context"
	"time"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/export"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/internal"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/otelarrowexporter"
	"github.com/open-telemetry/otel-arrow/collector/processor/concurrentbatchprocessor"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/otel"
	metricapi "go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	traceapi "go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/multierr"
)

type Option func(*Config)

// TODO: Config, Option, and the option impls are duplicated between
// this package and the metric exporter.  Fix this.
type Config struct {
	SelfMetrics bool
	SelfSpans   bool
	Batcher     concurrentbatchprocessor.Config
	Exporter    otelarrowexporter.Config
}

type client struct {
	internal.ResourceMap

	exporter exporter.Metrics
	batcher  processor.Metrics
	settings exporter.Settings

	// self-observability
	tracer  traceapi.Tracer
	counter metricapi.Int64Counter
}

func NewDefaultConfig() Config {
	cfg := Config{
		SelfMetrics: true,
		SelfSpans:   true,
		Batcher: concurrentbatchprocessor.Config{
			Timeout:          0,
			SendBatchSize:    1000,
			SendBatchMaxSize: 1500,
		},
		Exporter: otelarrowexporter.Config{
			TimeoutSettings: exporterhelper.TimeoutConfig{
				Timeout: 15 * time.Second,
			},
			RetryConfig:   configretry.NewDefaultBackOffConfig(),
			QueueSettings: exporterhelper.NewDefaultQueueConfig(),
			ClientConfig: configgrpc.ClientConfig{
				Headers:         map[string]configopaque.String{},
				Compression:     configcompression.TypeZstd,
				WriteBufferSize: 512 * 1024,
				WaitForReady:    true,
			},
			Arrow: otelarrowexporter.ArrowConfig{
				Disabled:         true,
				NumStreams:       1,
				DisableDowngrade: true,
			},
		},
	}
	// All RetryConfig and QueueSettings are taken from the defaults, but disabled.
	cfg.Exporter.RetryConfig.Enabled = false
	cfg.Exporter.QueueSettings.Enabled = false
	return cfg
}

func NewConfig(opts ...Option) Config {
	cfg := NewDefaultConfig()
	for _, option := range opts {
		if option != nil {
			option(&cfg)
		}
	}
	return cfg
}

func WithEndpoint(addr string) Option {
	return func(cfg *Config) {
		cfg.Exporter.ClientConfig.Endpoint = addr
	}
}

func WithHeaders(hdrs map[string]string) Option {
	return func(cfg *Config) {
		for key, val := range hdrs {
			cfg.Exporter.ClientConfig.Headers[key] = configopaque.String(val)
		}
	}
}

func WithCompressor(comp string) Option {
	return func(cfg *Config) {
		cfg.Exporter.ClientConfig.Compression = configcompression.Type(comp)
	}
}

func WithInsecure() Option {
	return func(cfg *Config) {
		cfg.Exporter.ClientConfig.TLSSetting.Insecure = true
	}
}

func WithTLSSetting(tlss configtls.ClientConfig) Option {
	return func(cfg *Config) {
		cfg.Exporter.ClientConfig.TLSSetting = tlss
	}
}

func NewExporter(ctx context.Context, cfg Config) (metric.PushExporter, error) {
	c := &client{}

	if !cfg.Exporter.Arrow.Disabled {
		c.settings.ID = component.NewID(component.MustNewType("otel_sdk_metric_arrow"))
	} else {
		c.settings.ID = component.NewID(component.MustNewType("otel_sdk_metric_otlp"))
	}

	var mp metricapi.MeterProvider = metricnoop.NewMeterProvider()
	var tp traceapi.TracerProvider = tracenoop.NewTracerProvider()

	if cfg.SelfSpans {
		tp = otel.GetTracerProvider()
	}
	if cfg.SelfMetrics {
		mp = otel.GetMeterProvider()
	}

	if settings, tracer, counter, err :=
		internal.ConfigureSelfTelemetry("lightstep-go/sdk/metric", tp, mp); err != nil {
		return nil, err
	} else {
		c.settings.TelemetrySettings = settings
		c.tracer = tracer
		c.counter = counter
	}

	exp, err := otelarrowexporter.NewFactory().CreateMetrics(ctx, c.settings, &cfg.Exporter)
	if err != nil {
		return nil, err
	}

	bset := processor.Settings{
		ID:                component.NewID(component.MustNewType("otel_sdk_metric_batch")),
		TelemetrySettings: c.settings.TelemetrySettings,
		BuildInfo:         c.settings.BuildInfo,
	}

	bat, err := concurrentbatchprocessor.NewFactory().CreateMetrics(ctx, bset, &cfg.Batcher, exp)
	if err != nil {
		return nil, err
	}

	c.exporter = exp
	c.batcher = bat

	err = exp.Start(ctx, c)
	if err != nil {
		return nil, err
	}

	err = bat.Start(ctx, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *client) String() string {
	return "otel-arrow-adapter/metricsexporter"
}

// ExportMetrics implements PushExporter.
func (c *client) ExportMetrics(ctx context.Context, data data.Metrics) error {
	return export.ExportMetrics(
		ctx,
		data,
		c.tracer,
		c.counter,
		&c.ResourceMap,
		c.batcher,
		true, // use exponential histograms
	)
}

// ShutdownMetrics implements PushExporter.
func (c *client) ShutdownMetrics(ctx context.Context, data data.Metrics) error {
	var err error
	if err1 := c.ForceFlushMetrics(ctx, data); err1 != nil {
		err = multierr.Append(err, err1)
	}
	if err2 := c.batcher.Shutdown(ctx); err2 != nil {
		err = multierr.Append(err, err2)
	}
	if err3 := c.exporter.Shutdown(ctx); err3 != nil {
		err = multierr.Append(err, err3)
	}
	return err
}

// ForceFlushMetrics implements PushExporter.
func (c *client) ForceFlushMetrics(ctx context.Context, data data.Metrics) error {
	return c.ExportMetrics(ctx, data)
}

// GetExtensions implements component.Host.
func (c *client) GetExtensions() map[component.ID]component.Component {
	return nil
}
