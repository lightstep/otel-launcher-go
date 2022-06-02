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

// Package main configures OpenTelemetry metrics and starts reporting
// runtime and host metrics.  For example:
//
//   LS_SERVICE_NAME=foo \
//   OTEL_EXPORTER_OTLP_SPAN_ENDPOINT=localhost:9999 \
//   OTEL_EXPORTER_OTLP_METRIC_ENDPOINT=localhost:9999 \
//   go run ./metrics.go

package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/lightstep/otel-launcher-go/launcher"
	"go.opentelemetry.io/otel/attribute"
	metricglobal "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

func main() {
	defer launcher.ConfigureOpentelemetry().Shutdown()

	name := os.Getenv("LS_SERVICE_NAME")
	prefix := fmt.Sprint("otel.", name, ".")

	ctx := context.Background()
	meter := metricglobal.Meter("testing")

	// This example shows 1 instrument each for synchronous and
	// asynchronous Counter and UpDownCounter.  The two monotonic
	// instruments of these have a fixed rate of +1.  The two
	// non-monotonic instruments of these have a fixed rate of -1.
	// There are 2 example Gauge instruments (one a Gaussian, one
	// a Sine wave), and one Histogram.

	c1, _ := meter.SyncInt64().Counter(prefix + "counter")
	c2, _ := meter.SyncInt64().UpDownCounter(prefix + "updowncounter")
	hist, _ := meter.SyncFloat64().Histogram(prefix + "histogram")
	go func() {
		for {
			c1.Add(ctx, 1)
			c2.Add(ctx, -1)

			secs := float64(time.Now().UnixNano()) / float64(time.Second)
			mult := math.Sin(secs / (40 * math.Pi))
			mult *= mult

			for i := 0; i < 10000; i++ {
				hist.Record(ctx, mult*(100+rand.NormFloat64()*100))
			}

			time.Sleep(time.Second)
		}
	}()

	startTime := time.Now()

	counterObserver, _ := meter.AsyncInt64().Counter(
		prefix + "counterobserver",
	)

	err := meter.RegisterCallback([]instrument.Asynchronous{counterObserver},
		func(ctx context.Context) {
			counterObserver.Observe(ctx, int64(time.Since(startTime).Seconds()))
		},
	)

	if err != nil {
		fmt.Printf("%v", err)
	}

	updownCounterObserver, _ := meter.AsyncInt64().UpDownCounter(
		prefix + "updowncounterobserver",
	)

	err = meter.RegisterCallback([]instrument.Asynchronous{updownCounterObserver},
		func(ctx context.Context) {
			updownCounterObserver.Observe(ctx, -int64(time.Since(startTime).Seconds()))
		},
	)

	if err != nil {
		fmt.Printf("%v", err)
	}

	gauge, _ := meter.AsyncInt64().Gauge(
		prefix + "gauge",
	)

	err = meter.RegisterCallback([]instrument.Asynchronous{gauge},
		func(ctx context.Context) {
			gauge.Observe(ctx, int64(50+rand.NormFloat64()*50))
		},
	)

	if err != nil {
		fmt.Printf("%v", err)
	}

	sineWave, _ := meter.AsyncFloat64().Gauge(
		prefix + "sine_wave",
	)

	err = meter.RegisterCallback([]instrument.Asynchronous{sineWave},
		func(ctx context.Context) {
			secs := float64(time.Now().UnixNano()) / float64(time.Second)

			sineWave.Observe(ctx, math.Sin(secs/(50*math.Pi)), attribute.String("period", "fastest"))
			sineWave.Observe(ctx, math.Sin(secs/(200*math.Pi)), attribute.String("period", "fast"))
			sineWave.Observe(ctx, math.Sin(secs/(1000*math.Pi)), attribute.String("period", "regular"))
			sineWave.Observe(ctx, math.Sin(secs/(5000*math.Pi)), attribute.String("period", "slow"))
		},
	)

	if err != nil {
		fmt.Printf("%v", err)
	}

	if os.Getenv("NOHANG") == "" {
		select {}
	}
}
