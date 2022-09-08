// Copyright Lightstep Authors
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

package pipelines

import (
	"google.golang.org/grpc/credentials"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
)

type PipelineConfig struct {
	Endpoint        string
	Insecure        bool
	Headers         map[string]string
	Resource        *resource.Resource
	ReportingPeriod string
	Propagators     []string

	// MetricsBuiltinsEnabled indicates whether to automatically start
	// standard host and runtime metrics.
	MetricsBuiltinsEnabled bool

	// MetricsBuiltinLibraries contains strings identifying which
	// builtin metrics libraries should be started.  The entry is
	// a single hosrt name (e.g., "host", "cpu", "runtime")
	// followed by an optional major version number (e.g.,
	// "host:v0", "cpu:v0", "runtime:v0").  Short names are mapped
	// to long names internally.
	//
	// Recognized names, presently:
	//
	//  runtime: v0 is go-contrib/instrumentation/runtime
	//           v1 is lightstep/instrumentation/runtime
	//  host:    v0 is go-contrib/instrumentation/host
	//           v1 is lightstep/instrumentation/host
	//  cputime: v1 is lightstep/instrumentation/cputime
	MetricsBuiltinLibraries []string

	// TemporalityPreference is one of "cumulative", "delta", or "stateless"
	TemporalityPreference string

	// Credentials carries the TLS settings.
	Credentials credentials.TransportCredentials

	// UseLightstepMetricsSDK determines whether to use the metrics
	// SDK at ../lightstep/sdk/metric.
	UseLightstepMetricsSDK bool
}

type PipelineSetupFunc func(PipelineConfig) (func() error, error)

func (p PipelineConfig) secureMetricOption() otlpmetricgrpc.Option {
	if p.Insecure {
		return otlpmetricgrpc.WithInsecure()
	} else if p.Credentials != nil {
		return otlpmetricgrpc.WithTLSCredentials(p.Credentials)
	}
	return otlpmetricgrpc.WithTLSCredentials(
		credentials.NewClientTLSFromCert(nil, ""),
	)
}

func (p PipelineConfig) secureTraceOption() otlptracegrpc.Option {
	if p.Insecure {
		return otlptracegrpc.WithInsecure()
	} else if p.Credentials != nil {
		return otlptracegrpc.WithTLSCredentials(p.Credentials)
	}
	return otlptracegrpc.WithTLSCredentials(
		credentials.NewClientTLSFromCert(nil, ""),
	)
}
