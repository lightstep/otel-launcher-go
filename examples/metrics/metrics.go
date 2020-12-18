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
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/metric"
)

func main() {
	defer launcher.ConfigureOpentelemetry().Shutdown()

	name := os.Getenv("LS_SERVICE_NAME")
	prefix := fmt.Sprint("otel.", name, ".")

	ctx := context.Background()
	meter := metric.Must(otel.Meter("testing"))

	// This example shows 1 instrument each for (UpDown)?(SumObserver|Counter)
	// The two monotonic instruments of these have a fixed rate of +1.
	// The two non-monotonic instruments of these have a fixed rate of -1.
	// There are 2 example ValueObserver instruments (one a Gaussian, one a Sine wave).
	//
	// TODO: ValueRecorder examples.

	c1 := meter.NewInt64Counter(prefix + "counter")
	c2 := meter.NewInt64UpDownCounter(prefix + "updowncounter")
	go func() {
		for {
			c1.Add(ctx, 1)
			c2.Add(ctx, -1)
			time.Sleep(time.Second)
		}
	}()

	startTime := time.Now()

	meter.NewInt64SumObserver(
		prefix+"sumobserver",
		func(_ context.Context, result metric.Int64ObserverResult) {
			result.Observe(int64(time.Since(startTime).Seconds()))
		},
	)

	meter.NewInt64UpDownSumObserver(
		prefix+"updownsumobserver",
		func(_ context.Context, result metric.Int64ObserverResult) {
			result.Observe(-int64(time.Since(startTime).Seconds()))
		},
	)

	meter.NewInt64ValueObserver(
		prefix+"valueobserver",
		func(_ context.Context, result metric.Int64ObserverResult) {
			result.Observe(int64(50 + rand.NormFloat64()*50))
		},
	)

	meter.NewFloat64ValueObserver(
		prefix+"sine_wave",
		func(_ context.Context, result metric.Float64ObserverResult) {
			secs := float64(time.Now().UnixNano()) / float64(time.Second)

			result.Observe(math.Sin(secs/(200*math.Pi)), label.String("period", "fast"))
			result.Observe(math.Sin(secs/(1000*math.Pi)), label.String("period", "regular"))
			result.Observe(math.Sin(secs/(5000*math.Pi)), label.String("period", "slow"))
		},
	)

	if os.Getenv("NOHANG") == "" {
		select {}
	}
}
