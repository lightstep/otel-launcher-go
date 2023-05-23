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

// Tested prior to 1.17.0 release
// goos: darwin
// goarch: arm64
// pkg: github.com/lightstep/otel-launcher-go/lightstep/sdk/metric
// BenchmarkCounterAddNoAttrs-10                    	32278780	        37.43 ns/op	       0 B/op	       0 allocs/op
// BenchmarkCounterAddOneAttrSafe-10                	14648950	        81.16 ns/op	      64 B/op	       1 allocs/op
// BenchmarkCounterAddOneAttrUnsafe-10              	16286137	        73.14 ns/op	      64 B/op	       1 allocs/op
// BenchmarkCounterAddOneAttrSliceReuseSafe-10      	22135064	        54.10 ns/op	       0 B/op	       0 allocs/op
// BenchmarkCounterAddOneAttrSliceReuseUnsafe-10    	25858933	        46.81 ns/op	       0 B/op	       0 allocs/op
// 2023/05/03 17:04:01 use of empty attribute key, e.g., metric name "hello" with value "V"
// BenchmarkCounterAddOneInvalidAttr-10             	 9953920	       118.5 ns/op	     128 B/op	       1 allocs/op
// BenchmarkCounterAddManyAttrs-10                  	12176029	        96.97 ns/op	      64 B/op	       1 allocs/op
// BenchmarkCounterAddManyAttrsUnsafe-10            	13241512	        88.42 ns/op	      64 B/op	       1 allocs/op
// BenchmarkCounterAddManyInvalidAttrs-10           	 8614119	       134.7 ns/op	     128 B/op	       1 allocs/op
// BenchmarkCounterAddManyInvalidAttrsUnsafe-10     	 9404649	       124.8 ns/op	     128 B/op	       1 allocs/op
// BenchmarkCounterAddManyFilteredAttrs-10          	 8649364	       137.7 ns/op	     128 B/op	       1 allocs/op
// BenchmarkCounterCollectOneAttrNoReuse-10         	 2633346	       454.0 ns/op	     400 B/op	       7 allocs/op
// BenchmarkCounterCollectOneAttrWithReuse-10       	 3539598	       338.0 ns/op	     136 B/op	       3 allocs/op
// BenchmarkCounterCollectTenAttrs-10               	  718842	      1659 ns/op	     712 B/op	      12 allocs/op
// BenchmarkCounterCollectTenAttrsTenTimes-10       	   71664	     16535 ns/op	    7120 B/op	     120 allocs/op

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
