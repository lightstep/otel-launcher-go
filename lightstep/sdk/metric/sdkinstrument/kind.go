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

//go:generate stringer -type=Kind

package sdkinstrument

// Kind describes the kind of instrument.
type Kind int8

const (
	// SyncCounter indicates a Counter instrument.
	SyncCounter Kind = iota
	// SyncUpDownCounter indicates a UpDownCounter instrument.
	SyncUpDownCounter
	// SyncGauge indicates a Gauge instrument.
	SyncGauge
	// SyncHistogram indicates a Histogram instrument.
	SyncHistogram

	// AsyncCounter indicates an asynchronous Counter instrument.
	AsyncCounter
	// AsyncUpDownCounter indicates a UpDownCounterObserver
	// instrument.
	AsyncUpDownCounter
	// AsyncGauge indicates an GaugeObserver instrument.
	AsyncGauge

	// NumKinds is the size of an array, useful for indexing by instrument kind.
	NumKinds
)

// Synchronous returns whether this is a synchronous kind of instrument.
func (k Kind) Synchronous() bool {
	switch k {
	case SyncCounter, SyncUpDownCounter, SyncGauge, SyncHistogram:
		return true
	}
	return false
}

// HasTemporality
func (k Kind) HasTemporality() bool {
	return k != AsyncGauge
}
