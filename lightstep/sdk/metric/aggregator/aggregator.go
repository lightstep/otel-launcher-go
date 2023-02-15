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

package aggregator // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"

import (
	"fmt"
	"time"

	histostruct "github.com/lightstep/go-expohisto/structure"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/doevery"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"go.opentelemetry.io/otel"
)

// Sentinel errors for Aggregator interface.
var (
	ErrNegativeInput = fmt.Errorf("negative value is out of range for this instrument")
	ErrNaNInput      = fmt.Errorf("NaN value is an invalid input")
	ErrInfInput      = fmt.Errorf("±Inf value is an invalid input")
)

// RangeTest is a common routine for testing for valid input values.
// This rejects NaN and Inf values.  This rejects negative values when the
// aggregation does not support negative values, including
// monotonic counter metrics and Histogram metrics.
func RangeTest[N number.Any, Traits number.Traits[N]](num N, desc sdkinstrument.Descriptor) bool {
	var traits Traits

	if traits.IsInf(num) {
		doevery.TimePeriod(30*time.Second, func() {
			otel.Handle(fmt.Errorf("%s: %w", desc.Name, ErrInfInput))
		})
		return false
	}

	if traits.IsNaN(num) {
		doevery.TimePeriod(30*time.Second, func() {
			otel.Handle(fmt.Errorf("%s: %w", desc.Name, ErrNaNInput))
		})
		return false
	}

	// Check for negative values
	switch desc.Kind {
	case sdkinstrument.SyncCounter,
		sdkinstrument.SyncHistogram:
		if num < 0 {
			doevery.TimePeriod(30*time.Second, func() {
				otel.Handle(fmt.Errorf("%s: %w", desc.Name, ErrNegativeInput))
			})
			return false
		}
	}
	return true
}

// JSONHistogramConfig configures the exponential histogram.
type JSONHistogramConfig struct {
	MaxSize int32 `json:"max_size"`
}

// JSONConfig supports the configuration for all aggregators in a single struct.
type JSONConfig struct {
	Histogram JSONHistogramConfig `json:"histogram"`
}

// ToConfig returns a Config from the fixed-JSON represented.
func (jc JSONConfig) ToConfig() Config {
	return Config{
		Histogram: histostruct.NewConfig(histostruct.WithMaxSize(jc.Histogram.MaxSize)),
	}
}

// Config supports the configuration for all aggregators in a single struct.
type Config struct {
	Histogram histostruct.Config
}

// Valid returns true for valid configurations.
func (c Config) Valid() bool {
	_, err := c.Validate()
	return err == nil
}

// Valid returns a valid Configuration along with an error if there
// were invalid settings.  Note that the empty state is considered valid and a correct
func (c Config) Validate() (Config, error) {
	var err error
	c.Histogram, err = c.Histogram.Validate()
	return c, err
}

// Methods implements a specific aggregation behavior for a specific
// type of aggregator Storage.  Methods are parameterized by the type
// of the number (int64, float64), the Storage (generally a `Storage`
// struct in the same package as the corresponding Methods).
//
// Methods have four methods that mutate the Storage. In every case,
// one of the Storage is synchronized against the other operations.
// The synchronized methods are:
//
// Update: Modifies one Storage (synchronized).
// Move: Reads-and-resets one Storage (synchronized), writes one Storage.
// Copy: Reads one Storage (synchronized), writes one Storage.
// Merge: Reads one Storage, writes one Storeage (synchronized).
//
// Generally, the sequence of operations from observation to export is
// different for synchronous and asynchronous instruments.  For
// synchronous instruments:
//
// 1. Update() from an API method call into the accumulator's current Storage
// 2. Move() from the accumulator's storage into the accumulator's snapshot Storage
// 3. Merge() from the snapshot Storage to the output Storage.
// 4. Copy() or Move() from the output Storage to the exported data.
//
// Note that these methods are responsible for synchronization between
// steps (1 vs 2) and (3 vs 4).  The accumulator uses its own lock to
// protect the snapshot Storage between steps 2 and 3.
type Methods[N number.Any, Storage any] interface {
	// Init initializes the storage.
	Init(ptr *Storage, cfg Config)

	// Update modifies Storage concurrently with respect to
	// concurrent Move(), Copy(), and Update() operations.
	Update(ptr *Storage, number N)

	// Move atomically copies `input` to `output` and resets
	// `input` to the zero state.  The change to `input` is
	// synchronized against concurrent `Update()` and `Merge()`
	// operations.  The change to `output` is not synchronized.
	Move(input, output *Storage)

	// Merge adds the contents of `input` to `output`.  The read
	// of `input` is not synchronized.  The write to `output` is
	// synchronized with concurrent `Merge()` calls (writing) and
	// concurrent `Copy()` or `Move()` calls (reading).
	Merge(input, output *Storage)

	// Copy replaces the contents of `output` with the contents of
	// `input`, which is unmodified.  The read from `input` is
	// synchronized with concurrent `Merge()` and `Update()` calls.
	Copy(input, output *Storage)

	// SubtractSwap performs `*operand = *argument - *operand`
	// with no synchronization.  We are not concerned with
	// synchronization because this is only used for asynchronous
	// instruments.  To use SubtractSwap in a synchronous
	// scenario, use Copy() or Move() first.
	SubtractSwap(operand, argument *Storage)

	// ToAggregation returns an exporter-ready value.
	ToAggregation(ptr *Storage) aggregation.Aggregation

	// ToStorage returns the underlying storage of an existing Aggregation.
	ToStorage(aggregation.Aggregation) (*Storage, bool)

	// Kind returns the Kind of aggregator.
	Kind() aggregation.Kind

	// HasChange returns true if there have been any (discernible)
	// Updates.  This tests whether an aggregation has zero sum,
	// zero count, or zero difference, depending on the
	// aggregation.  If the instrument is asynchronous, this will
	// be called after subtraction.  Not synchronized.
	HasChange(ptr *Storage) bool
}

// ConfigSelector is a per-instrument-kind, per-number-kind Config choice.
type ConfigSelector func(sdkinstrument.Kind) (int64Config, float64Config Config)
