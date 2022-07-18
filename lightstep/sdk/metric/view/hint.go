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

package view // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"

import (
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
)

// Hint is a structure that can be serialized into a Description field
// to control the aggregation and config. These hints are only taken
// when no other View clauses match the instrument.
type Hint struct {
	// Description takes the place of the hint after decoding
	// this Hint from the Description.
	Description string `json:"description"`

	// Aggregation determines the kind of aggregator used.  When
	// this is set, semantic compatibility checking is bypassed.
	Aggregation string `json:"aggregation"`

	// Config configures the aggregator selected in the
	// Aggregation field.
	Config aggregator.JSONConfig `json:"config"`
}
