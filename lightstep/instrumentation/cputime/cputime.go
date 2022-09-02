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
	"fmt"
	"math"
	"runtime"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
)

// processStartTime should be initialized before the first GC, ideally.
var processStartTime = time.Now()

// cputime reports the work-in-progress conventional cputime metrics specified by OpenTelemetry.
type cputime struct {
	meter metric.Meter
}

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
	c := newCputime(cfg)
	return c.register()
}

func newCputime(c config) *cputime {
	return &cputime{
		meter: c.MeterProvider.Meter("otel_launcher_go/cputime"),
	}
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

	err = c.meter.RegisterCallback(
		[]instrument.Asynchronous{
			processCPUTime,
			processUptime,
			processGCCPUTime,
		},
		func(ctx context.Context) {
			processUser, processSystem, processGC, uptime := getProcessTimes()

			// Uptime
			processUptime.Observe(ctx, uptime)

			// Process CPU time
			processCPUTime.Observe(ctx, processUser, AttributeCPUTimeUser...)
			processCPUTime.Observe(ctx, processSystem, AttributeCPUTimeSystem...)

			// Process GC CPU time
			processGCCPUTime.Observe(ctx, processGC)
		})

	if err != nil {
		return err
	}

	return nil
}

// getProcessTimes calls ReadMemStats() for GCCPUFraction because as
// of Go-1.19 there is no such runtime metric.  User and system sum to
// 100% of CPU time; gc is an independent, comparable metric value.
// These are correlated with uptime.
func getProcessTimes() (userSeconds, systemSeconds, gcSeconds, uptimeSeconds float64) {
	// Would really be better if runtime/metrics exposed this,
	// making an expensive call for a single field that is not
	// exposed via ReadMemStats().
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	gomaxprocs := float64(runtime.GOMAXPROCS(0))

	uptimeSeconds = time.Since(processStartTime).Seconds()
	gcSeconds = memStats.GCCPUFraction * uptimeSeconds * gomaxprocs

	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		userSeconds = math.NaN()
		systemSeconds = math.NaN()
		otel.Handle(fmt.Errorf("getrusage: %w", err))
		return
	}

	utime := time.Duration(ru.Utime.Sec)*time.Second + time.Duration(ru.Utime.Usec)*time.Microsecond
	stime := time.Duration(ru.Stime.Sec)*time.Second + time.Duration(ru.Stime.Usec)*time.Microsecond

	userSeconds = utime.Seconds()
	systemSeconds = stime.Seconds()
	return
}
