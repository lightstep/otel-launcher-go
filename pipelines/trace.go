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
	"context"
	"fmt"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/trace/exporters/otlp/otelcol"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/encoding/gzip"
)

func NewTracePipeline(c PipelineConfig) (func() error, error) {
	spanExporter, err := c.newTraceExporter(c.secureTraceOption())
	if err != nil {
		return nil, fmt.Errorf("failed to create span exporter: %v", err)
	}

	// Note: this processor does not metric the spans it drops.
	// TODO: either improve the spec, so the otel-go SDK BSP will do
	// this, or else add the lightstep-internal BSP w/ export metrics
	// to this repo.
	bsp := trace.NewBatchSpanProcessor(spanExporter)
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSpanProcessor(bsp),
		trace.WithResource(c.Resource),
	)

	if err = configurePropagators(c); err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tp)

	return func() error {
		return bsp.Shutdown(context.Background())
	}, nil
}

func (c PipelineConfig) newTraceExporter(secure otelcol.Option) (trace.SpanExporter, error) {
	return otelcol.NewExporter(
		context.Background(),
		otelcol.NewConfig(
			secure,
			otelcol.WithEndpoint(c.Endpoint),
			otelcol.WithHeaders(c.Headers),
			otelcol.WithCompressor(gzip.Name),
		),
	)
}

// configurePropagators configures B3 propagation by default
func configurePropagators(c PipelineConfig) error {
	propagatorsMap := map[string]propagation.TextMapPropagator{
		"b3":           b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)),
		"baggage":      propagation.Baggage{},
		"tracecontext": propagation.TraceContext{},
		"ottrace":      ot.OT{},
	}
	var props []propagation.TextMapPropagator
	for _, key := range c.Propagators {
		prop := propagatorsMap[key]
		if prop != nil {
			props = append(props, prop)
		}
	}
	if len(props) == 0 {
		return fmt.Errorf("invalid configuration: unsupported propagators. Supported options: b3,baggage,tracecontext,ottrace")
	}
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		props...,
	))
	return nil
}
