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

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/internal"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/open-telemetry/otel-arrow/collector/exporter/otelarrowexporter"
	"github.com/open-telemetry/otel-arrow/collector/processor/concurrentbatchprocessor"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	metricapi "go.opentelemetry.io/otel/metric"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	traceapi "go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/multierr"
	"go.uber.org/zap"
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
	settings exporter.CreateSettings

	// self-observability
	tracer  traceapi.Tracer
	counter metricapi.Int64Counter
}

func NewDefaultConfig() Config {
	return Config{
		SelfMetrics: true,
		SelfSpans:   true,
		Batcher: concurrentbatchprocessor.Config{
			Timeout:          0,
			SendBatchSize:    1000,
			SendBatchMaxSize: 1500,
			MaxInFlightBytes: 32 * 1024 * 1024,
		},
		Exporter: otelarrowexporter.Config{
			TimeoutSettings: exporterhelper.TimeoutSettings{
				Timeout: 15 * time.Second,
			},
			RetrySettings: exporterhelper.RetrySettings{
				Enabled: false,
			},
			QueueSettings: exporterhelper.QueueSettings{
				Enabled: false,
			},
			GRPCClientSettings: configgrpc.GRPCClientSettings{
				Headers:         map[string]configopaque.String{},
				Compression:     configcompression.Zstd,
				WriteBufferSize: 512 * 1024,
				WaitForReady:    true,
			},
			Arrow: otelarrowexporter.ArrowSettings{
				Disabled:         true,
				NumStreams:       1,
				DisableDowngrade: true,
			},
		},
	}
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
		cfg.Exporter.GRPCClientSettings.Endpoint = addr
	}
}

func WithHeaders(hdrs map[string]string) Option {
	return func(cfg *Config) {
		for key, val := range hdrs {
			cfg.Exporter.GRPCClientSettings.Headers[key] = configopaque.String(val)
		}
	}
}

func WithCompressor(comp string) Option {
	return func(cfg *Config) {
		cfg.Exporter.GRPCClientSettings.Compression = configcompression.CompressionType(comp)
	}
}

func WithInsecure() Option {
	return func(cfg *Config) {
		cfg.Exporter.GRPCClientSettings.TLSSetting.Insecure = true
	}
}

func WithTLSSetting(tlss configtls.TLSClientSetting) Option {
	return func(cfg *Config) {
		cfg.Exporter.GRPCClientSettings.TLSSetting = tlss
	}
}

func NewExporter(ctx context.Context, cfg Config) (metric.PushExporter, error) {
	c := &client{}

	if !cfg.Exporter.Arrow.Disabled {
		c.settings.ID = component.NewID("otel/sdk/metric/arrow")
	} else {
		c.settings.ID = component.NewID("otel/sdk/metric/otlp")
	}
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	c.settings.TelemetrySettings.Logger = logger

	if cfg.SelfSpans {
		c.settings.TelemetrySettings.TracerProvider = otel.GetTracerProvider()
		c.tracer = c.settings.TelemetrySettings.TracerProvider.Tracer("lightstep-go/sdk/metric")
	} else {
		c.settings.TelemetrySettings.TracerProvider = nooptrace.NewTracerProvider()
	}

	if cfg.SelfMetrics {
		c.settings.TelemetrySettings.MeterProvider = otel.GetMeterProvider()
		c.settings.TelemetrySettings.MetricsLevel = configtelemetry.LevelNormal
		c.counter, _ = c.settings.TelemetrySettings.MeterProvider.Meter("lightstep/sdk/metric").Int64Counter("otelsdk.telemetry.items")
	} else {
		c.settings.TelemetrySettings.MeterProvider = noopmetric.NewMeterProvider()
	}

	exp, err := otelarrowexporter.NewFactory().CreateMetricsExporter(ctx, c.settings, &cfg.Exporter)
	if err != nil {
		return nil, err
	}

	bset := processor.CreateSettings{
		ID:                component.NewID("otel/sdk/metric/batch"),
		TelemetrySettings: c.settings.TelemetrySettings,
		BuildInfo:         c.settings.BuildInfo,
	}

	bat, err := concurrentbatchprocessor.NewFactory().CreateMetricsProcessor(ctx, bset, &cfg.Batcher, exp)
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
	ctx, span := c.tracer.Start(
		ctx,
		"otelsdk_export_metrics",
	)
	converted := c.d2pd(data)
	points := int64(converted.DataPointCount())

	defer span.End()

	err := c.batcher.ConsumeMetrics(ctx, converted)

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
	c.counter.Add(ctx, points, metricapi.WithAttributes(attrs...))
	span.SetAttributes(append(attrs, attribute.Int64("num_points", points))...)
	if err == nil {
		span.SetStatus(otelcodes.Ok, state)
	} else {
		span.SetStatus(otelcodes.Error, state)
	}
	return err
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

// ReportFatalError implements component.Host.
func (c *client) ReportFatalError(err error) {
	c.settings.Logger.Fatal("exporter fatal", zap.Error(err))
}

// GetFactory implements component.Host.
func (c *client) GetFactory(component.Kind, component.Type) component.Factory {
	return nil
}

// GetExtensions implements component.Host.
func (c *client) GetExtensions() map[component.ID]component.Component {
	return nil
}

// GetExporters implements component.Host.
func (c *client) GetExporters() map[component.DataType]map[component.ID]component.Component {
	return nil
}
