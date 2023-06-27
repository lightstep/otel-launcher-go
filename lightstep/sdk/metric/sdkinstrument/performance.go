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

package sdkinstrument

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

// DefaultInactiveCollectionPeriods is how many collection periods to
// delay before removing records from memory.
const DefaultInactiveCollectionPeriods = 10

// DefaultAggregatorCardinalityLimit is a hard limit on the number of
// aggregators that can be emitted in a single period.
const DefaultAggregatorCardinalityLimit = 2000

// DefaultInstrumentCardinalityLimit is a hard limit on the number of
// aggregators that can be accumulated in intermediate state belonging
// to the instrument.
const DefaultInstrumentCardinalityLimit = 3000

// Performace configures features that allow the user to control
// performance.
type Performance struct {
	// IgnoreCollisions indicates the user is willing to bypass an
	// attributes-set comparison after finding a fingerprint
	// match.
	IgnoreCollisions bool

	// InactiveCollectionPeriods is the number of allowed
	// collection periods having no updates before the record is
	// removed from memory.
	InactiveCollectionPeriods uint32

	// InstrumentCardinalityLimit is the point at which the
	// SDK's emergency overflow breaker begins dropping attributes
	// to avoid memory buildup at intermediate pipeline stages.
	InstrumentCardinalityLimit uint32

	// AggregatorCardinalityLimit is a hard limit on output
	// cardinality for all aggregators in the SDK.
	AggregatorCardinalityLimit uint32

	// MeasurementProcessor supports modifying the attributes
	// based on context.  Only applies to synchronous instruments.
	MeasurementProcessor MeasurementProcessor
}

// MeasurementProcessor allows applications to extend metric events
// based on context.
type MeasurementProcessor interface {
	Process(ctx context.Context, inAttrs []attribute.KeyValue) (outAttrs []attribute.KeyValue)
}

// Validate returns a Performance object with 0 values replaced by
// defaults and errors checked.
func (p Performance) Validate() Performance {
	// If InactiveCollectionPeriods 0 is a valid setting, but can
	// lead to poor performance, so we let it use the default. The
	// user can configure 1 for the same effect as 0.
	if p.InactiveCollectionPeriods == 0 {
		p.InactiveCollectionPeriods = DefaultInactiveCollectionPeriods
	}
	if p.InstrumentCardinalityLimit == 0 {
		p.InstrumentCardinalityLimit = DefaultInstrumentCardinalityLimit
	}
	if p.AggregatorCardinalityLimit == 0 {
		p.AggregatorCardinalityLimit = DefaultAggregatorCardinalityLimit
	}

	return p
}
