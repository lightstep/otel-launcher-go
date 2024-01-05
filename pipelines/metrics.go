// Copyright Lightstep Authors
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

package pipelines

import (
	"context"
	"fmt"
	"strings"
	"time"

	// The Lightstep SDK
	sdkmetric "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp/otelcol"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"
	"google.golang.org/grpc/encoding/gzip"

	// The v1 instrumentation
	lightstepCputime "github.com/lightstep/otel-launcher-go/lightstep/instrumentation/cputime"
	lightstepHost "github.com/lightstep/otel-launcher-go/lightstep/instrumentation/host"
	lightstepRuntime "github.com/lightstep/otel-launcher-go/lightstep/instrumentation/runtime"

	// OTel APIs
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func stableHostMetrics(provider metric.MeterProvider) error {
	return lightstepHost.Start(lightstepHost.WithMeterProvider(provider))
}

func stableRuntimeMetrics(provider metric.MeterProvider) error {
	return lightstepRuntime.Start(lightstepRuntime.WithMeterProvider(provider))
}

func stableCputimeMetrics(provider metric.MeterProvider) error {
	return lightstepCputime.Start(lightstepCputime.WithMeterProvider(provider))
}

type initFunc func(metric.MeterProvider) error

func libraries(inits ...initFunc) []initFunc {
	return inits
}

const defaultVersion = "stable"

var builtinMetricsVersions = map[string]map[string][]initFunc{
	"all": {
		defaultVersion: libraries(stableHostMetrics, stableRuntimeMetrics, stableCputimeMetrics),
	},
	"cputime": {
		defaultVersion: libraries(stableCputimeMetrics),
	},
	"host": {
		defaultVersion: libraries(stableHostMetrics),
	},
	"runtime": {
		defaultVersion: libraries(stableRuntimeMetrics),
	},
}

func NewMetricsPipeline(c PipelineConfig) (func() error, error) {
	var err error

	period := 30 * time.Second

	if c.ReportingPeriod != "" {
		period, err = time.ParseDuration(c.ReportingPeriod)
		if err != nil {
			return nil, fmt.Errorf("invalid metric reporting period: %v", err)
		}
		if period <= 0 {
			return nil, fmt.Errorf("invalid metric reporting period: %v", c.ReportingPeriod)
		}
	}
	var provider metric.MeterProvider
	var shutdown func() error

	lsPref, err := tempoOptions(c)
	lsSecure := c.secureMetricOption()
	if err != nil {
		return nil, fmt.Errorf("invalid metric view configuration: %v", err)
	}

	// Install the Lightstep metrics SDK
	metricExporter, err := c.newMetricsExporter(lsSecure)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %v", err)
	}

	sdk := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(c.Resource),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(metricExporter, period),
			lsPref,
		),
	)

	provider = sdk
	shutdown = func() error {
		return sdk.Shutdown(context.Background())
	}

	if c.MetricsBuiltinsEnabled {
		for _, lib := range c.MetricsBuiltinLibraries {
			name, version, _ := strings.Cut(lib, ":")

			if version == "" {
				version = defaultVersion
			}

			vm, has := builtinMetricsVersions[name]
			if !has {
				otel.Handle(fmt.Errorf("unrecognized builtin: %q", name))
				continue
			}
			fs, has := vm[version]
			if !has {
				otel.Handle(fmt.Errorf("unrecognized builtin version: %v: %q", name, version))
				continue
			}
			for _, f := range fs {

				if err := f(provider); err != nil {
					otel.Handle(fmt.Errorf("failed to start %v instrumentation: %w", name, err))
				}
			}
		}
	}

	otel.SetMeterProvider(provider)
	return shutdown, nil
}

func (c PipelineConfig) newMetricsExporter(secure otelcol.Option) (sdkmetric.PushExporter, error) {
	return otelcol.NewExporter(
		context.Background(),
		otelcol.NewConfig(
			secure,
			otelcol.WithEndpoint(c.Endpoint),
			otelcol.WithHeaders(c.Headers),
			otelcol.WithCompressor(gzip.Name),
		),
	)
}

func tempoOptions(c PipelineConfig) (view.Option, error) {
	syncPref := aggregation.CumulativeTemporality
	asyncPref := aggregation.CumulativeTemporality

	switch lower := strings.ToLower(c.TemporalityPreference); lower {
	case "delta":
		// Delta means exercising the cumulative-to-delta
		// export path.  This is an unusual setting for
		// Lightstep users to choose.
		syncPref = aggregation.DeltaTemporality
		asyncPref = aggregation.DeltaTemporality
	case "stateless", "lowmemory":
		syncPref = aggregation.DeltaTemporality
	case "", "cumulative":
	default:
		return nil, fmt.Errorf("invalid temporality preference: %v", c.TemporalityPreference)

	}
	return view.WithDefaultAggregationTemporalitySelector(
		func(k sdkinstrument.Kind) aggregation.Temporality {
			switch k {
			case sdkinstrument.SyncUpDownCounter, sdkinstrument.AsyncUpDownCounter:
				return aggregation.CumulativeTemporality
			case sdkinstrument.SyncCounter, sdkinstrument.SyncHistogram:
				return syncPref
			default:
				return asyncPref
			}
		},
	), nil
}
