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
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
)

func copyAttributes(dest pcommon.Map, src attribute.Set) {
	for iter := src.Iter(); iter.Next(); {
		inA := iter.Attribute()
		key := string(inA.Key)
		switch inA.Value.Type() {
		case attribute.BOOL:
			dest.PutBool(key, inA.Value.AsBool())
		case attribute.INT64:
			dest.PutInt(key, inA.Value.AsInt64())
		case attribute.FLOAT64:
			dest.PutDouble(key, inA.Value.AsFloat64())
		case attribute.STRING:
			dest.PutStr(key, inA.Value.AsString())
		case attribute.BOOLSLICE:
			sl := dest.PutEmptySlice(key)
			sl.EnsureCapacity(len(inA.Value.AsBoolSlice()))
			for _, v := range inA.Value.AsBoolSlice() {
				sl.AppendEmpty().SetBool(v)
			}
		case attribute.INT64SLICE:
			sl := dest.PutEmptySlice(key)
			sl.EnsureCapacity(len(inA.Value.AsInt64Slice()))
			for _, v := range inA.Value.AsInt64Slice() {
				sl.AppendEmpty().SetInt(v)
			}
		case attribute.FLOAT64SLICE:
			sl := dest.PutEmptySlice(key)
			sl.EnsureCapacity(len(inA.Value.AsFloat64Slice()))
			for _, v := range inA.Value.AsFloat64Slice() {
				sl.AppendEmpty().SetDouble(v)
			}
		case attribute.STRINGSLICE:
			sl := dest.PutEmptySlice(key)
			sl.EnsureCapacity(len(inA.Value.AsStringSlice()))
			for _, v := range inA.Value.AsStringSlice() {
				sl.AppendEmpty().SetStr(v)
			}
		default:
			panic("unhandled case")
		}
	}
}

func (c *client) d2pd(in []trace.ReadOnlySpan) ptrace.Traces {
	c.once.Do(func() {
		c.resource = pcommon.NewResource()
		for _, tr := range in {
			copyAttributes(
				c.resource.Attributes(),
				attribute.NewSet(tr.Resource().Attributes()...),
			)
		}
	})
	out := ptrace.NewTraces()
	rs := out.ResourceSpans().AppendEmpty()

	c.resource.CopyTo(rs.Resource())

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
		s.SetKind(ptrace.SpanKind(tr.SpanKind()))
		s.SetName(tr.Name())
		s.SetParentSpanID(pcommon.SpanID(tr.Parent().SpanID()))
		s.SetSpanID(pcommon.SpanID(tr.SpanContext().SpanID()))
		s.SetStartTimestamp(pcommon.NewTimestampFromTime(tr.StartTime()))
		s.SetEndTimestamp(pcommon.NewTimestampFromTime(tr.EndTime()))
		s.SetTraceID(pcommon.TraceID(tr.SpanContext().TraceID()))
	}

	return out
}
