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

//go:generate stringer -type=Category,Kind

package aggregation // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"

import (
	"strings"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
)

// These interfaces describe the various ways to access state from an
// Aggregation.

type (
	// Aggregation is an interface returned by the Aggregator
	// containing an interval of metric data.
	Aggregation interface {
		Kind() Kind
	}

	// HasASum includes Sum and Histogram aggregators.
	HasASum interface {
		Sum() number.Number
	}

	// Sum returns an aggregated sum.
	Sum interface {
		// Review NOTE: Should this be Total() or Value()?
		HasASum
		IsMonotonic() bool
	}

	// Gauge returns the latest value that was aggregated.
	Gauge interface {
		Aggregation

		// Gauge returns the most recently observed value.
		Gauge() number.Number
	}

	// Histogram returns the count of events in exponential-scale
	// buckets defined as a function of a scale parameter.  See a
	// detailed explanation in the OpenTelemetry metrics data
	// model:
	// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/metrics/datamodel.md#exponentialhistogram
	Histogram interface {
		Aggregation
		Count() uint64
		HasASum
		Scale() int32
		ZeroCount() uint64
		Positive() Buckets
		Negative() Buckets
	}

	// Buckets describes a range of consecutive buckets, starting
	// at Offset().  This type is used to encode either the
	// positive or negative ranges of an Histogram.
	Buckets interface {
		Offset() int32
		Len() uint32
		At(uint32) uint64
	}

	// MinMaxSumCount is a low cost HistogramCategory aggregator
	// that records the Min, Max, Sum, and Count.
	MinMaxSumCount interface {
		Aggregation
		Count() uint64
		HasASum
		Min() number.Number
		Max() number.Number
	}
)

// Category constants describe semantic kind.  For the histogram
// category there are multiple implementations, for those distinctions
// as well as Drop, use Kind.
type Category int

const (
	UndefinedCategory Category = iota
	MonotonicSumCategory
	NonMonotonicSumCategory
	GaugeCategory
	HistogramCategory
)

type Kind int

const (
	UndefinedKind Kind = iota
	DropKind
	AnySumKind
	MonotonicSumKind
	NonMonotonicSumKind
	GaugeKind
	HistogramKind
	MinMaxSumCountKind
)

func (k Kind) Category(ik sdkinstrument.Kind) Category {
	switch k {
	case AnySumKind:
		switch ik {
		case sdkinstrument.SyncHistogram, sdkinstrument.SyncCounter, sdkinstrument.AsyncCounter:
			return MonotonicSumCategory
		case sdkinstrument.SyncUpDownCounter, sdkinstrument.AsyncUpDownCounter:
			return NonMonotonicSumCategory
		}
		return UndefinedCategory
	case MonotonicSumKind:
		return MonotonicSumCategory
	case NonMonotonicSumKind:
		return NonMonotonicSumCategory
	case GaugeKind:
		return GaugeCategory
	case HistogramKind, MinMaxSumCountKind:
		return HistogramCategory
	default:
		return UndefinedCategory
	}
}

// KindSelector is a per-instrument-kind Kind choice.
type KindSelector func(sdkinstrument.Kind) Kind

// Valid returns true when the value is one of the enumerated
// constants.
func (k Kind) Valid() bool {
	switch k {
	case UndefinedKind, DropKind, AnySumKind,
		MonotonicSumKind, NonMonotonicSumKind,
		GaugeKind, HistogramKind, MinMaxSumCountKind:
		return true
	}
	return false
}

// ParseKind returns the enumerated constant and true if the string
// corresponds with a known aggregation kind.
func ParseKind(str string) (Kind, bool) {
	switch strings.ToLower(str) {
	case "drop":
		return DropKind, true
	case "sum":
		return AnySumKind, true
	case "monotonic_sum":
		return MonotonicSumKind, true
	case "nonmonotonic_sum":
		return NonMonotonicSumKind, true
	case "gauge":
		return GaugeKind, true
	case "histogram", "exponential_histogram":
		return HistogramKind, true
	case "minmaxsumcount":
		return MinMaxSumCountKind, true
	}
	return UndefinedKind, false
}
