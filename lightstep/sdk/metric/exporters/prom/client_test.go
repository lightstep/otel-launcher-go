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

package prom

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	sdkmetric "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"
	"github.com/stretchr/testify/suite"
)

const promPort = 2356

var testResourceAttrs = []attribute.KeyValue{
	attribute.String("property", "value"),
	attribute.String("service.name", "tester"),
}

type clientTestSuite struct {
	suite.Suite
	sdk *sdkmetric.MeterProvider
}

func TestExporterSuite(t *testing.T) {
	suite.Run(t, new(clientTestSuite))
}

func (t *clientTestSuite) SetupSuite() {
	ctx := context.Background()

	exp, err := NewExporter(
		ctx,
		NewConfig(WithPort(promPort)),
	)
	require.NoError(t.T(), err)

	t.sdk = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp, time.Second), view.WithDefaultAggregationTemporalitySelector(aggregation.LowMemoryTemporality)),
		sdkmetric.WithResource(
			resource.NewSchemaless(testResourceAttrs...),
		),
	)
}

func (t *clientTestSuite) TearDownSuite() {
	require.NoError(t.T(), t.sdk.Shutdown(context.Background()))
}

func (t *clientTestSuite) TestInt64Counter() {
	ctx := context.Background()

	meter := t.sdk.Meter("test-meter")
	counter, err := meter.Int64Counter("requests")
	require.NoError(t.T(), err)

	counter.Add(ctx, 12)

	require.Eventuallyf(t.T(), func() bool {
		lines := readMetricsEndpoint(t.T())

		return slices.Contains(lines, `requests{job="tester",property="value",service_name="tester"} 12`)
	}, 15*time.Second, time.Second, "verify requests metric")
}

func (t *clientTestSuite) TestInt64Histogram() {
	ctx := context.Background()

	meter := t.sdk.Meter("test-meter")
	counter, err := meter.Int64Histogram("request-size")
	require.NoError(t.T(), err)

	counter.Record(ctx, 0)

	require.Eventuallyf(t.T(), func() bool {
		lines := readMetricsEndpoint(t.T())

		return slices.Contains(lines, `request_size_bucket{job="tester",property="value",service_name="tester",le="0"} 1`)
	}, 15*time.Second, time.Second, "verify request-size metric")
}

func readMetricsEndpoint(t *testing.T) []string {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", promPort))
	require.NoError(t, err)
	defer resp.Body.Close() // Ensure that the response body is closed after reading

	// Check if the HTTP status is 200 (OK)
	require.Equal(t, resp.StatusCode, http.StatusOK)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return strings.Split(string(body), "\n")
}
