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

// Package otlpmetricgrpc provides an otlpmetric.Client that communicates
// with an OTLP receiving endpoint using gRPC.  This is a fork of the client
// code in "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc",
// except the `NewClient()` function is exported, which allows us to use it
// from the otlpmetric.Exporter.
package otlpmetricgrpc // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp/otlpmetric/otlpmetricgrpc"
