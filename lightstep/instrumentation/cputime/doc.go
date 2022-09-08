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

// Package cputime provides the process CPU time metrics specified by
// OpenTelemetry, as observed by the process itself.  These are close
// to runtime metrics, but they not currently provided by the Go
// runtime/metrics package.  These metrics are included together
// because they are comparable.  Note that the runtime.MemStats
// definition of GCCPUFraction can be derived as:
//
//     GCCPUFraction = process.runtime.go.gc.cpu.time /
//                     (process.runtime.uptime *
//                      process.runtime.go.sched.gomaxprocs)
//
//
// where:
//
//     process.runtime.go.gc.cpu.time < sum(process.cpu.time)
//
// See also https://github.com/open-telemetry/opentelemetry-go-contrib/issues/316.
//
// The metrics produced are listed here with attribute dimensions.
//
//	Name			Attribute
//
// ----------------------------------------------------------------------
//
//	process.cpu.time           state=user|system
//      process.uptime
//      process.runtime.go.gc.cpu.time
//
// See https://github.com/open-telemetry/oteps/blob/main/text/0119-standard-system-metrics.md
// for the definition of these metric instruments.

package cputime // import "github.com/lightstep/otel-launcher-go/lightstep/instrumentation/cputime"
