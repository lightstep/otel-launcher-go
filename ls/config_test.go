package ls

import (
	"fmt"
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
	if expected != logger.output[0] {
		t.Errorf("\nExpected: %v\ngot: %v", expected, logger.output[0])
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
	if expected != logger.output[0] {
		t.Errorf("\nExpected: %v\ngot: %v", expected, logger.output[0])
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
