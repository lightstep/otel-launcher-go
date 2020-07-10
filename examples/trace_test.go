package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinimumConfiguration(t *testing.T) {
	os.Setenv("LS_SERVICE_NAME", "test-service-name")
	os.Setenv("OTEL_EXPORTER_OTLP_SPAN_ENDPOINT", "invalid-endpoint")
	assert.NotPanics(t, main)
}
