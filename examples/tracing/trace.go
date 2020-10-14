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
	"os"

	"github.com/lightstep/otel-launcher-go/launcher"
	"go.opentelemetry.io/collector/translator/conventions"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	otel := launcher.ConfigureOpentelemetry()
	defer otel.Shutdown()
	tracer := global.Tracer("ex.com/basic")

	ctx0 := context.Background()

	ctx1, finish1 := tracer.Start(ctx0, "foo")
	defer finish1.End()

	ctx2, finish2 := tracer.Start(ctx1, "bar")
	defer finish2.End()

	ctx3, finish3 := tracer.Start(ctx2, "baz")
	defer finish3.End()

	ctx := ctx3
	getSpan(ctx)
	addAttribute(ctx)
	addEvent(ctx)
	recordException(ctx)
	createChild(ctx, tracer)
	setResourceAttributes()
	setBaggage()

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
	span.SetAttributes(label.String("attr1", "value1"))
}

// example of adding an event to a span
func addEvent(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(ctx, "event1", []label.KeyValue{
		label.String("event-attr1", "event-string1"),
		label.Int64("event-attr2", 10),
	}...)
}

// example of recording an exception
func recordException(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(ctx, errors.New("exception has occurred"))
	span.SetStatus(codes.Error, "internal error")
}

// example of creating a child span
func createChild(ctx context.Context, tracer trace.Tracer) {
	// span := trace.SpanFromContext(ctx)
	_, childSpan := tracer.Start(ctx, "child")
	defer childSpan.End()
	fmt.Printf("child span: %v\n", childSpan)
}

// example of setting resource attributes
func setResourceAttributes() {
	host, _ := os.Hostname()
	attributes := []label.KeyValue{
		label.String(conventions.AttributeServiceName, "service123"),
		label.String(conventions.AttributeServiceVersion, "1.2.3"),
		label.String(conventions.AttributeHostName, host),
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource.New(attributes...)),
	)
	global.SetTracerProvider(tp)
}

// example of setting baggage
func setBaggage() {
	ctx := otel.ContextWithBaggageValues(context.Background(),
		label.String("keyone", "foo1"),
		label.String("keytwo", "bar1"),
	)
	bags := otel.Baggage(ctx)

	for iter := bags.Iter(); iter.Next(); {
		fmt.Printf("key %s: %s\n", iter.Label().Key, iter.Label().Value.AsString())
	}
}
