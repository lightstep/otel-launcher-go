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

//go:build go1.19 && !go1.20

package runtime

import "go.opentelemetry.io/otel/attribute"

var expectRuntimeMetrics = map[builtinNameExpected]map[attribute.Set]bool{
	expectCounter("cgo.go-to-c-calls"): expectSingleton,
	expectCounter("gc.cycles"): map[attribute.Set]bool{
		attribute.NewSet(classKey.String("automatic")): true,
		attribute.NewSet(classKey.String("forced")):    true,
	},
	expectCounter("gc.heap.allocs"):                nil,
	expectCounter("gc.heap.allocs.objects"):        nil,
	expectCounter("gc.heap.frees"):                 nil,
	expectCounter("gc.heap.frees.objects"):         nil,
	expectUpDownCounter("gc.heap.goal"):            nil,
	expectCounter("gc.heap.objects"):               nil,
	expectCounter("gc.heap.tiny.allocs"):           nil,
	expectUpDownCounter("gc.limiter.last-enabled"): nil,
	expectUpDownCounter("gc.stack.starting-size"):  nil,
	expectUpDownCounter("memory.usage"): map[attribute.Set]bool{
		attribute.NewSet(classKey.String("heap"), subclassKey.String("free")): true,
	},
	expectUpDownCounter("sched.gomaxprocs"): nil,
	expectUpDownCounter("sched.goroutines"): nil,
}
