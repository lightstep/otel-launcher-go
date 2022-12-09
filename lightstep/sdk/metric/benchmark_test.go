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

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view"
	"go.opentelemetry.io/otel/attribute"
)

// Tested prior to 0.11.0 release
// goos: darwin
// goarch: arm64
// pkg: github.com/lightstep/otel-launcher-go/lightstep/sdk/metric
// BenchmarkCounterAddNoAttrs-10                 	35354023	        33.79 ns/op	       0 B/op	       0 allocs/op
// BenchmarkCounterAddOneAttr-10                 	14354538	        82.77 ns/op	      64 B/op	       1 allocs/op
// BenchmarkCounterAddOneInvalidAttr-10          	 9307794	       128.4 ns/op	     128 B/op	       1 allocs/op
// BenchmarkCounterAddManyAttrs-10               	 1000000	      1075 ns/op	     569 B/op	       6 allocs/op
// BenchmarkCounterAddManyInvalidAttrs-10        	  832549	      1654 ns/op	    1080 B/op	      10 allocs/op
// BenchmarkCounterAddManyFilteredAttrs-10       	 1000000	      1304 ns/op	     953 B/op	       8 allocs/op
// BenchmarkCounterCollectOneAttrNoReuse-10      	 2537348	       468.0 ns/op	     400 B/op	       7 allocs/op
// BenchmarkCounterCollectOneAttrWithReuse-10    	 3679694	       328.2 ns/op	     136 B/op	       3 allocs/op
// BenchmarkCounterCollectTenAttrs-10            	  715490	      1635 ns/op	     712 B/op	      12 allocs/op
// BenchmarkCounterCollectTenAttrsTenTimes-10    	   72478	     16475 ns/op	    7128 B/op	     120 allocs/op

func BenchmarkCounterAddNoAttrs(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1)
	}
}

func BenchmarkCounterAddOneAttr(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, attribute.String("K", "V"))
	}
}

func BenchmarkCounterAddOneInvalidAttr(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, attribute.String("", "V"), attribute.String("K", "V"))
	}
}

func BenchmarkCounterAddManyAttrs(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, attribute.Int("K", i))
	}
}

func BenchmarkCounterAddManyInvalidAttrs(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, attribute.Int("", i), attribute.Int("K", i))
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

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, attribute.Int("L", i), attribute.Int("K", i))
	}
}

func BenchmarkCounterCollectOneAttrNoReuse(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, attribute.Int("K", 1))

		_ = rdr.Produce(nil)
	}
}

func BenchmarkCounterCollectOneAttrWithReuse(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	var reuse data.Metrics

	for i := 0; i < b.N; i++ {
		cntr.Add(ctx, 1, attribute.Int("K", 1))

		reuse = rdr.Produce(&reuse)
	}
}

func BenchmarkCounterCollectTenAttrs(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	var reuse data.Metrics

	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			cntr.Add(ctx, 1, attribute.Int("K", j))
		}
		reuse = rdr.Produce(&reuse)
	}
}

func BenchmarkCounterCollectTenAttrsTenTimes(b *testing.B) {
	ctx := context.Background()
	rdr := NewManualReader("bench")
	provider := NewMeterProvider(WithReader(rdr))
	b.ReportAllocs()

	cntr, _ := provider.Meter("test").SyncInt64().Counter("hello")

	var reuse data.Metrics

	for i := 0; i < b.N; i++ {
		for k := 0; k < 10; k++ {
			for j := 0; j < 10; j++ {
				cntr.Add(ctx, 1, attribute.Int("K", j))
			}
			reuse = rdr.Produce(&reuse)
		}
	}
}
