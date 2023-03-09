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

package metric // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"

import (
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
)

// config contains configuration options for a MeterProvider.
type config struct {
	// res is the resource for this MeterProvider.
	res *resource.Resource

	// readers is a slice of Reader instances corresponding with views.
	// the i'th reader uses the i'th entry in views.
	readers []Reader

	// views is a slice of *Views instances corresponding with readers.
	// the i'th views applies to the i'th reader.
	views []*view.Views

	// performance settings
	performance sdkinstrument.Performance
}

// Option applies a configuration option value to a MeterProvider.
type Option interface {
	apply(config) config
}

// optionFunction makes a functional Option out of a function object.
type optionFunction func(cfg config) config

// apply implements Option.
func (of optionFunction) apply(in config) config {
	return of(in)
}

// WithResource associates a Resource with a new MeterProvider.
func WithResource(res *resource.Resource) Option {
	return optionFunction(func(cfg config) config {
		cfg.res = res
		return cfg
	})
}

// WithReader associates a new Reader and associated View options with
// a new MeterProvider
func WithReader(r Reader, opts ...view.Option) Option {
	return optionFunction(func(cfg config) config {
		v, err := view.Validate(view.New(r.String(), opts...))
		if err != nil {
			otel.Handle(err)
		}
		cfg.readers = append(cfg.readers, r)
		cfg.views = append(cfg.views, v)
		return cfg
	})
}

// WithPerformance supports modifying performance settings.
func WithPerformance(perf sdkinstrument.Performance) Option {
	return optionFunction(func(cfg config) config {
		cfg.performance = perf
		return cfg
	})
}
