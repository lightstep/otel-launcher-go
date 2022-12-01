package pipelines

import (
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/sdk/trace"
	"os"
	"strconv"
	"testing"
)

func Test_newSampler(t *testing.T) {
	t.Run("Should always sample when no env is configured", func(t *testing.T) {
		assert.Equal(t, trace.AlwaysSample(), newSampler())
	})

	t.Run("Should never sample when configured with 0%", func(t *testing.T) {
		t.Cleanup(setSamplerEnvs(true, 0))
		assert.Equal(t, trace.NeverSample(), newSampler())
	})

	t.Run("Should sample 75%", func(t *testing.T) {
		t.Cleanup(setSamplerEnvs(true, 75))
		sampler := newSampler()
		assert.Equal(t, "ParentBased{root:TraceIDRatioBased{0.75}", sampler.Description()[:40])
	})
}

func setSamplerEnvs(enabled bool, percent int) func() {
	_ = os.Setenv("LS_SPAN_SAMPLING_ENABLED", strconv.FormatBool(enabled))
	_ = os.Setenv("LS_SPAN_SAMPLING_PERCENT", strconv.Itoa(percent))

	return func() {
		_ = os.Unsetenv("LS_SPAN_SAMPLING_ENABLED")
		_ = os.Unsetenv("LS_SPAN_SAMPLING_PERCENT")
	}
}
