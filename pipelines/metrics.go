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
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp/otlpmetric"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"

	// The v1 instrumentation
	lightstepCputime "github.com/lightstep/otel-launcher-go/lightstep/instrumentation/cputime"
	lightstepHost "github.com/lightstep/otel-launcher-go/lightstep/instrumentation/host"
	lightstepRuntime "github.com/lightstep/otel-launcher-go/lightstep/instrumentation/runtime"

	// The v0 instrumentation
	contribHost "go.opentelemetry.io/contrib/instrumentation/host"
	contribRuntime "go.opentelemetry.io/contrib/instrumentation/runtime"

	// OTel APIs
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	metricglobal "go.opentelemetry.io/otel/metric/global"

	// The otel Metrics SDK
	otelotlpmetricgrpc "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	otelsdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/grpc/encoding/gzip"
)

func prestableHostMetrics(provider metric.MeterProvider) error {
	return contribHost.Start(contribHost.WithMeterProvider(provider))
}

func prestableRuntimeMetrics(provider metric.MeterProvider) error {
	return contribRuntime.Start(contribRuntime.WithMeterProvider(provider))
}

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

const prestableVersion = "prestable"
const defaultVersion = "stable"

var builtinMetricsVersions = map[string]map[string][]initFunc{
	"all": {
		defaultVersion:   libraries(stableHostMetrics, stableRuntimeMetrics, stableCputimeMetrics),
		prestableVersion: libraries(prestableHostMetrics, prestableRuntimeMetrics),
	},
	"cputime": {
		defaultVersion:   libraries(stableCputimeMetrics),
		prestableVersion: libraries(),
	},
	"host": {
		defaultVersion:   libraries(stableHostMetrics),
		prestableVersion: libraries(prestableHostMetrics),
	},
	"runtime": {
		defaultVersion:   libraries(stableRuntimeMetrics),
		prestableVersion: libraries(prestableRuntimeMetrics),
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

	lsPref, otelPref, err := tempoOptions(c)
	lsSecure, otelSecure := c.secureMetricOption()
	if err != nil {
		return nil, fmt.Errorf("invalid metric view configuration: %v", err)
	}

	if c.UseLightstepMetricsSDK {
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

	} else {
		// Install the OTel-Go community metrics SDK.
		metricExporter, err := c.newOtelMetricsExporter(otelPref, otelSecure)
		if err != nil {
			return nil, fmt.Errorf("failed to create metric exporter: %v", err)
		}
		meterProvider := otelsdkmetric.NewMeterProvider(
			otelsdkmetric.WithResource(c.Resource),
			otelsdkmetric.WithReader(otelsdkmetric.NewPeriodicReader(
				metricExporter,
				otelsdkmetric.WithInterval(period),
			)),
		)
		metricglobal.SetMeterProvider(meterProvider)
		provider = meterProvider
		shutdown = func() error {
			return meterProvider.Shutdown(context.Background())
		}
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

	metricglobal.SetMeterProvider(provider)
	return shutdown, nil
}

func (c PipelineConfig) newClient(secure otlpmetricgrpc.Option) (otlpmetric.Client, error) {
	return otlpmetricgrpc.NewClient(
		context.Background(),
		secure,
		otlpmetricgrpc.WithEndpoint(c.Endpoint),
		otlpmetricgrpc.WithHeaders(c.Headers),
		otlpmetricgrpc.WithCompressor(gzip.Name),
	)
}

func (c PipelineConfig) newMetricsExporter(secure otlpmetricgrpc.Option) (*otlpmetric.Exporter, error) {
	client, err := c.newClient(secure)
	if err != nil {
		return nil, err
	}
	return otlpmetric.New(client), nil
}

func (c PipelineConfig) newOtelMetricsExporter(temporality otelsdkmetric.TemporalitySelector, secureOpt otelotlpmetricgrpc.Option) (otelsdkmetric.Exporter, error) {
	return otelotlpmetricgrpc.New(
		context.Background(),
		secureOpt,
		otelotlpmetricgrpc.WithTemporalitySelector(temporality),
		otelotlpmetricgrpc.WithEndpoint(c.Endpoint),
		otelotlpmetricgrpc.WithHeaders(c.Headers),
		otelotlpmetricgrpc.WithCompressor(gzip.Name),
	)
}

func tempoOptions(c PipelineConfig) (view.Option, otelsdkmetric.TemporalitySelector, error) {
	syncPref := aggregation.CumulativeTemporality
	asyncPref := aggregation.CumulativeTemporality
	var otelSelector otelsdkmetric.TemporalitySelector

	switch lower := strings.ToLower(c.TemporalityPreference); lower {
	case "delta":
		// Delta means exercising the cumulative-to-delta
		// export path.  This is an unusual setting for
		// Lightstep users to choose.
		syncPref = aggregation.DeltaTemporality
		asyncPref = aggregation.DeltaTemporality

		// Note: the following is incorrect for UpDownCounter
		// and async UpDownCounter, which the OTel
		// specification stipulates are not affected by the
		// preference setting.  We WILL NOT FIX this defect.
		// Instead, as otel-launcher-go v1.10.x will use the
		// Lightstep metrics SDK by default.
		otelSelector = func(otelsdkmetric.InstrumentKind) metricdata.Temporality {
			return metricdata.DeltaTemporality
		}
	case "stateless":
		// asyncPref set above.
		syncPref = aggregation.DeltaTemporality

		otelSelector = func(otelsdkmetric.InstrumentKind) metricdata.Temporality {
			return metricdata.DeltaTemporality
		}
	case "", "cumulative":
		// syncPref, asyncPref set above.
		otelSelector = func(otelsdkmetric.InstrumentKind) metricdata.Temporality {
			return metricdata.CumulativeTemporality
		}
	default:
		return nil, nil, fmt.Errorf("invalid temporality preference: %v", c.TemporalityPreference)

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
	), otelSelector, nil
}
