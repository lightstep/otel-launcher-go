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

package doevery

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	end := time.Now().Add(2 * time.Second)
	var invocations int
	for time.Now().Before(end) {
		TimePeriod(1*time.Second, func() {
			t.Logf("running at %v", time.Now())
			invocations++
		})
		time.Sleep(100 * time.Millisecond)
	}
	// It should be less, but with a little slop.
	// Without TimePeriod, we anticipate 20 invocations.
	// With TimePeriod anticipate 2 invocations.
	require.Less(t, invocations, 5)
}

func TestZero(t *testing.T) {
	end := time.Now().Add(3 * time.Second)
	var invocations int
	for time.Now().Before(end) {
		if invocations > 0 {
			break
		}
		TimePeriod(0, func() {
			t.Logf("running at %v", time.Now())
			invocations++
		})
		time.Sleep(100 * time.Millisecond)
	}
	// Basically just check that it ever executes.
	require.Greater(t, invocations, 0)
}

func TestConcurrentSamePC(t *testing.T) {
	var wg sync.WaitGroup
	var invocations int64

	end := time.Now().Add(2 * time.Second)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for time.Now().Before(end) {
				TimePeriod(1*time.Second, func() {
					t.Logf("running at %v", time.Now())
					atomic.AddInt64(&invocations, 1)
				})
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}
	wg.Wait()
	// It should be less, but with a little slop.
	// Without TimePeriod, we anticipate 20 invocations.
	// With TimePeriod anticipate 2 invocations.
	require.Less(t, invocations, int64(5))
}

func TestConcurrentDifferentPC(t *testing.T) {
	var wg sync.WaitGroup
	var invocations int64

	wg.Add(1)
	go func() {
		defer wg.Done()

		end := time.Now().Add(2 * time.Second)
		for time.Now().Before(end) {
			TimePeriod(1*time.Second, func() {
				t.Logf("running (0) at %v", time.Now())
				atomic.AddInt64(&invocations, 1)
			})
			time.Sleep(100 * time.Millisecond)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		end := time.Now().Add(2 * time.Second)
		for time.Now().Before(end) {
			TimePeriod(1*time.Second, func() {
				t.Logf("running (1) at %v", time.Now())
				atomic.AddInt64(&invocations, 1)
			})
			time.Sleep(100 * time.Millisecond)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		end := time.Now().Add(2 * time.Second)
		for time.Now().Before(end) {
			TimePeriod(1*time.Second, func() {
				t.Logf("running (2) at %v", time.Now())
				atomic.AddInt64(&invocations, 1)
			})
			time.Sleep(100 * time.Millisecond)
		}
	}()

	wg.Wait()
	// We expect exactly 6. Allow a little slop.
	require.GreaterOrEqual(t, invocations, int64(4))
}

func BenchmarkDoEvery(b *testing.B) {
	invocations := 0
	for i := 0; i < b.N; i++ {
		TimePeriod(1, func() {
			invocations++
		})
	}
	if invocations != b.N {
		b.Fatalf("incorrectness: %v != %v", invocations, b.N)
	}
}
