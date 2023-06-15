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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func main() {
	defer launcher.ConfigureOpentelemetry().Shutdown()

	name := os.Getenv("LS_SERVICE_NAME")
	prefix := fmt.Sprint("otel.", name, ".")

	ctx := context.Background()
	meter := otel.Meter("testing")

	// This example shows 1 instrument each for synchronous and
	// asynchronous Counter and UpDownCounter.  The two monotonic
	// instruments of these have a fixed rate of +1.  The two
	// non-monotonic instruments of these have a fixed rate of -1.
	// There are 2 example Gauge instruments (one a Gaussian, one
	// a Sine wave), and one Histogram.

	c1, _ := meter.Int64Counter(prefix + "counter")
	c2, _ := meter.Int64UpDownCounter(prefix + "updowncounter")
	hist, _ := meter.Float64Histogram(prefix + "histogram")
	mmsc, _ := meter.Float64Histogram(prefix+"mmsc",
		metric.WithDescription(`{
  "aggregation": "minmaxsumcount"
}`),
	)
	go func() {
		for {
			c1.Add(ctx, 1)
			c2.Add(ctx, -1)

			secs := float64(time.Now().UnixNano()) / float64(time.Second)
			mult := math.Sin(secs / (40 * math.Pi))
			mult *= mult

			for i := 0; i < 10000; i++ {
				value := math.Abs(mult * (100 + rand.NormFloat64()*100))
				hist.Record(ctx, value)
				mmsc.Record(ctx, value)
			}

			time.Sleep(time.Second)
		}
	}()

	startTime := time.Now()

	_, err := meter.Int64ObservableCounter(
		prefix+"counterobserver",
		metric.WithInt64Callback(
			func(ctx context.Context, obs metric.Int64Observer) error {
				obs.Observe(int64(time.Since(startTime).Seconds()))
				return nil
			},
		),
	)

	if err != nil {
		fmt.Printf("%v", err)
	}

	_, err = meter.Int64ObservableUpDownCounter(
		prefix+"updowncounterobserver",
		metric.WithInt64Callback(
			func(ctx context.Context, obs metric.Int64Observer) error {
				obs.Observe(-int64(time.Since(startTime).Seconds()))
				return nil
			},
		),
	)

	if err != nil {
		fmt.Printf("%v", err)
	}

	_, err = meter.Int64ObservableGauge(
		prefix+"gauge",
		metric.WithInt64Callback(
			func(ctx context.Context, obs metric.Int64Observer) error {
				obs.Observe(int64(50 + rand.NormFloat64()*50))
				return nil
			},
		),
	)

	if err != nil {
		fmt.Printf("%v", err)
	}

	_, err = meter.Float64ObservableGauge(
		prefix+"sine_wave",
		metric.WithFloat64Callback(
			func(ctx context.Context, obs metric.Float64Observer) error {
				secs := float64(time.Now().UnixNano()) / float64(time.Second)

				obs.Observe(math.Sin(secs/(50*math.Pi)), metric.WithAttributes(attribute.String("period", "fastest")))
				obs.Observe(math.Sin(secs/(200*math.Pi)), metric.WithAttributes(attribute.String("period", "fast")))
				obs.Observe(math.Sin(secs/(1000*math.Pi)), metric.WithAttributes(attribute.String("period", "regular")))
				obs.Observe(math.Sin(secs/(5000*math.Pi)), metric.WithAttributes(attribute.String("period", "slow")))
				return nil
			},
		),
	)

	if err != nil {
		fmt.Printf("%v", err)
	}

	if os.Getenv("NOHANG") == "" {
		select {}
	}
}
