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

package pipeline // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/pipeline"

// Register is a per-pipeline slice of data.  Although it is nothing
// more than a simple slice, the use of this generic wrapper serves as
// a notice that the associated data will be indexed by an integer,
// generally named "pipe", referring to the ordinal position of the
// reader in the list of readers.
type Register[T any] []T

// NewRegister returns a new slice of per-pipeline data.
func NewRegister[T any](size int) Register[T] {
	return Register[T](make([]T, size))
}
