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

package cputime // import "github.com/lightstep/otel-launcher-go/lightstep/instrumentation/cputime"

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
)

// processStartTime should be initialized before the first GC, ideally.
var processStartTime = time.Now()

// config contains optional settings for reporting host metrics.
type config struct {
	// MeterProvider sets the metric.MeterProvider.  If nil, the global
	// Provider will be used.
	MeterProvider metric.MeterProvider
}

// Option supports configuring optional settings for host metrics.
type Option interface {
	apply(*config)
}

// WithMeterProvider sets the Metric implementation to use for
// reporting.  If this option is not used, the global metric.MeterProvider
// will be used.  `provider` must be non-nil.
func WithMeterProvider(provider metric.MeterProvider) Option {
	return metricProviderOption{provider}
}

type metricProviderOption struct{ metric.MeterProvider }

func (o metricProviderOption) apply(c *config) {
	if o.MeterProvider != nil {
		c.MeterProvider = o.MeterProvider
	}
}

// Attribute sets.
var (
	// Attribute sets for CPU time measurements.

	AttributeCPUTimeUser   = []attribute.KeyValue{attribute.String("state", "user")}
	AttributeCPUTimeSystem = []attribute.KeyValue{attribute.String("state", "system")}
)

// newConfig computes a config from a list of Options.
func newConfig(opts ...Option) config {
	c := config{
		MeterProvider: global.MeterProvider(),
	}
	for _, opt := range opts {
		opt.apply(&c)
	}
	return c
}

// Start initializes reporting of host metrics using the supplied config.
func Start(opts ...Option) error {
	cfg := newConfig(opts...)
	if cfg.MeterProvider == nil {
		cfg.MeterProvider = global.MeterProvider()
	}
	c, err := newCputime(cfg)
	if err != nil {
		return err
	}
	return c.register()
}

func (c *cputime) register() error {
	var (
		err error

		processCPUTime   asyncfloat64.Counter
		processUptime    asyncfloat64.UpDownCounter
		processGCCPUTime asyncfloat64.Counter
	)

	if processCPUTime, err = c.meter.AsyncFloat64().Counter(
		"process.cpu.time",
		instrument.WithUnit("s"),
		instrument.WithDescription(
			"Accumulated CPU time spent by this process attributed by state (User, System, ...)",
		),
	); err != nil {
		return err
	}

	if processUptime, err = c.meter.AsyncFloat64().UpDownCounter(
		"process.uptime",
		instrument.WithUnit("s"),
		instrument.WithDescription("Seconds since application was initialized"),
	); err != nil {
		return err
	}

	if processGCCPUTime, err = c.meter.AsyncFloat64().UpDownCounter(
		// Note: this name is selected so that if Go's runtime/metrics package
		// were to start generating this it would be named /gc/cpu/time:seconds (float64).
		"process.runtime.go.gc.cpu.time",
		instrument.WithUnit("s"),
		instrument.WithDescription("Seconds of garbage collection since application was initialized"),
	); err != nil {
		return err
	}

	return c.meter.RegisterCallback(
		[]instrument.Asynchronous{
			processCPUTime,
			processUptime,
			processGCCPUTime,
		},
		func(ctx context.Context) {
			processUser, processSystem, processGC, uptime := c.getProcessTimes(ctx)

			// Uptime
			processUptime.Observe(ctx, uptime)

			// Process CPU time
			processCPUTime.Observe(ctx, processUser, AttributeCPUTimeUser...)
			processCPUTime.Observe(ctx, processSystem, AttributeCPUTimeSystem...)

			// Process GC CPU time
			processGCCPUTime.Observe(ctx, processGC)
		},
	)
}
