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

package fprint

import (
	"fmt"
	"unsafe"
)

var errStringTooLong = fmt.Errorf("string length exceeds max")

// unsafeStringToBytes returns a []byte slice
// that has a zero-copy view into the underlying
// array backing s.
//
// Note that strings in Go are immutable, so
// the returned byte slice must not be mutated.
//
// If it's mutated, this could lead to extremely
// strange issues: for example, strings may
// be inserted in maps. If the backing array under
// a string is mutated, that string will now
// be in the map at the wrong location.
//
// This function is based off code from
// https://groups.google.com/g/golang-nuts/c/Zsfk-VMd_fU/m/O1ru4fO-BgAJ.
func unsafeStringToBytes(s string) ([]byte, error) {
	const max = 0x7fff0000 // ~2 GiB
	if len(s) > max {
		return nil, errStringTooLong
	}
	if len(s) == 0 {
		// If s is the empty string, this branch preserves legacy
		// behavior that returns a nil byte slice rather than
		// potentially a non-nil, length=0 slice (which is otherwise
		// allowed by the combined specs of unsafe.StringData/Slice).
		return nil, nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s)), nil
}
