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
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
)

// Descriptor contains all the settings that describe an instrument,
// including its name, metric kind, number kind, and the configurable
// options.
type Descriptor struct {
	// Name returns the metric instrument's name.
	Name string

	// Kind returns the specific kind of instrument.
	Kind Kind

	// NumberKind returns whether this instrument captures int64
	// or float64 values.
	NumberKind number.Kind

	// Description provides a human-readable description of the metric.
	Description string

	// Unit describes the units of the metric instrument.  Unitless
	// metrics return the empty string.
	Unit string
}

// NewDescriptor returns a Descriptor with the given contents.
func NewDescriptor(name string, ikind Kind, nkind number.Kind, description, unit string) Descriptor {
	return Descriptor{
		Name:        name,
		Kind:        ikind,
		NumberKind:  nkind,
		Description: description,
		Unit:        unit,
	}
}
