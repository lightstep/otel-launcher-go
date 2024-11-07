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
	"errors"
	"time"

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
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	metricapi "go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/sdk/trace"
	traceapi "go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/multierr"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/internal"
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

type ExporterOptions struct {
	TracerProvider traceapi.TracerProvider
	MeterProvider  metricapi.MeterProvider
}

func WithTracerProvider(tp traceapi.TracerProvider) func(*ExporterOptions) {
	return func(opts *ExporterOptions) {
		opts.TracerProvider = tp
	}
}

func WithMeterProvider(mp metricapi.MeterProvider) func(*ExporterOptions) {
	return func(opts *ExporterOptions) {
		opts.MeterProvider = mp
	}
}

type client struct {
	internal.ResourceMap

	exporter exporter.Traces
	batcher  processor.Traces
	settings exporter.Settings
	tracer   traceapi.Tracer
	counter  metricapi.Int64Counter
}

func (c *client) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	ctx, span := c.tracer.Start(
		ctx,
		"otelsdk_export_traces",
	)
	defer span.End()

	converted := c.d2pd(spans)
	count := converted.SpanCount()

	err := c.batcher.ConsumeTraces(ctx, converted)
	success := err == nil
	var state string
	if success {
		state = "ok"
	} else if errors.Is(err, context.Canceled) {
		state = "canceled"
	} else if errors.Is(err, context.DeadlineExceeded) {
		state = "timeout"
	} else {
		state = "error"
	}

	var attrs = []attribute.KeyValue{
		attribute.Bool("success", success),
		attribute.String("state", state),
	}
	c.counter.Add(ctx, int64(count), metricapi.WithAttributes(attrs...))
	span.SetAttributes(append(attrs, attribute.Int("num_spans", count))...)
	if err == nil {
		span.SetStatus(otelcodes.Ok, state)
	} else {
		span.SetStatus(otelcodes.Error, err.Error())
	}
	return err
}

func (c *client) Shutdown(ctx context.Context) error {
	var err error
	if err1 := c.batcher.Shutdown(ctx); err1 != nil {
		err = multierr.Append(err, err1)
	}
	if err2 := c.exporter.Shutdown(ctx); err2 != nil {
		err = multierr.Append(err, err2)
	}
	return err
}

func NewDefaultConfig() Config {
	cfg := Config{
		SelfMetrics: true,
		SelfSpans:   true,
		Batcher: concurrentbatchprocessor.Config{
			Timeout:          time.Second,
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

func NewExporter(ctx context.Context, cfg Config, opts ...func(options *ExporterOptions)) (trace.SpanExporter, error) {
	options := ExporterOptions{
		MeterProvider:  nil,
		TracerProvider: nil,
	}
	for _, opt := range opts {
		opt(&options)
	}

	c := &client{}

	if !cfg.Exporter.Arrow.Disabled {
		c.settings.ID = component.NewID(component.MustNewType("otel_sdk_trace_arrow"))
	} else {
		c.settings.ID = component.NewID(component.MustNewType("otel_sdk_trace_otlp"))
	}

	if options.MeterProvider == nil {
		if cfg.SelfMetrics {
			options.MeterProvider = otel.GetMeterProvider()
		} else {
			options.MeterProvider = metricnoop.NewMeterProvider()
		}
	}
	if options.TracerProvider == nil {
		if cfg.SelfSpans {
			options.TracerProvider = otel.GetTracerProvider()
		} else {
			options.TracerProvider = tracenoop.NewTracerProvider()
		}
	}

	if settings, tracer, counter, err :=
		internal.ConfigureSelfTelemetry(
			"lightstep-go/sdk/trace",
			options.TracerProvider,
			options.MeterProvider); err != nil {
		return nil, err
	} else {
		c.settings.TelemetrySettings = settings
		c.tracer = tracer
		c.counter = counter
	}

	exp, err := otelarrowexporter.NewFactory().CreateTraces(ctx, c.settings, &cfg.Exporter)
	if err != nil {
		return nil, err
	}

	bset := processor.Settings{
		ID:                component.NewID(component.MustNewType("otel_sdk_trace_batch")),
		TelemetrySettings: c.settings.TelemetrySettings,
		BuildInfo:         c.settings.BuildInfo,
	}

	bat, err := concurrentbatchprocessor.NewFactory().CreateTraces(ctx, bset, &cfg.Batcher, exp)
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

// GetExtensions implements component.Host.
func (c *client) GetExtensions() map[component.ID]component.Component {
	return nil
}
