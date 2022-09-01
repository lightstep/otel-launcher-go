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

// package doevery provides primitives for per-call-site rate-limiting.
package doevery

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

var (
	// mu protects below.
	mu sync.Mutex

	// mostRecentInvocationMap maintains the last time the function passed at
	// the caller's program counter (PC) was called.
	// TODO: consider sharding this e.g. [8]map[invocationKey]time.Time
	// if it bottlenecks.
	mostRecentInvocationMap = make(map[invocationKey]time.Time)
)

// invocationKey is an identifier for a unique
// line of source code.
type invocationKey struct {
	// file is the filename used to compile the
	// line of code we are rate-limiting e.g.
	// go/src/pkg/main.go
	file string
	// line is the line number in file corresponding to
	// the line of code we are rate-limiting.
	line int
}

// TimePeriod rate limits each call site of this by the duration specified
// as the first argument. This may be useful in logging scenarios, where
// you only want to log every few seconds - or every second - instead of
// tens of hundreds of times per second.
//
// Each unique call site that calls TimePeriod is rate-limited independently.
// Each invocation of TimePeriod at the same call-site should provide the
// same duration.
//
// Example usage:
// 	end := time.Now().Add(5 * time.Second)
//	for end.After(time.Now()) {
//		doevery.TimePeriod(1*time.Second, func() {
//			fmt.Println("This will only appear once per second.")
//		})
//	}
//
// Please note that each individual thread does not have a distinct
// rate-limit; the rate-limit is global for the file/line.
//
// TimePeriod is safe for concurrent use.
func TimePeriod(dur time.Duration, f func()) {
	if dur < 0 {
		panic(fmt.Sprintf("negative duration unsupported: %v", dur))
	}
	// Find our unique location so we can check when we last invoked f.
	// Skip 0 is us (TimePeriod); skip 1 is our caller.
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		// If we don't know our own caller, we can't help.
		// We can either fail open or fail closed, here we choose
		// to fail open.
		f()
		return
	}

	// Use the file/line as the source-of-truth for
	// deduping invocations. We can't use the program counter (PC)
	// as the PC can differ for the same LoC that is
	// inlined in multiple places.
	key := invocationKey{
		file: file,
		line: line,
	}

	shouldInvoke := func() bool {
		mu.Lock()
		defer mu.Unlock()

		prevInvocation, ok := mostRecentInvocationMap[key]

		invoking := !ok || time.Since(prevInvocation) > dur
		if invoking {
			mostRecentInvocationMap[key] = time.Now()
		}
		return invoking
	}()

	if !shouldInvoke {
		// Just return early, nothing to change.
		return
	}

	// Invoke. We already updated the time.
	f()
}
