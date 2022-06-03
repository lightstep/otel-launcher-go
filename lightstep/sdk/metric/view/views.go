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

// Views is a configured set of view clauses with an associated Name
// that is used for debugging.
type Views struct {
	// Name of these views, used in error reporting.
	Name string

	// Config is the configuration for these views.
	Config
}

// New configures the clauses and default settings of a Views.
func New(name string, opts ...Option) *Views {
	return &Views{
		Name:   name,
		Config: NewConfig(opts...),
	}
}

// TODO: call views.Validate() to check for:
// - empty (?)
// - duplicate name
// - invalid inst/number/aggregation kind
// - both instrument name and regexp
// - schemaURL or Version without library name
// - empty attribute keys
// - Name w/o SingleInst
