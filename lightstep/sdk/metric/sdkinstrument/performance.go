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

// DefaultInactiveCollectionPeriods is how many collection periods to
// delay before removing records from memory.
const DefaultInactiveCollectionPeriods = 10

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
	return p
}
