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

// package runtime generates metrics from the Golang runtime/metrics package.
//
// There are several conventions used to translate these metrics into
// the OpenTelemetry model.  Builtin metrics are defined in terms of
// the expected OpenTelemetry instrument kind in defs.go.
//
//  1. Single Counter, UpDownCounter, and Gauge instruments.  No
//  wildcards are used.  E.g., /cgo/go-to-c-calls:calls becomes
//  process.runtime.go.cgo.go-to-c-calls with unit {calls}.
//  2. Objects/Bytes Counter.  There are two runtime/metrics with the
//  same name and different units.  The objects counter has a suffix,
//  the bytes counter has a unit, to disambiguate.
//  3. Multi-dimensional Counter/UpDownCounter (generally), ignore any
//  "total" elements to avoid double-counting.  4. Multi-dimensional
//  Counter/UpDownCounter (named ".classes"), map to "usage" for bytes
//  and "time" for cpu-seconds.
//
// Histograms are not currently implemented, see the related issues
// for an explanation:
// https://github.com/open-telemetry/opentelemetry-specification/issues/2713
// https://github.com/open-telemetry/opentelemetry-specification/issues/2714

package runtime // import "github.com/lightstep/otel-launcher-go/lightstep/instrumentation/runtime"
