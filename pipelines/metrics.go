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

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"
	hostMetrics "go.opentelemetry.io/contrib/instrumentation/host"
	runtimeMetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	metricglobal "go.opentelemetry.io/otel/metric/global"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

func NewMetricsPipeline(c PipelineConfig) (func() error, error) {
	metricExporter, err := newMetricsExporter(c.Endpoint, c.Insecure, c.Headers)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %v", err)
	}

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
	vopts, err := viewOptions(c)
	if err != nil {
		return nil, fmt.Errorf("invalid metric view configuration: %v", err)
	}
	provider := metric.NewMeterProvider(
		metric.WithResource(c.Resource),
		metric.WithReader(
			metric.NewPeriodicReader(metricExporter, period),
			vopts...,
		),
	)

	if err = runtimeMetrics.Start(runtimeMetrics.WithMeterProvider(provider)); err != nil {
		return nil, fmt.Errorf("failed to start runtime metrics: %v", err)
	}

	if err = hostMetrics.Start(hostMetrics.WithMeterProvider(provider)); err != nil {
		return nil, fmt.Errorf("failed to start host metrics: %v", err)
	}

	metricglobal.SetMeterProvider(provider)
	return func() error {
		return provider.Shutdown(context.Background())
	}, nil
}

func newMetricsExporter(endpoint string, insecure bool, headers map[string]string) (metric.PushExporter, error) {
	secureOption := otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if insecure {
		secureOption = otlpmetricgrpc.WithInsecure()
	}
	return otlp.New(
		context.Background(),
		otlpmetricgrpc.NewClient(
			secureOption,
			otlpmetricgrpc.WithEndpoint(endpoint),
			otlpmetricgrpc.WithHeaders(headers),
			otlpmetricgrpc.WithCompressor(gzip.Name),
		),
	)
}

func viewOptions(c PipelineConfig) ([]view.Option, error) {
	syncPref := aggregation.CumulativeTemporality
	asyncPref := aggregation.CumulativeTemporality

	switch lower := strings.ToLower(c.TemporalityPreference); lower {
	case "delta":
		// Delta means exercising the cumulative-to-delta
		// export path.  This is an unusual setting for
		// Lightstep users to choose, but could be
		syncPref = aggregation.DeltaTemporality
		asyncPref = aggregation.DeltaTemporality
	case "stateless":
		syncPref = aggregation.DeltaTemporality
	case "cumulative":
		// Defaults set above.
	default:
		return nil, fmt.Errorf("invalid temporality preference: %v", c.TemporalityPreference)
	}
	return []view.Option{
		view.WithDefaultAggregationTemporalitySelector(
			func(k sdkinstrument.Kind) aggregation.Temporality {
				switch k {
				case sdkinstrument.UpDownCounterKind, sdkinstrument.UpDownCounterObserverKind:
					return aggregation.CumulativeTemporality
				case sdkinstrument.CounterKind, sdkinstrument.HistogramKind:
					return syncPref
				default:
					return asyncPref
				}
			},
		),
	}, nil
}
