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

package otelcol

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestSizeOfSimpleROSpan(t *testing.T) {

	tests := []struct {
		name          string
		numSpanAttrs  int
		numSpanEvents int
		numSpanLinks  int
	}{
		{
			name:         "simple",
			numSpanAttrs: 10,
		},
		{
			name:         "many_attributes",
			numSpanAttrs: 100000,
		},
		{
			name:          "with_events",
			numSpanAttrs:  100000,
			numSpanEvents: 10000,
		},
		{
			name:          "with_links",
			numSpanAttrs:  100000,
			numSpanEvents: 10000,
			numSpanLinks:  1000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tprovider := sdktrace.NewTracerProvider(
				sdktrace.WithResource(
					resource.NewSchemaless(testResourceAttrs...),
				),
			)
			tracer := tprovider.Tracer("test-tracer")
			_, span := tracer.Start(ctx, "ExecuteRequest", trace.WithSpanKind(trace.SpanKindServer))

			for i := 0; i < tt.numSpanAttrs; i++ {
				key := fmt.Sprintf("test-attribute-%d", i)
				val := fmt.Sprintf("test-value-%d", i)
				span.SetAttributes(attribute.String(key, val))
			}

			for i := 0; i < tt.numSpanEvents; i++ {
				key := fmt.Sprintf("test-event-attribute-%d", i)
				val := fmt.Sprintf("test-event-value-%d", i)
				span.AddEvent("test event", trace.WithAttributes(attribute.String(key, val)))
			}

			for i := 0; i < tt.numSpanLinks; i++ {
				key := fmt.Sprintf("test-link-attribute-%d", i)
				val := fmt.Sprintf("test-link-value-%d", i)
				linkCtx, _ := tracer.Start(context.Background(), "LinkedRequest", trace.WithSpanKind(trace.SpanKindServer))

				link := trace.LinkFromContext(linkCtx, attribute.String(key, val))

				span.AddLink(link)
			}

			span.SetStatus(codes.Error, "failed")
			span.End()

			roSpan := span.(sdktrace.ReadOnlySpan)

			c := &client{}
			traces := ptraceotlp.NewExportRequestFromTraces(c.d2pd([]sdktrace.ReadOnlySpan{roSpan}))
			data, err := traces.MarshalProto()
			require.NoError(t, err)

			protobufSz := float64(len(data))
			ROSpanSz := float64(sizeOfROSpan(roSpan))
			percentDiff := math.Abs((ROSpanSz - protobufSz) / protobufSz)

			assert.LessOrEqual(t, percentDiff, 0.1)
		})
	}
}
