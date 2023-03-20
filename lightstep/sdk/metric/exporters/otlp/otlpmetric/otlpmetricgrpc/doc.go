package otlpmetricgrpc

import (
	"context"

	"github.com/f5/otel-arrow-adapter/collector/gen/exporter/otlpexporter"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type Option func(*Config)

type Config = otlpexporter.Config

type client struct {
}

func NewConfig(opts ...Option) Config {
	cfg := otlpexporter.Config{
		TimeoutSettings: exporterhelper.NewDefaultTimeoutSettings(),
		RetrySettings:   exporterhelper.NewDefaultRetrySettings(),
		QueueSettings:   exporterhelper.NewDefaultQueueSettings(),
		GRPCClientSettings: configgrpc.GRPCClientSettings{
			Headers:         map[string]configopaque.String{},
			Compression:     configcompression.Zstd,
			WriteBufferSize: 512 * 1024,
		},
		Arrow: otlpexporter.ArrowSettings{
			NumStreams: 1,
			Enabled:    true,
		},
	}
	for _, option := range opts {
		option(&cfg)
	}

	return cfg
}

func WithEndpoint(addr string) Option {
	return func(cfg *Config) {
		cfg.GRPCClientSettings.Endpoint = addr
	}
}

func WithHeaders(hdrs map[string]string) Option {
	return func(cfg *Config) {
		for key, val := range hdrs {
			cfg.GRPCClientSettings.Headers[key] = configopaque.String(val)
		}
	}
}

func WithCompressor(comp string) Option {
	return func(cfg *Config) {
		cfg.GRPCClientSettings.Compression = configcompression.CompressionType(comp)
	}
}

func WithInsecure() Option {
	return func(cfg *Config) {
		cfg.GRPCClientSettings.TLSSetting.Insecure = true
	}
}

func WithTLSSetting(tlss configtls.TLSClientSetting) Option {
	return func(cfg *Config) {
		cfg.GRPCClientSettings.TLSSetting = tlss
	}
}

func NewClient(ctx context.Context, cfg *Config) (metric.PushExporter, error) {
	var set exporter.CreateSettings
	if cfg.Arrow.Enabled {
		set.ID = component.NewID("otlp/arrow")
	} else {
		set.ID = component.NewID("otlp/proto")
	}
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	set.TelemetrySettings.Logger = logger
	set.TelemetrySettings.MeterProvider = otel.GetMeterProvider()
	set.TelemetrySettings.MetricsLevel = configtelemetry.LevelNormal

	mexp, err := otlpexporter.NewFactory().CreateMetricsExporter(ctx, set, cfg)
	return &client{}, nil
}

func (c *client) String() string {
	return "otel-arrow-adapter/otlpexporter"
}

func (c *client) ExportMetrics(ctx context.Context, data data.Metrics) error {

}

func (c *client) ShutdownMetrics(ctx context.Context, data data.Metrics) error {
	err := c.ForceFlushMetrics(ctx, data)
	if err != nil {
		// @@@
	}
	return nil
}

func (c *client) ForceFlushMetrics(ctx context.Context, data data.Metrics) error {
	return c.ExportMetrics(ctx, data)
}
