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

func getTotalizedAttributeName(n string) string {
	x := strings.Split(n, ".")
	// It's a plural, make it singular.
	switch x[len(x)-1] {
	case "cycles":
		return "cycle"
	case "usage":
		return "class"
	}
	panic(fmt.Sprint("unrecognized attribute name: ", n))
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
	case "bytes":
		return s + "usage"
	case "cpu-seconds":
		return s + "time"
	default:
		panic("unrecognized metric suffix")
	}
}

func (r *builtinRuntime) register() error {
	all := r.allFunc()
	totals := map[string]bool{}
	counts := map[string]int{}
	toName := func(in string) (string, string) {
		n, statedUnits, _ := strings.Cut(in, ":")
		n = "process.runtime.go" + strings.ReplaceAll(n, "/", ".")
		return n, statedUnits
	}

	for _, m := range all {
		name, _ := toName(m.Name)

		// Totals map includes the '.' suffix.
		if strings.HasSuffix(name, ".total") {
			totals[name[:len(name)-len("total")]] = true
		}

		counts[name]++
	}

	var samples []metrics.Sample
	var instruments []instrument.Asynchronous
	var totalAttrs [][]attribute.KeyValue

	for _, m := range all {
		n, statedUnits := toName(m.Name)

		if strings.HasSuffix(n, ".total") {
			continue
		}

		var u string
		switch statedUnits {
		case "bytes", "seconds":
			// Real units
			u = statedUnits
		default:
			// Pseudo-units
			u = "{" + statedUnits + "}"
		}

		// Remove any ".total" suffix, this is redundant for Prometheus.
		var totalAttrVal string
		for totalize := range totals {
			if strings.HasPrefix(n, totalize) {
				// Units is unchanged.
				// Name becomes the overall prefix.
				// Remember which attribute to use.
				totalAttrVal = n[len(totalize):]
				n = getTotalizedMetricName(totalize[:len(totalize)-1], u)
				break
			}
		}

		if counts[n] > 1 {
			if totalAttrVal != "" {
				// This has not happened, hopefully never will.
				// Indicates the special case for objects/bytes
				// overlaps with the special case for total.
				panic("special case collision")
			}

			// This is treated as a special case, we know this happens
			// with "objects" and "bytes" in the standard Go 1.19 runtime.
			switch statedUnits {
			case "objects":
				// In this case, use `.objects` suffix.
				n = n + ".objects"
				u = "{objects}"
			case "bytes":
				// In this case, use no suffix.  In Prometheus this will
				// be appended as a suffix.
			default:
				panic(fmt.Sprint(
					"unrecognized duplicate metrics names, ",
					"attention required: ",
					n,
				))
			}
		}

		opts := []instrument.Option{
			instrument.WithUnit(unit.Unit(u)),
			instrument.WithDescription(m.Description),
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
		if totalAttrVal == "" {
			totalAttrs = append(totalAttrs, nil)
		} else {
			// Append a singleton list.
			totalAttrs = append(totalAttrs, []attribute.KeyValue{
				attribute.String(getTotalizedAttributeName(n), totalAttrVal),
			})
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
