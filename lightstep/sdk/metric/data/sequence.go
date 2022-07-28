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

package data // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"

import (
	"time"
)

// Sequence provides the three relevant timestamps that are used by
// the SDK during collection.  Depending on aggregation temporality,
// either `Start` or `Last` will be used.
type Sequence struct {
	// Start is the time when the MeterProvider was initialized.
	Start time.Time
	// Last is the time when the previous collection
	// happened.  If there was no previous collection,
	// this will match Start.
	Last time.Time
	// Now is the moment the current collection began.  This value
	// will be used as the subsequent value for Last.
	Now time.Time
}

// Collector is an interface for producing a single Instrument of data.
type Collector interface {
	// Collect gathers data points from processed accumulator
	// snapshots into output.
	Collect(sequence Sequence, output *[]Instrument)

	// Size returns the number of entries held in memory.  Size()
	// is meant to be called following Collect().
	Size() int
}
