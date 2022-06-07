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

// package aggregator defines the interface used between the SDK and
// various kinds of aggregation.
//
// Note, about the use of Go-1.18 generics!  This code makes use of a
// pattern that gives the SDK fine control over memory layout via the
// use of a `Storage any` type constraint.  The SDK is expected to
// pair the allocated `Storage any` value with corresponding `Methods`
// type, which provides pointer-receiver methods that operate on one
// or two `Storage` objects.
//
// This means there are places where a Storage object may appear in a
// generic type.  Where that happens without the corresponding
// Methods, it means the Storage is opaque at that level in the code.
package aggregator // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
