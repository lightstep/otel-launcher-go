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

package ls

import (
	"fmt"
	"strings"
	"testing"
)

type testLogger struct {
	output []string
}

func (t *testLogger) addOutput(output string) {
	t.output = append(t.output, output)
}

func (t *testLogger) Fatalf(format string, v ...interface{}) {
	t.addOutput(fmt.Sprintf(format, v...))
}

func (t *testLogger) Debugf(format string, v ...interface{}) {
	t.addOutput(fmt.Sprintf(format, v...))
}

func TestInvalidServiceName(t *testing.T) {
	logger := testLogger{output: []string{}}
	lsOtel := ConfigureOpentelemetry(WithLogger(&logger))
	defer lsOtel.Shutdown()

	expected := "invalid configuration: service name missing"
	if !strings.Contains(logger.output[0], expected) {
		t.Errorf("\nString not found: %v\nIn: %v", expected, logger.output[0])
	}
}

func TestInvalidAccessToken(t *testing.T) {
	logger := testLogger{output: []string{}}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(&logger),
		WithServiceName("test-service"),
	)
	defer lsOtel.Shutdown()

	expected := "invalid configuration: access token missing, must be set when reporting to ingest.lightstep.com:443"
	if !strings.Contains(logger.output[0], expected) {
		t.Errorf("\nString not found: %v\nIn: %v", expected, logger.output[0])
	}
}

func TestValidConfig(t *testing.T) {
	logger := testLogger{output: []string{}}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(&logger),
		WithServiceName("test-service"),
		WithAccessToken("access-token-123"),
		WithSatelliteURL("localhost:443"),
	)
	defer lsOtel.Shutdown()
	expected := 0
	if len(logger.output) > expected {
		t.Errorf("\nExpected: %v\ngot: %v", expected, len(logger.output))
	}
}

func TestDebugEnabled(t *testing.T) {
	logger := testLogger{output: []string{}}
	lsOtel := ConfigureOpentelemetry(
		WithLogger(&logger),
		WithServiceName("test-service"),
		WithAccessToken("access-token-123"),
		WithSatelliteURL("localhost:443"),
		WithDebug(true),
	)
	defer lsOtel.Shutdown()
	expected := "debug logging enabled"
	if expected != logger.output[0] {
		t.Errorf("\nExpected: %v\ngot: %v", expected, logger.output[0])
	}
}

func TestDefaultConfig(t *testing.T) {
	logger := testLogger{}
	config := newConfig(WithLogger(&logger))

	expected := LightstepConfig{
		ServiceName:    "",
		ServiceVersion: "unknown",
		SatelliteURL:   "ingest.lightstep.com:443",
		MetricsURL:     "ingest.lightstep.com:443/metrics",
		AccessToken:    "",
		Debug:          false,
		Insecure:       false,
		logger:         &logger,
	}
	if config != expected {
		t.Errorf("\nExpected: %v\ngot: %v", expected, config)
	}
}

func TestEnvironmentVariables(t *testing.T) {

}
