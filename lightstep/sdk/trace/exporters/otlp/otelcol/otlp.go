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
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/internal"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	apitrace "go.opentelemetry.io/otel/trace"
)

func copyEvents(dest ptrace.SpanEventSlice, events []trace.Event) {
	for _, event := range events {
		e1 := dest.AppendEmpty()
		e1.SetDroppedAttributesCount(uint32(event.DroppedAttributeCount))
		e1.SetName(event.Name)
		e1.SetTimestamp(pcommon.NewTimestampFromTime(event.Time))

		internal.CopyAttributes(e1.Attributes(), attribute.NewSet(event.Attributes...))
	}
}

func copyLinks(dest ptrace.SpanLinkSlice, links []trace.Link) {
	for _, link := range links {
		l1 := dest.AppendEmpty()
		l1.SetDroppedAttributesCount(uint32(link.DroppedAttributeCount))
		l1.SetSpanID(pcommon.SpanID(link.SpanContext.SpanID()))
		l1.SetTraceID(pcommon.TraceID(link.SpanContext.TraceID()))
		l1.TraceState().FromRaw(link.SpanContext.TraceState().String())

		internal.CopyAttributes(l1.Attributes(), attribute.NewSet(link.Attributes...))
	}
}

func translateSpanStatusCode(in codes.Code) ptrace.StatusCode {
	switch in {
	case codes.Error:
		return ptrace.StatusCodeError
	case codes.Ok:
		return ptrace.StatusCodeOk
	default:
		return ptrace.StatusCodeUnset
	}
}

func translateSpanKind(in apitrace.SpanKind) ptrace.SpanKind {
	switch in {
	case apitrace.SpanKindInternal:
		return ptrace.SpanKindInternal
	case apitrace.SpanKindServer:
		return ptrace.SpanKindServer
	case apitrace.SpanKindClient:
		return ptrace.SpanKindClient
	case apitrace.SpanKindProducer:
		return ptrace.SpanKindProducer
	case apitrace.SpanKindConsumer:
		return ptrace.SpanKindConsumer
	default:
		return ptrace.SpanKindUnspecified
	}
}

func (c *client) d2pd(in []trace.ReadOnlySpan) ptrace.Traces {
	out := ptrace.NewTraces()

	if len(in) == 0 {
		return out
	}
	rs := out.ResourceSpans().AppendEmpty()

	c.ResourceMap.Get(in[0].Resource()).CopyTo(rs.Resource())

	curName := ""
	var ss ptrace.ScopeSpans
	for _, tr := range in {
		// input is list of ReadOnlySpans. Add spans to the same scopespan
		// until new scope name is encountered.
		if ssName := tr.InstrumentationScope().Name; ssName != curName {
			curName = ssName
			ss = rs.ScopeSpans().AppendEmpty()
			ss.SetSchemaUrl(tr.InstrumentationScope().SchemaURL)
			ss.Scope().SetName(tr.InstrumentationScope().Name)
			ss.Scope().SetVersion(tr.InstrumentationScope().Version)
		}

		s := ss.Spans().AppendEmpty()
		s.SetDroppedAttributesCount(uint32(tr.DroppedAttributes()))
		s.SetDroppedEventsCount(uint32(tr.DroppedEvents()))
		s.SetDroppedLinksCount(uint32(tr.DroppedLinks()))
		s.SetKind(translateSpanKind(tr.SpanKind()))
		s.SetName(tr.Name())
		s.SetParentSpanID(pcommon.SpanID(tr.Parent().SpanID()))
		s.SetSpanID(pcommon.SpanID(tr.SpanContext().SpanID()))
		s.SetStartTimestamp(pcommon.NewTimestampFromTime(tr.StartTime()))
		s.SetEndTimestamp(pcommon.NewTimestampFromTime(tr.EndTime()))
		s.SetTraceID(pcommon.TraceID(tr.SpanContext().TraceID()))
		s.TraceState().FromRaw(tr.SpanContext().TraceState().String())
		s.Status().SetCode(translateSpanStatusCode(tr.Status().Code))
		s.Status().SetMessage(tr.Status().Description)

		internal.CopyAttributes(s.Attributes(), attribute.NewSet(tr.Attributes()...))
		copyEvents(s.Events(), tr.Events())
		copyLinks(s.Links(), tr.Links())
	}

	return out
}
