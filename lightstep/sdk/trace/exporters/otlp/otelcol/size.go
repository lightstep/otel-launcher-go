package otelcol

import (
	"fmt"
	math_bits "math/bits"

	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	traceapi "go.opentelemetry.io/otel/trace"
)

func sovTrace(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}

func sizeOfSpanContext(sc traceapi.SpanContext) int {
	sz := 0
	spanIDLen := len(sc.SpanID())
	sz += 1 + spanIDLen + sovTrace(uint64(spanIDLen))
	// sc.TraceFlags() is a byte
	sz += 1 + sovTrace(1)
	tid := len(sc.TraceID())
	sz += 1 + tid + sovTrace(uint64(tid))
	ts := sc.TraceState().String()
	lts := len(ts)
	sz += 1 + lts + sovTrace(uint64(lts))

	return sz
}

func sizeOfSpanString(str string) int {
	length := len(str)
	return 1 + length + sovTrace(uint64(length))
}

func sizeOfSpanAttributes(attrs []attribute.KeyValue) int {
	sz := 0

	for _, attr := range attrs {

		sz += sizeOfSpanString(string(attr.Key))

		switch v := attr.Value.Type(); v {
		case attribute.BOOL:
			sz += 1 + sovTrace(1)
		case attribute.BOOLSLICE:
			l := len(attr.Value.AsBoolSlice())
			sz += 1 + l + sovTrace(uint64(l))
		case attribute.INT64:
			sz += 1 + sovTrace(8)
		case attribute.INT64SLICE:
			l := 8 * len(attr.Value.AsInt64Slice())
			sz += 1 + l + sovTrace(uint64(l))
		case attribute.FLOAT64:
			sz += 1 + sovTrace(8)
		case attribute.FLOAT64SLICE:
			l := 8 * len(attr.Value.AsFloat64Slice())
			sz += 1 + l + sovTrace(uint64(l))
		case attribute.STRING:
			l := len(attr.Value.AsString())
			sz += 1 + l + sovTrace(uint64(l))
		case attribute.STRINGSLICE:
			for _, s := range attr.Value.AsStringSlice() {
				l := len(s)
				sz += 1 + l + sovTrace(uint64(l))
			}
		}

	}
	return sz
}

func sizeOfSpanLinks(links []trace.Link) int {
	sz := 0

	for _, link := range links {
		sz += sizeOfSpanContext(link.SpanContext)
		sz += sizeOfSpanAttributes(link.Attributes)
		if link.DroppedAttributeCount != 0 {
			sz += 1 + sovTrace(uint64(link.DroppedAttributeCount))
		}
	}
	fmt.Println("LINK SZ")
	fmt.Println(sz)
	return sz
}

func sizeOfSpanEvents(events []trace.Event) int {
	sz := 0

	for _, event := range events {
		sz += sizeOfSpanString(event.Name)
		sz += sizeOfSpanAttributes(event.Attributes)
		// event.Time is a time.Time
		sz += 1 + 24 + sovTrace(24)
		if event.DroppedAttributeCount != 0 {
			sz += 1 + sovTrace(uint64(event.DroppedAttributeCount))
		}
	}
	return sz
}


func sizeOfROSpan(span trace.ReadOnlySpan) int {
	sz := 0
	sz += sizeOfSpanString(span.Name())

	sz += sizeOfSpanContext(span.SpanContext())

	sz += sizeOfSpanContext(span.Parent())

	sz += sizeOfSpanString(span.SpanKind().String())

	// span.StartTime() and span.EndTime() is a time.Time that consists of a two int64
	// and an optional pointer for location. To be conservative count 24 bytes.
	sz += 1 + 24 + sovTrace(uint64(24))
	sz += 1 + 24 + sovTrace(uint64(24))

	sz += sizeOfSpanAttributes(span.Attributes())

	sz += sizeOfSpanLinks(span.Links())

	sz += sizeOfSpanEvents(span.Events())

	if span.Status().Code != otelcodes.Unset {
		sz += 1 + sovTrace(uint64(span.Status().Code))
	}
	sz += sizeOfSpanString(span.Status().Description)

	sz += sizeOfSpanString(span.InstrumentationScope().Name)
	sz += sizeOfSpanString(span.InstrumentationScope().Version)
	sz += sizeOfSpanString(span.InstrumentationScope().SchemaURL)

	sz += sizeOfSpanString(span.InstrumentationLibrary().Name)
	sz += sizeOfSpanString(span.InstrumentationLibrary().Version)
	sz += sizeOfSpanString(span.InstrumentationLibrary().SchemaURL)

	sz += sizeOfSpanString(span.Resource().SchemaURL())
	sz += sizeOfSpanAttributes(span.Resource().Attributes())

	if span.DroppedAttributes() != 0 {
		sz += 1 + sovTrace(uint64(span.DroppedAttributes()))
	}

	if span.DroppedLinks() != 0 {
		sz += 1 + sovTrace(uint64(span.DroppedLinks()))
	}

	if span.DroppedEvents() != 0 {
		sz += 1 + sovTrace(uint64(span.DroppedEvents()))
	}
	if span.ChildSpanCount() != 0 {
		sz += 1 + sovTrace(uint64(span.ChildSpanCount()))
	}

	return sz
}
