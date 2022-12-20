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

package runtime // import "github.com/lightstep/otel-launcher-go/lightstep/instrumentation/runtime"

import (
	"context"
	"fmt"
	"runtime/metrics"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
)

const (
	namePrefix     = "process.runtime.go"
	classKey       = attribute.Key("class")
	subclassKey    = attribute.Key("subclass")
	subsubclassKey = attribute.Key("subsubclass")
)

// LibraryName is the value of instrumentation.Library.Name.
const LibraryName = "otel-launcher-go/runtime"

// config contains optional settings for reporting runtime metrics.
type config struct {
	// MeterProvider sets the metric.MeterProvider.  If nil, the global
	// Provider will be used.
	MeterProvider metric.MeterProvider
}

// Option supports configuring optional settings for runtime metrics.
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

// newConfig computes a config from the supplied Options.
func newConfig(opts ...Option) config {
	c := config{
		MeterProvider: global.MeterProvider(),
	}
	for _, opt := range opts {
		opt.apply(&c)
	}
	return c
}

// Start initializes reporting of runtime metrics using the supplied config.
func Start(opts ...Option) error {
	c := newConfig(opts...)
	if c.MeterProvider == nil {
		c.MeterProvider = global.MeterProvider()
	}
	meter := c.MeterProvider.Meter(
		LibraryName,
	)

	r := newBuiltinRuntime(meter, metrics.All, metrics.Read)
	return r.register(expectRuntimeMetrics())
}

type allFunc = func() []metrics.Description
type readFunc = func([]metrics.Sample)

type builtinRuntime struct {
	meter    metric.Meter
	allFunc  allFunc
	readFunc readFunc
}

type int64Observer interface {
	Observe(ctx context.Context, x int64, attrs ...attribute.KeyValue)
}

type float64Observer interface {
	Observe(ctx context.Context, x float64, attrs ...attribute.KeyValue)
}

func newBuiltinRuntime(meter metric.Meter, af allFunc, rf readFunc) *builtinRuntime {
	return &builtinRuntime{
		meter:    meter,
		allFunc:  af,
		readFunc: rf,
	}
}

func (r *builtinRuntime) register(desc *builtinDescriptor) error {
	all := r.allFunc()

	var instruments []instrument.Asynchronous
	var samples []metrics.Sample
	var instAttrs [][]attribute.KeyValue

	for _, m := range all {
		// each should match one
		mname, munit, pattern, attrs, err := desc.findMatch(m.Name)
		if err != nil {
			return err
		}
		if mname == "" {
			continue
		}
		description := fmt.Sprintf("%s from runtime/metrics", pattern)

		opts := []instrument.Option{
			instrument.WithUnit(unit.Unit(munit)),
			instrument.WithDescription(description),
		}
		var inst instrument.Asynchronous
		if m.Cumulative {
			switch m.Kind {
			case metrics.KindUint64:
				inst, err = r.meter.AsyncInt64().Counter(mname, opts...)
			case metrics.KindFloat64:
				inst, err = r.meter.AsyncFloat64().Counter(mname, opts...)
			case metrics.KindFloat64Histogram:
				// Not implemented Histogram[float64].
				continue
			}
		} else {
			switch m.Kind {
			case metrics.KindUint64:
				inst, err = r.meter.AsyncInt64().UpDownCounter(mname, opts...)
			case metrics.KindFloat64:
				// Note: this has never been used.
				inst, err = r.meter.AsyncFloat64().Gauge(mname, opts...)
			case metrics.KindFloat64Histogram:
				// Not implemented GaugeHistogram[float64].
				continue
			}
		}
		if err != nil {
			return err
		}

		samp := metrics.Sample{
			Name: m.Name,
		}
		samples = append(samples, samp)
		instruments = append(instruments, inst)
		instAttrs = append(instAttrs, attrs)
	}

	if err := r.meter.RegisterCallback(instruments, func(ctx context.Context) {
		r.readFunc(samples)

		for idx, samp := range samples {

			switch samp.Value.Kind() {
			case metrics.KindUint64:
				instruments[idx].(int64Observer).Observe(ctx, int64(samp.Value.Uint64()), instAttrs[idx]...)
			case metrics.KindFloat64:
				instruments[idx].(float64Observer).Observe(ctx, samp.Value.Float64(), instAttrs[idx]...)
			default:
				// KindFloat64Histogram (unsupported in OTel) and KindBad
				// (unsupported by runtime/metrics).  Neither should happen
				// if runtime/metrics and the code above are working correctly.
				otel.Handle(fmt.Errorf("invalid runtime/metrics value kind: %v", samp.Value.Kind()))
			}
		}
	}); err != nil {
		return err
	}
	return nil
}
