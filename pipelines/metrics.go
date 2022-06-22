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
	"time"

	hostMetrics "go.opentelemetry.io/contrib/instrumentation/host"
	runtimeMetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	metricglobal "go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"google.golang.org/grpc/encoding/gzip"
)

func NewMetricsPipeline(c PipelineConfig) (func() error, error) {
	metricExporter, err := c.newMetricsExporter()
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %v", err)
	}

	period := controller.DefaultPeriod
	if c.ReportingPeriod != "" {
		period, err = time.ParseDuration(c.ReportingPeriod)
		if err != nil {
			return nil, fmt.Errorf("invalid metric reporting period: %v", err)
		}
		if period <= 0 {

			return nil, fmt.Errorf("invalid metric reporting period: %v", c.ReportingPeriod)
		}
	}
	pusher := controller.New(
		processor.NewFactory(
			selector.NewWithInexpensiveDistribution(),
			metricExporter,
		),
		controller.WithExporter(metricExporter),
		controller.WithResource(c.Resource),
		controller.WithCollectPeriod(period),
	)

	if err = pusher.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to start controller: %v", err)
	}

	if err = runtimeMetrics.Start(runtimeMetrics.WithMeterProvider(pusher)); err != nil {
		return nil, fmt.Errorf("failed to start runtime metrics: %v", err)
	}

	if err = hostMetrics.Start(hostMetrics.WithMeterProvider(pusher)); err != nil {
		return nil, fmt.Errorf("failed to start host metrics: %v", err)
	}

	metricglobal.SetMeterProvider(pusher)
	return func() error {
		_ = pusher.Stop(context.Background())
		return metricExporter.Shutdown(context.Background())
	}, nil
}

func (c PipelineConfig) newMetricsExporter() (*otlpmetric.Exporter, error) {
	return otlpmetric.New(
		context.Background(),
		otlpmetricgrpc.NewClient(
			c.secureMetricOption(),
			otlpmetricgrpc.WithEndpoint(c.Endpoint),
			otlpmetricgrpc.WithHeaders(c.Headers),
			otlpmetricgrpc.WithCompressor(gzip.Name),
		),
	)
}
