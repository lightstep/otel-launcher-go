package pipelines

import (
	"context"
	"github.com/sethvargo/go-envconfig"
	"go.opentelemetry.io/otel/sdk/trace"
	"log"
)

type samplingConfig struct {
	// SamplingEnabled turns on span sampling. This should be set alongside the
	// SamplingPercent attribute. If sampling is disabled, all traces are sent
	// to the endpoint.
	SpanSamplingEnabled bool `env:"LS_SPAN_SAMPLING_ENABLED,default=false"`
	// SamplingPercent is the percentage of spans will be sent to the endpoint,
	// in the range 0-100. It is only consulted if SamplingEnabled is set to true
	SpanSamplingPercent int `env:"LS_SPAN_SAMPLING_PERCENT,default=100"`
}

func newSampler() trace.Sampler {

	var c samplingConfig

	if err := envconfig.Process(context.Background(), &c); err != nil {
		log.Printf("Could not load sampling config, default to AlwaysSample(): %s ", err)
		return trace.AlwaysSample()
	}

	if !c.SpanSamplingEnabled || c.SpanSamplingPercent == 100 {
		return trace.AlwaysSample()
	}

	if c.SpanSamplingPercent == 0 {
		return trace.NeverSample()
	}

	return trace.ParentBased(
		trace.TraceIDRatioBased(float64(c.SpanSamplingPercent) / 100.0),
	)
}
