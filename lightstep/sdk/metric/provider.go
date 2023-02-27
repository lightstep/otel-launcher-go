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
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/pipeline"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/viewstate"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/multierr"
)

// MeterProvider handles the creation and coordination of Meters. All Meters
// created by a MeterProvider will be associated with the same Resource, have
// the same Views applied to them, and have their produced metric telemetry
// passed to the configured Readers.
type MeterProvider struct {
	cfg       config
	startTime time.Time
	lock      sync.Mutex
	ordered   []*meter
	meters    map[instrumentation.Library]*meter
}

// Compile-time check MeterProvider implements metric.MeterProvider.
var _ metric.MeterProvider = (*MeterProvider)(nil)

var ErrAlreadyShutdown = fmt.Errorf("provider was already shut down")

// NewMeterProvider returns a new and configured MeterProvider.
//
// By default, the returned MeterProvider is configured with the default
// Resource and no Readers. Readers cannot be added after a MeterProvider is
// created. This means the returned MeterProvider, one created with no
// Readers, will be perform no operations.
func NewMeterProvider(options ...Option) *MeterProvider {
	cfg := config{
		res: resource.Default(),
	}
	for _, option := range options {
		cfg = option.apply(cfg)
	}
	cfg.performance = cfg.performance.Validate()
	p := &MeterProvider{
		cfg:       cfg,
		startTime: time.Now(),
		meters:    map[instrumentation.Library]*meter{},
	}
	for pipe := 0; pipe < len(cfg.readers); pipe++ {
		cfg.readers[pipe].Register(p.producerFor(pipe))
	}
	return p
}

// Meter returns a Meter with the given name and configured with options.
//
// The name should be the name of the instrumentation scope creating
// telemetry. This name may be the same as the instrumented code only if that
// code provides built-in instrumentation.
//
// If name is empty, the default (go.opentelemetry.io/otel/sdk/meter) will be
// used.
//
// Calls to the Meter method after Shutdown has been called will return Meters
// that perform no operations.
//
// This method is safe to call concurrently.
func (mp *MeterProvider) Meter(name string, options ...metric.MeterOption) metric.Meter {
	cfg := metric.NewMeterConfig(options...)
	lib := instrumentation.Library{
		Name:      name,
		Version:   cfg.InstrumentationVersion(),
		SchemaURL: cfg.SchemaURL(),
	}

	mp.lock.Lock()
	defer mp.lock.Unlock()

	m := mp.meters[lib]
	if m != nil {
		return m
	}
	m = &meter{
		provider:  mp,
		library:   lib,
		byDesc:    map[sdkinstrument.Descriptor]interface{}{},
		compilers: pipeline.NewRegister[*viewstate.Compiler](len(mp.cfg.readers)),
	}
	for pipe := range m.compilers {
		m.compilers[pipe] = viewstate.New(lib, mp.cfg.views[pipe])
	}
	mp.ordered = append(mp.ordered, m)
	mp.meters[lib] = m
	return m
}

// ForceFlush flushes all pending telemetry.
//
// This method honors the deadline or cancellation of ctx. An appropriate
// error will be returned in these situations. There is no guaranteed that all
// telemetry be flushed or all resources have been released in these
// situations.
//
// This method is safe to call concurrently.
func (mp *MeterProvider) ForceFlush(ctx context.Context) error {
	var err error
	for _, r := range mp.cfg.readers {
		err = multierr.Append(err, r.ForceFlush(ctx))
	}
	return err
}

// Shutdown shuts down the MeterProvider flushing all pending telemetry and
// releasing any held computational resources.
//
// This call is idempotent. The first call will perform all flush and
// releasing operations. Subsequent calls will perform no action.
//
// Measurements made by instruments from meters this MeterProvider created
// will not be exported after Shutdown is called.
//
// This method honors the deadline or cancellation of ctx. An appropriate
// error will be returned in these situations. There is no guaranteed that all
// telemetry be flushed or all resources have been released in these
// situations.
//
// This method is safe to call concurrently.
func (mp *MeterProvider) Shutdown(ctx context.Context) error {
	var err error

	// Note that the readers will call getOrdered(), so take the
	// lock to check for previous shutdown and release.
	mp.lock.Lock()
	meters := mp.meters
	mp.meters = nil
	mp.lock.Unlock()

	if meters == nil {
		return ErrAlreadyShutdown
	}

	for _, r := range mp.cfg.readers {
		err = multierr.Append(err, r.Shutdown(ctx))
	}

	return err
}

// getOrdered returns meters in the order they were registered.
func (mp *MeterProvider) getOrdered() []*meter {
	mp.lock.Lock()
	defer mp.lock.Unlock()
	return mp.ordered
}
