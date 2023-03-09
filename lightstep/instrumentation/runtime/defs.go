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

package runtime

func expectRuntimeMetrics() *builtinDescriptor {
	bd := newBuiltinDescriptor()
	bd.classesCounter("/cpu/classes/*:cpu-seconds")
	bd.classesCounter("/gc/cycles/*:gc-cycles")
	bd.classesUpDownCounter("/memory/classes/*:bytes")
	bd.ignoreHistogram("/gc/heap/allocs-by-size:bytes")
	bd.ignoreHistogram("/gc/heap/frees-by-size:bytes")
	bd.ignoreHistogram("/gc/pauses:seconds")
	bd.ignoreHistogram("/sched/latencies:seconds")
	bd.objectBytesCounter("/gc/heap/allocs:*")
	bd.objectBytesCounter("/gc/heap/frees:*")
	bd.singleCounter("/cgo/go-to-c-calls:calls")
	bd.singleCounter("/gc/heap/tiny/allocs:objects")
	bd.singleCounter("/sync/mutex/wait/total:seconds")
	bd.singleGauge("/gc/heap/goal:bytes")
	bd.singleGauge("/gc/limiter/last-enabled:gc-cycle")
	bd.singleGauge("/gc/stack/starting-size:bytes")
	bd.singleGauge("/sched/gomaxprocs:threads")
	bd.singleUpDownCounter("/gc/heap/objects:objects")
	bd.singleUpDownCounter("/sched/goroutines:goroutines")
	return bd
}
