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
	"fmt"
	"sync"

	"github.com/f5/otel-arrow-adapter/collector/gen/exporter/otlpexporter"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/otel"
	globalmetric "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type Option func(*Config)

type Config struct {
	// Batcher  batchprocessor.Config
	Exporter otlpexporter.Config
}

type client struct {
	once     sync.Once
	resource pcommon.Resource

	exporter exporter.Traces
	// batcher  processor.Traces
	settings exporter.CreateSettings
}

func NewDefaultConfig() Config {
	return Config{
		// Batcher: batchprocessor.Config{
		// 	Timeout:          0,
		// 	SendBatchSize:    0,
		// 	SendBatchMaxSize: 10,
		// },
		Exporter: otlpexporter.Config{
			TimeoutSettings: exporterhelper.NewDefaultTimeoutSettings(),
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
			},
			Arrow: otlpexporter.ArrowSettings{
				NumStreams: 1,
				Enabled:    true,
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

func NewExporter(ctx context.Context, cfg Config) (trace.SpanExporter, error) {
	c := &client{}

	if cfg.Exporter.Arrow.Enabled {
		c.settings.ID = component.NewID("otlp/arrow")
	} else {
		c.settings.ID = component.NewID("otlp/proto")
	}
	// @@@
	// logger, err := zap.NewProduction()
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	c.settings.TelemetrySettings.Logger = logger
	c.settings.TelemetrySettings.TracerProvider = otel.GetTracerProvider()
	c.settings.TelemetrySettings.MeterProvider = globalmetric.MeterProvider() // Note: becomes otel.GetMeterProvider()
	c.settings.TelemetrySettings.MetricsLevel = configtelemetry.LevelNormal

	exp, err := otlpexporter.NewFactory().CreateTracesExporter(ctx, c.settings, &cfg.Exporter)
	if err != nil {
		return nil, err
	}

	// bset := processor.CreateSettings{
	// 	ID:                component.NewID("otlp/batch"),
	// 	TelemetrySettings: c.settings.TelemetrySettings,
	// 	BuildInfo:         c.settings.BuildInfo,
	// }

	// bat, err := batchprocessor.NewFactory().CreateTracesProcessor(ctx, bset, &cfg.Batcher, exp)
	// if err != nil {
	// 	return nil, err
	// }

	c.exporter = exp
	// c.batcher = bat

	fmt.Println("START, with config", cfg)

	err = exp.Start(ctx, c)
	if err != nil {
		return nil, err
	}

	// err = bat.Start(ctx, c)
	// if err != nil {
	// 	return nil, err
	// }

	return c, nil
}

func (c *client) String() string {
	return "otel-arrow-adapter/tracesexporter"
}

// ExportSpans implements PushExporter.
func (c *client) ExportSpans(ctx context.Context, data []trace.ReadOnlySpan) error {
	fmt.Println("Export spans start")
	defer fmt.Println("Export spans done")
	// @@@ return c.batcher.ConsumeSpans(ctx, c.d2pd(data))
	return c.exporter.ConsumeTraces(ctx, c.d2pd(data))
}

// ShutdownSpans implements PushExporter.
func (c *client) Shutdown(ctx context.Context) error {
	var err error
	// if err2 := c.batcher.Shutdown(ctx); err2 != nil {
	// 	err = multierr.Append(err, err2)
	// }
	if err3 := c.exporter.Shutdown(ctx); err3 != nil {
		err = multierr.Append(err, err3)
	}
	return err
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
