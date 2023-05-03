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

package metric // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric"

import (
	"context"
	"testing"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/bypass"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/sdkinstrument"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var unsafePerf = WithPerformance(sdkinstrument.Performance{
	IgnoreCollisions: true,
})

// Tested prior to 0.17.0 release

func BenchmarkCounterAddNoAttrs(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1)
	}
}

func BenchmarkCounterAddOneAttrSafe(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, metric.WithAttributes(attribute.String("K", "V")))
	}
}

func BenchmarkCounterAddOneAttrSafeBypass(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.(bypass.FastInt64Adder).AddWithKeyValues(ctx, 1, attribute.String("K", "V"))
	}
}

func BenchmarkCounterAddOneAttrUnsafe(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr), unsafePerf)
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, metric.WithAttributes(attribute.String("K", "V")))
	}
}

func BenchmarkCounterAddOneAttrUnsafeBypass(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr), unsafePerf)
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.(bypass.FastInt64Adder).AddWithKeyValues(ctx, 1, attribute.String("K", "V"))
	}
}

func BenchmarkCounterAddOneAttrSliceReuseSafe(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	attrs := []attribute.KeyValue{
		attribute.String("K", "V"),
	}
	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

func BenchmarkCounterAddOneAttrSliceReuseSafeBypass(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	attrs := []attribute.KeyValue{
		attribute.String("K", "V"),
	}
	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.(bypass.FastInt64Adder).AddWithKeyValues(ctx, 1, attrs...)
	}
}

func BenchmarkCounterAddOneAttrSliceReuseUnsafe(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr), unsafePerf)
	b.ReportAllocs()

	attrs := []attribute.KeyValue{
		attribute.String("K", "V"),
	}
	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
}

func BenchmarkCounterAddOneAttrSliceReuseUnsafeBypass(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr), unsafePerf)
	b.ReportAllocs()

	attrs := []attribute.KeyValue{
		attribute.String("K", "V"),
	}
	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.(bypass.FastInt64Adder).AddWithKeyValues(ctx, 1, attrs...)
	}
}

func BenchmarkCounterAddOneInvalidAttr(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, metric.WithAttributes(attribute.String("", "V"), attribute.String("K", "V")))
	}
}

func BenchmarkCounterAddManyAttrs(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, metric.WithAttributes(attribute.Int("K", i)))
	}
}

func BenchmarkCounterAddManyAttrsUnsafe(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr), unsafePerf)
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, metric.WithAttributes(attribute.Int("K", i)))
	}
}

func BenchmarkCounterAddManyInvalidAttrs(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, metric.WithAttributes(attribute.Int("", i), attribute.Int("K", i)))
	}
}

func BenchmarkCounterAddManyInvalidAttrsUnsafe(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr), unsafePerf)
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, metric.WithAttributes(attribute.Int("", i), attribute.Int("K", i)))
	}
}

func BenchmarkCounterAddManyFilteredAttrs(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(
		WithReader(rdr,
			view.WithClause(view.WithKeys([]attribute.Key{"K"})),
		),
	)
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, metric.WithAttributes(attribute.Int("L", i), attribute.Int("K", i)))
	}
}

func BenchmarkCounterCollectOneAttrNoReuse(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, metric.WithAttributes(attribute.Int("K", 1)))

		_ = rdr.Produce(nil)
	}
}

func BenchmarkCounterCollectOneAttrWithReuse(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	var reuse data.Metrics

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, metric.WithAttributes(attribute.Int("K", 1)))

		reuse = rdr.Produce(&reuse)
	}
}

func BenchmarkCounterCollectTenAttrs(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	var reuse data.Metrics

	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			cntr.Add(ctx, 1, metric.WithAttributes(attribute.Int("K", j)))
		}
		reuse = rdr.Produce(&reuse)
	}
}

func BenchmarkCounterCollectTenAttrsTenTimes(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").Int64Counter("hello")

	var reuse data.Metrics

	for i := 0; i < b.N; i++ {
		for k := 0; k < 10; k++ {
			for j := 0; j < 10; j++ {
				cntr.Add(ctx, 1, metric.WithAttributes(attribute.Int("K", j)))
			}
			reuse = rdr.Produce(&reuse)
		}
	}
}
