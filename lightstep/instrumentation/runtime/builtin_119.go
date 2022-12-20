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

func expectRuntimeMetrics() *builtinDescriptor {
	bd := newBuiltinDescriptor()
	bd.singleCounter("/cgo/go-to-c-calls:calls")
	bd.classesCounter("/gc/cycles/*:gc-cycles",
		attribute.NewSet(classKey.String("automatic")),
		attribute.NewSet(classKey.String("forced")),
	)
	bd.objectBytesCounter("/gc/heap/allocs:*")
	bd.objectBytesCounter("/gc/heap/frees:*")

	bd.singleGauge("/gc/heap/goal:bytes")

	bd.singleUpDownCounter("/gc/heap/objects:objects")

	bd.singleCounter("/gc/heap/tiny/allocs:objects")
	bd.singleGauge("/gc/limiter/last-enabled:gc-cycle")
	bd.singleGauge("/gc/stack/starting-size:bytes")
	bd.classesUpDownCounter("memory.usage",
		attribute.NewSet(classKey.String("heap"), subclassKey.String("free")),
		attribute.NewSet(classKey.String("heap"), subclassKey.String("objects")),
		attribute.NewSet(classKey.String("heap"), subclassKey.String("released")),
		attribute.NewSet(classKey.String("heap"), subclassKey.String("stacks")),
		attribute.NewSet(classKey.String("heap"), subclassKey.String("unused")),
		attribute.NewSet(classKey.String("metadata"), subclassKey.String("mcache"), subsubclassKey.String("free")),
		attribute.NewSet(classKey.String("metadata"), subclassKey.String("mcache"), subsubclassKey.String("inuse")),
		attribute.NewSet(classKey.String("metadata"), subclassKey.String("mspan"), subsubclassKey.String("free")),
		attribute.NewSet(classKey.String("metadata"), subclassKey.String("mspan"), subsubclassKey.String("inuse")),
		attribute.NewSet(classKey.String("metadata"), subclassKey.String("other")),
		attribute.NewSet(classKey.String("os-stacks")),
		attribute.NewSet(classKey.String("other")),
	)
	bd.singleGauge("sched.gomaxprocs")
	bd.singleUpDownCounter("sched.goroutines")
	return bd
}
