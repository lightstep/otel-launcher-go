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

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/lightstep/otel-launcher-go/pipelines/test"
	"github.com/stretchr/testify/assert"
)

func TestMinimumConfiguration(t *testing.T) {
	server := test.NewServer()
	defer server.Stop()

	os.Setenv("LS_SERVICE_NAME", "test-service-name")
	os.Setenv("OTEL_EXPORTER_OTLP_SPAN_ENDPOINT", fmt.Sprint("0.0.0.0:", server.InsecureTracePort))
	os.Setenv("OTEL_EXPORTER_OTLP_METRIC_ENDPOINT", fmt.Sprint("0.0.0.0:", server.InsecureMetricsPort))
	os.Setenv("OTEL_EXPORTER_OTLP_SPAN_INSECURE", "true")
	os.Setenv("OTEL_EXPORTER_OTLP_METRIC_INSECURE", "true")
	os.Setenv("NOHANG", "true")
	assert.NotPanics(t, main)
}
