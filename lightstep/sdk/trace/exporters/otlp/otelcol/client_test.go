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
	"encoding/hex"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/otelarrowreceiver"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/trace/v1"

	// "google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

// Note: unclear which test support library we should use until this
// moves into an otel repo.  Some simple test supports are developed
// here anyway to defer this question.

var (
	testResourceAttrs = []attribute.KeyValue{
		attribute.String("service.name", "tester"),
		attribute.String("property", "value"),
	}
)

type clientTestSuite struct {
	suite.Suite

	addr   string
	recv   receiver.Traces
	sink   *consumertest.TracesSink
	sdk    *sdktrace.TracerProvider
	before time.Time
	after  time.Time
}

type timedPoint interface {
	StartTimestamp() pcommon.Timestamp
	EndTimestamp() pcommon.Timestamp
	SetStartTimestamp(pcommon.Timestamp)
	SetEndTimestamp(pcommon.Timestamp)
}

func TestExporterSuite(t *testing.T) {
	suite.Run(t, new(clientTestSuite))
}

func (t *clientTestSuite) SetupTest() {
	ctx := context.Background()

	t.sink.Reset()
	t.before = timeNow()

	exp, err := NewExporter(
		ctx,
		NewConfig(
			WithInsecure(),
			WithEndpoint(t.addr),
			WithHeaders(map[string]string{"lightstep-access-token": "${TOKEN}"}),
		),
	)
	t.NoError(err)

	t.sdk = sdktrace.NewTracerProvider(
		sdktrace.WithResource(
			resource.NewSchemaless(testResourceAttrs...),
		),
		sdktrace.WithBatcher(exp),
	)
}

func (t *clientTestSuite) SetupSuite() {
	ctx := context.Background()

	listener, err := net.Listen("tcp", "127.0.0.1:")
	t.NoError(err)
	t.addr = listener.Addr().String()

	t.NoError(listener.Close())

	factory := otelarrowreceiver.NewFactory()
	cfg := factory.CreateDefaultConfig().(*otelarrowreceiver.Config)
	cfg.Protocols.Arrow = otelarrowreceiver.ArrowConfig{}
	cfg.GRPC.NetAddr = confignet.AddrConfig{Endpoint: t.addr, Transport: "tcp"}

	set := receivertest.NewNopSettings()
	tc := &consumertest.TracesSink{}

	mr, err := factory.CreateTraces(ctx, set, cfg, tc)
	t.NoError(err)

	err = mr.Start(ctx, componenttest.NewNopHost())
	t.NoError(err)

	t.recv = mr
	t.sink = tc
}

func (t *clientTestSuite) checkSpanTimestamps(p timedPoint) {
	if p.StartTimestamp() != 0 {
		t.LessOrEqual(t.before, p.StartTimestamp().AsTime())
	}

	t.LessOrEqual(t.before, t.after)
	t.LessOrEqual(p.EndTimestamp().AsTime(), t.after)

	t.GreaterOrEqual(t.after, p.EndTimestamp().AsTime())
	t.LessOrEqual(t.before, p.EndTimestamp().AsTime())

	// Set these fields to zero, making them omitted fields in JSON,
	// which allows cmp.Diff() uses following this call to work.
	p.SetStartTimestamp(0)
	p.SetEndTimestamp(0)
}

func (t *clientTestSuite) assertTimestamps() {
	t.after = timeNow()
	for _, export := range t.sink.AllTraces() {
		for ri := 0; ri < export.ResourceSpans().Len(); ri++ {
			rs := export.ResourceSpans().At(ri)
			for si := 0; si < rs.ScopeSpans().Len(); si++ {
				ss := rs.ScopeSpans().At(si)
				for mi := 0; mi < ss.Spans().Len(); mi++ {
					s := ss.Spans().At(mi)
					t.checkSpanTimestamps(s)
				}
			}
		}
	}
}

func timeNow() time.Time {
	return time.Now()
}

func (t *clientTestSuite) TestSpan() {
	ctx := context.Background()

	tracer := t.sdk.Tracer("test-tracer")
	_, span := tracer.Start(ctx, "ExecuteRequest", trace.WithSpanKind(trace.SpanKindServer))
	span.SetAttributes(attribute.String("test-attribute-1", "test-value-1"))
	span.AddEvent("test event", trace.WithAttributes(attribute.String("test-event-attribute-1", "test-event-value-1")))
	span.SetStatus(codes.Error, "failed")
	span.End()

	_ = t.sdk.Shutdown(ctx)

	t.Equal(1, len(t.sink.AllTraces()))

	t.assertTimestamps()

	data, err := ptraceotlp.NewExportRequestFromTraces(t.sink.AllTraces()[0]).MarshalProto()
	t.NoError(err)

	expectedSpanID, err := span.SpanContext().SpanID().MarshalJSON()
	t.NoError(err)
	expectedTraceID, err := span.SpanContext().TraceID().MarshalJSON()
	t.NoError(err)
	// trim quotes
	unqSpanID, _ := strconv.Unquote(string(expectedSpanID))
	unqTraceID, _ := strconv.Unquote(string(expectedTraceID))
	roSpan := span.(sdktrace.ReadOnlySpan)
	expect := coltracepb.ExportTraceServiceRequest{
		ResourceSpans: []*tracev1.ResourceSpans{
			{
				Resource: &resourcev1.Resource{
					Attributes: []*commonpb.KeyValue{
						&commonpb.KeyValue{
							Key: "property",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "value",
								},
							},
						},
						&commonpb.KeyValue{
							Key: "service.name",
							Value: &commonpb.AnyValue{
								Value: &commonpb.AnyValue_StringValue{
									StringValue: "tester",
								},
							},
						},
					},
				},
				ScopeSpans: []*tracev1.ScopeSpans{
					&tracev1.ScopeSpans{
						Scope: &commonpb.InstrumentationScope{
							Name: "test-tracer",
						},
						Spans: []*tracev1.Span{
							&tracev1.Span{
								SpanId:  []byte(unqSpanID),
								TraceId: []byte(unqTraceID),
								Kind:    tracev1.Span_SPAN_KIND_SERVER,
								Name:    "ExecuteRequest",
								Status: &tracev1.Status{
									Code:    tracev1.Status_STATUS_CODE_ERROR,
									Message: "failed",
								},
								Attributes: []*commonpb.KeyValue{
									&commonpb.KeyValue{
										Key: "test-attribute-1",
										Value: &commonpb.AnyValue{
											Value: &commonpb.AnyValue_StringValue{
												StringValue: "test-value-1",
											},
										},
									},
								},
								Events: []*tracev1.Span_Event{
									{
										TimeUnixNano: uint64(roSpan.Events()[0].Time.UnixNano()),
										Name:         "test event",
										Attributes: []*commonpb.KeyValue{
											{
												Key: "test-event-attribute-1",
												Value: &commonpb.AnyValue{
													Value: &commonpb.AnyValue_StringValue{
														StringValue: "test-event-value-1",
													},
												},
											},
										},
										DroppedAttributesCount: uint32(roSpan.DroppedAttributes()),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	var export coltracepb.ExportTraceServiceRequest
	t.NoError(proto.Unmarshal(data, &export))

	// For some reason SpanId gets marshaled into a string with hex escape sequence
	// e.g. \xaf\xa3\x0e\xcdR\xe7\xe2\x1e instead of afa30ecd52e7e21e
	exportSpan := export.ResourceSpans[0].ScopeSpans[0].Spans[0]
	exportSpan.SpanId = []byte(hex.EncodeToString(exportSpan.SpanId))
	exportSpan.TraceId = []byte(hex.EncodeToString(exportSpan.TraceId))

	t.Empty(cmp.Diff(expect.String(), export.String()))
}

func (t *clientTestSuite) TestD2PD() {
	ctx := context.Background()

	tracer := t.sdk.Tracer("test-tracer")
	_, span := tracer.Start(ctx, "ExecuteRequest", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(attribute.String("test-attribute-1", "test-value-1"))
	span.AddEvent("test event", trace.WithAttributes(attribute.String("test-event-attribute-1", "test-event-value-1")))
	span.SetStatus(codes.Ok, "this is suppressed")

	_, child := tracer.Start(ctx, "child", trace.WithSpanKind(trace.SpanKindInternal))
	child.SetAttributes(attribute.String("test-attribute-2", "test-value-2"))
	child.AddEvent("child test event", trace.WithAttributes(attribute.String("test-child-event-attribute-2", "test-child-event-value-2")))
	child.End()

	span.End()

	t.NoError(t.sdk.Shutdown(ctx))

	t.Equal(1, len(t.sink.AllTraces()))

	t.assertTimestamps()

	c := client{}
	ptraceObj := c.d2pd([]sdktrace.ReadOnlySpan{
		span.(sdktrace.ReadOnlySpan),
		child.(sdktrace.ReadOnlySpan),
	})

	parentSpan := ptraceObj.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0)
	t.Equal(uint32(0), parentSpan.DroppedAttributesCount())
	t.Equal(uint32(0), parentSpan.DroppedLinksCount())
	t.Equal(uint32(0), parentSpan.DroppedEventsCount())
	t.Equal(ptrace.SpanKindClient, parentSpan.Kind())
	t.Equal("", parentSpan.TraceState().AsRaw())
	t.Equal(ptrace.StatusCodeOk, parentSpan.Status().Code())
	t.Equal("", parentSpan.Status().Message()) // status is suppressed by the SDK

	t.Equal(1, parentSpan.Attributes().Len())

	sattr, _ := parentSpan.Attributes().Get("test-attribute-1")
	t.Equal("test-value-1", sattr.Str())

	t.Equal(1, parentSpan.Events().Len())

	event := span.(sdktrace.ReadOnlySpan).Events()[0]
	parentEvent := parentSpan.Events().At(0)
	t.Equal(event.Time.UnixNano(), parentEvent.Timestamp().AsTime().UnixNano())
	t.Equal(event.Name, parentEvent.Name())
	t.Equal(uint32(0), parentEvent.DroppedAttributesCount())

	for _, attr := range event.Attributes {
		parentEventVal, ok := parentEvent.Attributes().Get(string(attr.Key))
		t.True(ok)
		t.Equal(attr.Value.AsString(), parentEventVal.AsString())
	}

	childSpan := ptraceObj.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(1)
	t.Equal(ptrace.SpanKindInternal, childSpan.Kind())
	t.Equal(ptrace.StatusCodeUnset, childSpan.Status().Code())

	t.Equal(1, childSpan.Attributes().Len())
	sattr, _ = childSpan.Attributes().Get("test-attribute-2")
	t.Equal("test-value-2", sattr.Str())
}
