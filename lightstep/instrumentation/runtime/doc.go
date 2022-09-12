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

// package runtime geneartes metrics run the Golang runtime/metrics package.
//
// There are two special policies that are used to translate these
// metrics into the OpenTelemetry model.
//
//  1. The runtime/metrics name is split into its name and unit part;
//     when there are two metrics with the same name and different
//     units, the only known case is where "objects" and "bytes" are
//     present.  In this case, the outputs are a unitless metric (with
//     suffix, e.g., ending `gc.heap.allocs.objects`) and a unitful
//     metric with no suffix (e.g., ending `gc.heap.allocs` having
//     bytes units).
//  2. When there are >= 2 metrics with the same prefix and one
//     matching `prefix.total`, the total is skipped and the other
//     members are assembled into a single Counter or UpDownCounter
//     metric with multiple attribute values.  The supported cases
//     are for `class` and `cycle` attributes.
//
// The following metrics are generated in go-1.19.
//
// Name                                                    Unit          Instrument
// ------------------------------------------------------------------------------------
// process.runtime.go.cgo.go-to-c-calls                    {calls}       Counter[int64]
// process.runtime.go.gc.cycles{cycle=forced,automatic}    {gc-cycles}   Counter[int64]
// process.runtime.go.gc.heap.allocs                       bytes (*)     Counter[int64]
// process.runtime.go.gc.heap.allocs.objects               {objects} (*) Counter[int64]
// process.runtime.go.gc.heap.allocs-by-size               bytes         Histogram[float64] (**)
// process.runtime.go.gc.heap.frees                        bytes (*)     Counter[int64]
// process.runtime.go.gc.heap.frees.objects                {objects} (*) Counter[int64]
// process.runtime.go.gc.heap.frees-by-size                bytes         Histogram[float64] (**)
// process.runtime.go.gc.heap.goal                         bytes         UpDownCounter[int64]
// process.runtime.go.gc.heap.objects                      {objects}     UpDownCounter[int64]
// process.runtime.go.gc.heap.tiny.allocs                  {objects}     Counter[int64]
// process.runtime.go.gc.limiter.last-enabled              {gc-cycle}    UpDownCounter[int64]
// process.runtime.go.gc.pauses                            seconds       Histogram[float64] (**)
// process.runtime.go.gc.stack.starting-size               bytes         UpDownCounter[int64]
// process.runtime.go.memory.usage{class=...}              bytes         UpDownCounter[int64]
// process.runtime.go.sched.gomaxprocs                     {threads}     UpDownCounter[int64]
// process.runtime.go.sched.goroutines                     {goroutines}  UpDownCounter[int64]
// process.runtime.go.sched.latencies                      seconds       GaugeHistogram[float64] (**)
//
// (*) Empty unit strings are cases where runtime/metric produces
// duplicate names ignoring the unit string (see policy #1).
// (**) Histograms are not currently implemented, see the related
// issues for an explanation:
// https://github.com/open-telemetry/opentelemetry-specification/issues/2713
// https://github.com/open-telemetry/opentelemetry-specification/issues/2714

package runtime // import "github.com/lightstep/otel-launcher-go/lightstep/instrumentation/runtime"
