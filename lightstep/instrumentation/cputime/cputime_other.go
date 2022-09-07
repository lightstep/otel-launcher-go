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

//go:build !(linux || darwin)

package cputime // import "github.com/lightstep/otel-launcher-go/lightstep/instrumentation/cputime"

import (
	"context"
	"fmt"
	"math"
	"os"
	"runtime"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/shirou/gopsutil/v3/process"
)

// cputime reports the work-in-progress conventional cputime metrics specified by OpenTelemetry.
type cputime struct {
	meter    metric.Meter
	selfProc *process.Process
}

func newCputime(c config) (*cputime, error) {
	selfProc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, err
	}
	return &cputime{
		meter:    c.MeterProvider.Meter("otel_launcher_go/cputime"),
		selfProc: selfProc,
	}, nil
}

// getProcessTimes calls ReadMemStats() for GCCPUFraction because as
// of Go-1.19 there is no such runtime metric.  User and system sum to
// 100% of CPU time; gc is an independent, comparable metric value.
// These are correlated with uptime.
func (c *cputime) getProcessTimes(ctx context.Context) (userSeconds, systemSeconds, gcSeconds, uptimeSeconds float64) {
	// Would really be better if runtime/metrics exposed this,
	// making an expensive call for a single field that is not
	// exposed via ReadMemStats().
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	gomaxprocs := float64(runtime.GOMAXPROCS(0))

	uptimeSeconds = time.Since(processStartTime).Seconds()
	gcSeconds = memStats.GCCPUFraction * uptimeSeconds * gomaxprocs

	processTimes, err := c.selfProc.TimesWithContext(ctx)
	if err != nil {
		userSeconds = math.NaN()
		systemSeconds = math.NaN()
		otel.Handle(fmt.Errorf("could not find this process: %w", err))
		return
	}

	userSeconds = processTimes.User
	systemSeconds = processTimes.System
	return
}
