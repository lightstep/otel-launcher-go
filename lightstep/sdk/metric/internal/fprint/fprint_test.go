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
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestMethods(t *testing.T) {
	require.NotEqual(t, Mix(1, 2), Mix(2, 1))
	require.NotEqual(t, FingerprintString("hello"), FingerprintString("world"))
	require.NotEqual(t, FingerprintInt64(13), FingerprintInt64(14))
	require.NotEqual(t, FingerprintBool(true), FingerprintBool(false))
	require.NotEqual(t, FingerprintFloat64(1.0), FingerprintFloat64(2.0))
}

func TestConsistent(t *testing.T) {
	var fp uint64 = 0xc43fb29ab5effcfe // for "foobar"

	f := func() error {
		for i := 0; i < 100; i++ {
			ss := []string{
				"foobar",
				mkHeapString("foobar"),
			}
			runtime.GC()
			for _, s := range ss {
				runtime.GC()
				// Perform a bunch of fingerprints and check that they match the golden value.
				for i, res := range []uint64{
					FingerprintString(s),
					unsafeFingerprintString(s),
				} {
					if res != fp {
						return fmt.Errorf("%d: %v != %v", i, res, fp)
					}
				}
			}
		}
		return nil
	}

	var eg errgroup.Group
	for i := 0; i < 10; i++ {
		eg.Go(f)
	}

	require.NoError(t, eg.Wait())
}
