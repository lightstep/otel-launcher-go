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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdkmetric "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	otlpmetric "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"

	hostMetrics "go.opentelemetry.io/contrib/instrumentation/host"
	runtimeMetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	metricglobal "go.opentelemetry.io/otel/metric/global"

	// The old Metrics SDK
	oldotlpmetric "go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	oldaggregation "go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
)

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

	newPref, oldPref, err := tempoOptions(c)
	if err != nil {
		return nil, fmt.Errorf("invalid metric view configuration: %v", err)
	}

	if c.UseLightstepMetricsSDK {
		// Install the Lightstep metrics SDK
		metricExporter, err := c.newMetricsExporter()
		if err != nil {
			return nil, fmt.Errorf("failed to create metric exporter: %v", err)
		}

		sdk := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(c.Resource),
			sdkmetric.WithReader(
				sdkmetric.NewPeriodicReader(metricExporter, period),
				newPref,
			),
		)

		provider = sdk
		shutdown = func() error {
			return sdk.Shutdown(context.Background())
		}

	} else {
		// Install the OTel-Go community metrics SDK.
		metricExporter, err := c.newOldMetricsExporter(oldPref)
		if err != nil {
			return nil, fmt.Errorf("failed to create metric exporter: %v", err)
		}
		sdk := controller.New(
			processor.NewFactory(
				selector.NewWithHistogramDistribution(),
				metricExporter,
			),
			controller.WithExporter(metricExporter),
			controller.WithResource(c.Resource),
			controller.WithCollectPeriod(period),
		)

		if err = sdk.Start(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to start controller: %v", err)
		}

		provider = sdk
		shutdown = func() error {
			return sdk.Stop(context.Background())
		}
	}

	if err = runtimeMetrics.Start(runtimeMetrics.WithMeterProvider(provider)); err != nil {
		return nil, fmt.Errorf("failed to start runtime metrics: %v", err)
	}

	if err = hostMetrics.Start(hostMetrics.WithMeterProvider(provider)); err != nil {
		return nil, fmt.Errorf("failed to start host metrics: %v", err)
	}

	metricglobal.SetMeterProvider(provider)
	return shutdown, nil
}

var errNoSingleCount = fmt.Errorf("no count")

func singleCount(values []string) (int, error) {
	if len(values) != 1 {
		return 0, errNoSingleCount
	}
	return strconv.Atoi(values[0])
}

type dropExample struct {
	Reason string   `json:"reason"`
	Names  []string `json:"names"`
}

type dropSummary struct {
	Dropped struct {
		Points  int `json:"points,omitempty"`
		Metrics int `json:"metrics,omitempty"`
	} `json:"dropped,omitempty"`
	Examples []dropExample `json:"examples,omitempty"`
}

func (ds *dropSummary) Empty() bool {
	return len(ds.Examples) == 0 && ds.Dropped.Points == 0 && ds.Dropped.Metrics == 0
}

func interceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	const invalidTrailerPrefix = "otlp-invalid-"

	var md metadata.MD
	err := invoker(ctx, method, req, reply, cc, append(opts, grpc.Trailer(&md))...)
	if err == nil && md != nil {
		var ds dropSummary
		for key, values := range md {
			key = strings.ToLower(key)
			if !strings.HasPrefix(key, "otlp-") {
				continue
			}

			if key == "otlp-points-dropped" {
				if points, err := singleCount(values); err == nil {
					ds.Dropped.Points = points
				}
			} else if key == "otlp-metrics-dropped" {
				if metrics, err := singleCount(values); err == nil {
					ds.Dropped.Metrics = metrics
				}
			} else if strings.HasPrefix(key, invalidTrailerPrefix) {
				key = key[len(invalidTrailerPrefix):]
				key = strings.ReplaceAll(key, "-", " ")
				ds.Examples = append(ds.Examples, dropExample{
					Reason: key,
					Names:  values,
				})
			}
		}
		if !ds.Empty() {
			data, _ := json.Marshal(ds)
			otel.Handle(fmt.Errorf("metrics partial failure: %v", string(data)))
		}
	}
	return err
}

func (c PipelineConfig) newClient() otlpmetric.Client {
	return otlpmetricgrpc.NewClient(
		c.secureMetricOption(),
		otlpmetricgrpc.WithEndpoint(c.Endpoint),
		otlpmetricgrpc.WithHeaders(c.Headers),
		otlpmetricgrpc.WithCompressor(gzip.Name),
		otlpmetricgrpc.WithDialOption(
			grpc.WithUnaryInterceptor(interceptor),
		),
	)
}

func (c PipelineConfig) newMetricsExporter() (*otlpmetric.Exporter, error) {
	return otlpmetric.New(
		context.Background(),
		c.newClient(),
	)
}

func (c PipelineConfig) newOldMetricsExporter(tempo oldaggregation.TemporalitySelector) (*oldotlpmetric.Exporter, error) {
	return oldotlpmetric.New(
		context.Background(),
		c.newClient(),
		oldotlpmetric.WithMetricAggregationTemporalitySelector(tempo),
	)
}

func tempoOptions(c PipelineConfig) (view.Option, oldaggregation.TemporalitySelector, error) {
	syncPref := aggregation.CumulativeTemporality
	asyncPref := aggregation.CumulativeTemporality
	var oldSelector oldaggregation.TemporalitySelector

	switch lower := strings.ToLower(c.TemporalityPreference); lower {
	case "delta":
		// Delta means exercising the cumulative-to-delta
		// export path.  This is an unusual setting for
		// Lightstep users to choose.
		syncPref = aggregation.DeltaTemporality
		asyncPref = aggregation.DeltaTemporality

		oldSelector = oldaggregation.DeltaTemporalitySelector()
	case "stateless":
		// asyncPref set above.
		syncPref = aggregation.DeltaTemporality

		oldSelector = oldaggregation.StatelessTemporalitySelector()
	case "", "cumulative":
		// syncPref, asyncPref set above.
		oldSelector = oldaggregation.CumulativeTemporalitySelector()
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
	), oldSelector, nil
}
