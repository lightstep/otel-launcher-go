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

package otlp // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp"

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp/internal/metrictransform"
)

var (
	errAlreadyStarted = errors.New("already started")
)

// Exporter exports metrics data in the OTLP wire format.
type Exporter struct {
	client Client

	mu      sync.RWMutex
	started bool

	startOnce sync.Once
	stopOnce  sync.Once
}

// Export exports a batch of metrics.
func (e *Exporter) ExportMetrics(ctx context.Context, metrics data.Metrics) error {
	rm, err := metrictransform.Metrics(metrics)
	if err != nil {
		return err
	}
	if rm == nil {
		return nil
	}

	return e.client.UploadMetrics(ctx, rm)
}

// Start establishes a connection to the receiving endpoint.
func (e *Exporter) Start(ctx context.Context) error {
	var err = errAlreadyStarted
	e.startOnce.Do(func() {
		e.mu.Lock()
		e.started = true
		e.mu.Unlock()
		err = e.client.Start(ctx)
	})

	return err
}

func (e *Exporter) String() string {
	return "otlp-lightstep"
}

func (e *Exporter) ForceFlushMetrics(ctx context.Context, metrics data.Metrics) error {
	return e.ExportMetrics(ctx, metrics)
}

// Shutdown flushes all exports and closes all connections to the receiving endpoint.
func (e *Exporter) ShutdownMetrics(ctx context.Context, metrics data.Metrics) error {
	e.mu.RLock()
	started := e.started
	e.mu.RUnlock()

	if !started {
		return nil
	}

	var err error

	e.stopOnce.Do(func() {
		err = e.ExportMetrics(ctx, metrics)
		if stopErr := e.client.Stop(ctx); stopErr != nil && err == nil {
			err = stopErr
		} else if stopErr != nil {
			err = fmt.Errorf("shutdown flush: %w, stop: %v", err, stopErr)
		}
		e.mu.Lock()
		e.started = false
		e.mu.Unlock()
	})

	return err
}

var _ metric.PushExporter = (*Exporter)(nil)

// New constructs a new Exporter and starts it.
func New(ctx context.Context, client Client, opts ...Option) (*Exporter, error) {
	exp := NewUnstarted(client, opts...)
	if err := exp.Start(ctx); err != nil {
		return nil, err
	}
	return exp, nil
}

// NewUnstarted constructs a new Exporter and does not start it.
func NewUnstarted(client Client, opts ...Option) *Exporter {
	cfg := config{}

	for _, opt := range opts {
		cfg = opt.apply(cfg)
	}

	e := &Exporter{
		client: client,
	}

	return e
}
