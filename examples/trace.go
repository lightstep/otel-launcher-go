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

package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/lightstep/otel-launcher-go/launcher"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/trace"
	"google.golang.org/grpc/codes"
)

func main() {
	otel := launcher.ConfigureOpentelemetry()
	defer otel.Shutdown()
	tracer := global.Tracer("ex.com/basic")

	tracer.WithSpan(context.Background(), "foo",
		func(ctx context.Context) error {
			tracer.WithSpan(ctx, "bar",
				func(ctx context.Context) error {
					tracer.WithSpan(ctx, "baz",
						func(ctx context.Context) error {
							getSpan(ctx)
							addAttribute(ctx)
							addEvent(ctx)
							recordException(ctx)
							createChild(ctx, tracer)
							return nil
						},
					)
					return nil
				},
			)
			return nil
		},
	)
	fmt.Println("OpenTelemetry example")
}

// example of getting the current span
func getSpan(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	fmt.Printf("current span: %v\n", span)
}

// example of adding an attribute to a span
func addAttribute(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.SetAttribute("attr1", "value1")
}

// example of adding an event to a span
func addEvent(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(ctx, "event1", []kv.KeyValue{
		kv.String("event-attr1", "event-string1"),
		kv.Int64("event-attr2", 10),
	}...)
}

// example of recording an exception
func recordException(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(ctx, errors.New("exception has occurred"))
	span.SetStatus(codes.Internal, "internal error")
}

// example of creating a child span
func createChild(ctx context.Context, tracer trace.Tracer) {
	// span := trace.SpanFromContext(ctx)
	_, childSpan := tracer.Start(ctx, "child")
	defer childSpan.End()
	fmt.Printf("child span: %v\n", childSpan)
}
