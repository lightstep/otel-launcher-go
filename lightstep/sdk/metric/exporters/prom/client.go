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

package prom

import (
	"context"
	"fmt"
	"time"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/internal"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/export"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/otel"
	metricapi "go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	traceapi "go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/multierr"
)

// TODO: Config, Option, and the option impls are duplicated between
// this package and the metric exporter.  Fix this.
type Config struct {
	SelfMetrics bool
	SelfSpans   bool
	Exporter    prometheusexporter.Config
}

type client struct {
	internal.ResourceMap

	exporter exporter.Metrics
	settings exporter.Settings

	// self-observability
	tracer  traceapi.Tracer
	counter metricapi.Int64Counter
}

var _ metric.PushExporter = &client{}
var _ component.Host = &client{}

type Option func(*Config)

func NewConfig(opts ...Option) Config {
	cfg := NewDefaultConfig()
	for _, option := range opts {
		if option != nil {
			option(&cfg)
		}
	}
	return cfg
}

func NewDefaultConfig() Config {
	cfg := Config{
		SelfMetrics: true,
		SelfSpans:   true,

		Exporter: prometheusexporter.Config{
			ServerConfig: confighttp.ServerConfig{
				// 9090 is the default port that prometheus metrics should be served on
				Endpoint: "0.0.0.0:9090",
			},
			MetricExpiration:            time.Hour,
			ResourceToTelemetrySettings: resourcetotelemetry.Settings{Enabled: true},
		},
	}
	return cfg
}

func WithPort(port int) Option {
	return func(c *Config) {
		c.Exporter.Endpoint = fmt.Sprintf("0.0.0.0:%d", port)
	}
}

// NewExporter creates a new prometheus exporter from the provided config.
//
// N.B. that before self telemetry can be reported a global MeterProvider and TracerProvider must be configured.
func NewExporter(ctx context.Context, cfg Config) (metric.PushExporter, error) {
	c := &client{}

	c.settings.ID = component.NewID(component.MustNewType("otel_sdk_metric_prom"))

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

	exp, err := prometheusexporter.NewFactory().CreateMetrics(ctx, c.settings, &cfg.Exporter)
	if err != nil {
		return nil, err
	}

	c.exporter = exp

	err = exp.Start(ctx, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *client) String() string {
	return "prometheus_exporter"
}

// ExportMetrics implements PushExporter.
func (c *client) ExportMetrics(ctx context.Context, data data.Metrics) error {
	return export.ExportMetrics(
		ctx,
		data,
		c.tracer,
		c.counter,
		&c.ResourceMap,
		c.exporter,
		// don't use exponential histograms, since the prometheus exporter doesn't support them
		false,
	)
}

// ShutdownMetrics implements PushExporter.
func (c *client) ShutdownMetrics(ctx context.Context, data data.Metrics) error {
	var err error
	if err1 := c.ForceFlushMetrics(ctx, data); err1 != nil {
		err = multierr.Append(err, err1)
	}
	if err2 := c.exporter.Shutdown(ctx); err2 != nil {
		err = multierr.Append(err, err2)
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
