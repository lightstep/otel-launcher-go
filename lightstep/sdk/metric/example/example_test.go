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

package example

import (
	"context"
	"fmt"
	"time"

	// Note the SDK, exporter, and test appratus are from this repository
	lightstep "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"
	exporter "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp"
	otlptest "github.com/lightstep/otel-launcher-go/pipelines/test"

	// Note the gRPC/OTLP exporter and protocol are from the community SDK.
	otlpmetricgrpc "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	otlpproto "go.opentelemetry.io/proto/otlp/metrics/v1"
)

// ExampleMinimumConfig tests the minimum configuration for a single point.
func ExampleMinimumConfig() {
	ctx := context.Background()
	server := otlptest.NewServer()

	// Configure an exporter.
	exp, _ := exporter.New(ctx, otlpmetricgrpc.NewClient(
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(fmt.Sprint(otlptest.ServerName, ":", server.InsecureMetricsPort)),
		otlpmetricgrpc.WithHeaders(map[string]string{"lightstep-access-token": "${TOKEN}"}),
	))

	// Configure the SDK.
	sdk := lightstep.NewMeterProvider(
		lightstep.WithReader(lightstep.NewPeriodicReader(exp, 30*time.Second)),
	)

	// Configure a Meter and instrument.
	meter := sdk.Meter("meter")
	counter, _ := meter.SyncInt64().Counter("how-many")

	// Count once and shutdown.
	counter.Add(ctx, 1)
	sdk.Shutdown(ctx)

	oneScope := server.MetricsRequests()[0].ResourceMetrics[0].ScopeMetrics[0]
	oneHeader := server.MetricsMDs()[0]

	fmt.Println(
		oneScope.Scope.Name,
		oneScope.Metrics[0].Name,
		oneScope.Metrics[0].Data.(*otlpproto.Metric_Sum).Sum.DataPoints[0].Value.(*otlpproto.NumberDataPoint_AsInt).AsInt,
		oneHeader["lightstep-access-token"][0],
	)

	// Output: meter how-many 1 ${TOKEN}
}
