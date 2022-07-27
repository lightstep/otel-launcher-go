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
	"math"

	// Our use of farmhash is sort of arbitrary: we want a fast,
	// fingerprint function and farmhash is familiar (to
	// Lightstep).  We do not require a fixed (never-changing)
	// hash function here, althought it helps with testability.
	farm "github.com/dgryski/go-farm"
)

// Mix combines two fingerprints together.  This function and the
// constant multiplier are copied from farmhash source.
func Mix(x uint64, y uint64) uint64 {
	const mul uint64 = 0x9ddfea08eb382d69
	a := (x ^ y) * mul
	a ^= a >> 47
	b := (y ^ a) * mul
	b ^= b >> 47
	b *= mul
	return b
}

func FingerprintInt64(i int64) uint64 {
	return uint64(i)
}

func FingerprintBool(b bool) uint64 {
	if b {
		return uint64(1)
	}
	return uint64(0)
}

func FingerprintFloat64(f float64) uint64 {
	return math.Float64bits(f)
}

func unsafeFingerprintString(s string) uint64 {
	bs, err := unsafeStringToBytes(s)
	if err != nil {
		// Gnarly! There should be an attribute size limit before this happens.
		bs = []byte(s)
	}
	return farm.Fingerprint64(bs)
}

func FingerprintString(s string) uint64 {
	// We know that the go-farm implementation
	// we use does not modify the []byte it is passed,
	// so we use an unsafe conversion here from string to
	// []byte to avoid a copy.
	return unsafeFingerprintString(s)
}
