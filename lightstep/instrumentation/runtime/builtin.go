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
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
)

const (
	namePrefix = "process.runtime.go"
	classKey     = attribute.Key("class")
	subclassKey  = attribute.Key("subclass")
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
	return r.register()
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

func getTotalizedAttributeName(order int) string {
	// It's a plural, make it singular.
	base := "class"
	for ; order > 0; order-- {
		base = "sub" + base
	}
	return base
}

func getTotalizedMetricName(n, u string) string {
	if !strings.HasSuffix(n, ".classes") {
		return n
	}

	s := n[:len(n)-len("classes")]

	// Note that ".classes" is (apparently) intended as a generic
	// suffix, while ".cycles" is an exception.
	// The ideal name depends on what we know.
	switch u {
	case "By":
		// OTel has similar conventions for memory usage, disk usage, etc, so
		// for metrics with /classes/*:bytes we create a .usage metric.
		return s + "usage"
	case "{cpu-seconds}":
		// Same argument above, except OTel uses .time for
		// cpu-timing metrics instead of .usage.
		return s + "time"
	default:
		// There are no other conventions in the
		// runtime/metrics that we know of.
		panic(fmt.Sprint("unrecognized metric suffix: ", u))
	}
}

func (r *builtinRuntime) register() error {
	all := r.allFunc()
	desc := expectRuntimeMetrics()

	for _, m := range all {
		// each should match one
		family, err = desc.find(m.Name)
		if err != nil {
			return err
		}

		// We need a fixed description, which depends which
		// convention is used.
		var description string
		if len(family.attrs) != 1 {
			// This case is an aggregation of known values
			description = fmt.Sprintf("runtime/metrics: %v/*:%s", m.Name, statedUnits)
		} else {
			// This includes the counts[n] > 1 case, both
			// use an exact name.
			description = fmt.Sprintf("runtime/metrics: %v", m.Name)
		}

		opts := []instrument.Option{
			instrument.WithUnit(unit.Unit(u)),
			instrument.WithDescription(description),
		}
		var inst instrument.Asynchronous
		var err error
		if m.Cumulative {
			switch m.Kind {
			case metrics.KindUint64:
				inst, err = r.meter.AsyncInt64().Counter(n, opts...)
			case metrics.KindFloat64:
				inst, err = r.meter.AsyncFloat64().Counter(n, opts...)
			case metrics.KindFloat64Histogram:
				// Not implemented Histogram[float64].
				continue
			}
		} else {
			switch m.Kind {
			case metrics.KindUint64:
				inst, err = r.meter.AsyncInt64().UpDownCounter(n, opts...)
			case metrics.KindFloat64:
				// Note: this has never been used.
				inst, err = r.meter.AsyncFloat64().Gauge(n, opts...)
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

		if totalAttrVals == nil {
			totalAttrs = append(totalAttrs, nil)
		} else {
			// Form a list of attributes to use in the callback.
			var attrs []attribute.KeyValue
			for i, val := range totalAttrVals {
				attrs = append(attrs, attribute.Key(getTotalizedAttributeName(i)).String(val))
			}
			totalAttrs = append(totalAttrs, attrs)
		}
	}

	if err := r.meter.RegisterCallback(instruments, func(ctx context.Context) {
		r.readFunc(samples)

		for idx, samp := range samples {

			switch samp.Value.Kind() {
			case metrics.KindUint64:
				instruments[idx].(int64Observer).Observe(ctx, int64(samp.Value.Uint64()), totalAttrs[idx]...)
			case metrics.KindFloat64:
				instruments[idx].(float64Observer).Observe(ctx, samp.Value.Float64(), totalAttrs[idx]...)
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
