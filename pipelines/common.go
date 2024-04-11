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
	otelcolmetric "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp/otelcol"
	otelcoltrace "github.com/lightstep/otel-launcher-go/lightstep/sdk/trace/exporters/otlp/otelcol"

	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc/credentials"
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

	// TemporalityPreference is one of "cumulative", "delta", or "lowmemory" (a.k.a. "stateless")
	TemporalityPreference string

	// TLSSetting is the newer form.
	TLSSetting *configtls.ClientConfig

	// The Lightstep SDK is always used, the OTel-Go SDK is no longer
	// an option supported by this library.
	DeprecatedUseLightstepMetricsSDK bool
	// Credentials carries the TLS settings used by OTel-Go SDKs.
	// This is not used.
	DeprecatedCredentials credentials.TransportCredentials
}

type PipelineSetupFunc func(PipelineConfig) (func() error, error)

func (p PipelineConfig) secureMetricOption() otelcolmetric.Option {
	if p.Insecure {
		return otelcolmetric.WithInsecure()
	} else if p.TLSSetting != nil {
		return otelcolmetric.WithTLSSetting(*p.TLSSetting)
	}
	return otelcolmetric.WithTLSSetting(configtls.ClientConfig{
		Insecure: false,
	})
}

func (p PipelineConfig) secureTraceOption() otelcoltrace.Option {
	if p.Insecure {
		return otelcoltrace.WithInsecure()
	} else if p.TLSSetting != nil {
		return otelcoltrace.WithTLSSetting(*p.TLSSetting)
	}
	return otelcoltrace.WithTLSSetting(configtls.ClientConfig{
		Insecure: false,
	})
}
