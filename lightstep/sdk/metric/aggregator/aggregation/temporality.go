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

package aggregation // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"

import (
	"strings"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
)

type Temporality uint8

const (
	// UndefinedTemporality indicates that temporality is not defined.
	UndefinedTemporality Temporality = 0

	// CumulativeTemporality indicates that an Exporter expects a
	// Cumulative Aggregation.
	CumulativeTemporality Temporality = 1

	// DeltaTemporality indicates that an Exporter expects a
	// Delta Aggregation.
	DeltaTemporality Temporality = 2
)

type TemporalitySelector func(sdkinstrument.Kind) Temporality

type TemporalityTrait interface {
	Temporality() Temporality
}

type DeltaTemporalityTrait struct{}
type CumulativeTemporalityTrait struct{}

func (DeltaTemporalityTrait) Temporality() Temporality {
	return DeltaTemporality
}

func (CumulativeTemporalityTrait) Temporality() Temporality {
	return CumulativeTemporality
}

func (t Temporality) Valid() bool {
	switch t {
	case UndefinedTemporality, DeltaTemporality, CumulativeTemporality:
		return true
	}
	return false
}

const (
	NameCumulative = "cumulative"
	NameDelta      = "delta"
)

func (t Temporality) String() string {
	switch t {
	case DeltaTemporality:
		return NameDelta
	case CumulativeTemporality:
		return NameCumulative
	}
	return "undefined"
}

func ParseTemporality(str string) (Temporality, bool) {
	switch strings.ToLower(str) {
	case NameCumulative:
		return CumulativeTemporality, true
	case NameDelta:
		return DeltaTemporality, true
	}
	return UndefinedTemporality, false
}

func LowMemoryTemporality(k sdkinstrument.Kind) Temporality {
	switch k {
	case sdkinstrument.SyncUpDownCounter, sdkinstrument.AsyncUpDownCounter:
		return CumulativeTemporality
	case sdkinstrument.SyncCounter, sdkinstrument.SyncHistogram:
		return DeltaTemporality
	default:
		return CumulativeTemporality
	}
}
