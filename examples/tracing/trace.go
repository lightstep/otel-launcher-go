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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func main() {

	attributes := map[string]string{
		"book.condition": "good",
		"book.category":  "science",
	}

	lsLauncher := launcher.ConfigureOpentelemetry(
		launcher.WithResourceAttributes(attributes),
	)
	defer lsLauncher.Shutdown()
	tracer := otel.Tracer("ex.com/basic")

	ctx0 := context.Background()

	ctx1, finish1 := tracer.Start(ctx0, "foo")
	defer finish1.End()

	ctx2, finish2 := tracer.Start(ctx1, "bar")
	defer finish2.End()

	ctx3, finish3 := tracer.Start(setBaggage(ctx2), "baz")
	defer finish3.End()

	ctx := ctx3
	getSpan(ctx)
	addAttribute(ctx)
	addEvent(ctx)
	recordException(ctx)
	createChild(ctx, tracer)

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
	span.SetAttributes(attribute.String("attr1", "value1"))
}

// example of adding an event to a span
func addEvent(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("event1", trace.WithAttributes([]attribute.KeyValue{
		attribute.String("event-attr1", "event-string1"),
		attribute.Int64("event-attr2", 10),
	}...))
}

// example of recording an exception
func recordException(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(errors.New("exception has occurred"))
	span.SetStatus(codes.Error, "internal error")
}

// example of creating a child span
func createChild(ctx context.Context, tracer trace.Tracer) {
	_, childSpan := tracer.Start(ctx, "child")
	defer childSpan.End()
	fmt.Printf("child span: %v\n", childSpan)
}

// example of setting baggage
func setBaggage(ctx context.Context) context.Context {
	mem1, _ := baggage.NewMember("keyone", "foo1")
	mem2, _ := baggage.NewMember("keytwo", "bar1")
	bag, _ := baggage.New(mem1, mem2)

	return baggage.ContextWithBaggage(context.Background(), bag)
}
