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

package otlpmetric // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp/otlpmetric"

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	mpb "go.opentelemetry.io/proto/otlp/metrics/v1"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp/otlpmetric/internal/transform"
)

var errShutdown = fmt.Errorf("Exporter is shutdown")

// Exporter exports metrics data as OTLP.
type Exporter struct {
	// Ensure synchronous access to the client across all functionality.
	clientMu sync.Mutex
	client   Client

	shutdownOnce sync.Once
}

func (e *Exporter) String() string {
	return "otlp-lightstep"
}

// Temporality returns the Temporality to use for an instrument kind.
func (e *Exporter) Temporality(k metric.InstrumentKind) metricdata.Temporality {
	e.clientMu.Lock()
	defer e.clientMu.Unlock()
	return e.client.Temporality(k)
}

// Aggregation returns the Aggregation to use for an instrument kind.
func (e *Exporter) Aggregation(k metric.InstrumentKind) aggregation.Aggregation {
	e.clientMu.Lock()
	defer e.clientMu.Unlock()
	return e.client.Aggregation(k)
}

// ExportMetrics transforms and transmits metric data to an OTLP receiver.
func (e *Exporter) ExportMetrics(ctx context.Context, rm data.Metrics) error {
	otlpRm, err := transform.Metrics(rm)
	// Best effort upload of transformable metrics.
	e.clientMu.Lock()
	upErr := e.client.UploadMetrics(ctx, otlpRm)
	e.clientMu.Unlock()
	if upErr != nil {
		if err == nil {
			return upErr
		}
		// Merge the two errors.
		return fmt.Errorf("failed to upload incomplete metrics (%s): %w", err, upErr)
	}
	return err
}

// ForceFlushMetrics flushes any metric data held by an Exporter.
func (e *Exporter) ForceFlushMetrics(ctx context.Context, rm data.Metrics) error {
	err := e.ExportMetrics(ctx, rm)
	if err != nil {
		return err
	}
	// The Exporter does not hold data, forward the command to the client.
	e.clientMu.Lock()
	defer e.clientMu.Unlock()
	return e.client.ForceFlush(ctx)
}

// ShutdownMetrics flushes all metric data held by an Exporter and releases any held
// computational resources.
func (e *Exporter) ShutdownMetrics(ctx context.Context, rm data.Metrics) error {
	err := e.ExportMetrics(ctx, rm)
	e.shutdownOnce.Do(func() {
		e.clientMu.Lock()
		client := e.client
		e.client = shutdownClient{
			temporalitySelector: client.Temporality,
			aggregationSelector: client.Aggregation,
		}
		e.clientMu.Unlock()
		if stopErr := client.Shutdown(ctx); stopErr != nil && err == errShutdown {
			err = stopErr
		} else if stopErr != nil {
			err = fmt.Errorf("shutdown flush: %w, stop: %v", err, stopErr)
		}
	})
	return err
}

// New return an Exporter that uses client to transmits the OTLP data it
// produces. The client is assumed to be fully started and able to communicate
// with its OTLP receiving endpoint.
func New(client Client) *Exporter {
	return &Exporter{client: client}
}

type shutdownClient struct {
	temporalitySelector metric.TemporalitySelector
	aggregationSelector metric.AggregationSelector
}

func (c shutdownClient) err(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return errShutdown
}

func (c shutdownClient) Temporality(k metric.InstrumentKind) metricdata.Temporality {
	return c.temporalitySelector(k)
}

func (c shutdownClient) Aggregation(k metric.InstrumentKind) aggregation.Aggregation {
	return c.aggregationSelector(k)
}

func (c shutdownClient) UploadMetrics(ctx context.Context, _ *mpb.ResourceMetrics) error {
	return c.err(ctx)
}

func (c shutdownClient) ForceFlush(ctx context.Context) error {
	return c.err(ctx)
}

func (c shutdownClient) Shutdown(ctx context.Context) error {
	return c.err(ctx)
}
