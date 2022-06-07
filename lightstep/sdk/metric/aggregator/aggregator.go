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

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
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
func RangeTest[N number.Any, Traits number.Traits[N]](num N, kind sdkinstrument.Kind) bool {
	var traits Traits

	if traits.IsInf(num) {
		otel.Handle(ErrInfInput)
		return false
	}

	if traits.IsNaN(num) {
		otel.Handle(ErrNaNInput)
		return false
	}

	// Check for negative values
	switch kind {
	case sdkinstrument.CounterKind,
		sdkinstrument.HistogramKind:
		if num < 0 {
			otel.Handle(ErrNegativeInput)
			return false
		}
	}
	return true
}

type HistogramConfig struct {
	MaxSize int32
}

type Config struct {
	Histogram HistogramConfig
}

// Methods implements a specific aggregation behavior.  Methods
// are parameterized by the type of the number (int64, float64),
// the Storage (generally an `Storage` struct in the same package).
type Methods[N number.Any, Storage any] interface {
	// Init initializes the storage.
	Init(ptr *Storage, cfg Config)

	// Update modifies the aggregator concurrently with respect to
	// Move() or Copy()
	Update(ptr *Storage, number N)

	// Move atomically copies `input` to `output` and resets the
	// `input` to the zero state.  The change to `input` is
	// synchronized with `Update()`.  The change to `output` is
	// synchronized with the accessor methods in ./aggregation.
	Move(input, output *Storage)

	// Merge adds the contents of `input` to `output`.  The read
	// of `input` is unsynchronized.  The write to `output` is
	// synchronized with concurrent `Merge()` calls (writing) and
	// concurrent `Copy()` calls (reading).
	Merge(input, output *Storage)

	// Copy replaces the contents of `output` with `input`.  The
	// read from `input` is synchronized with `Merge()` calls.
	Copy(input, output *Storage)

	// SubtractSwap performs `*operand = *value - *operand`
	// without synchronization.  We are not concerned with
	// synchronization because this is only used for asynchronous
	// instruments.
	SubtractSwap(value, operand *Storage)

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
	// be called after subtraction.
	HasChange(ptr *Storage) bool
}

// ConfigSelector is a per-instrument-kind, per-number-kind Config choice.
type ConfigSelector func(sdkinstrument.Kind) (int64Config, float64Config Config)
