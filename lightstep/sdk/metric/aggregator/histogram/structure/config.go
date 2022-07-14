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

package structure // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram/structure"

// DefaultMaxSize is the default maximum number of buckets per
// positive or negative number range.  The value 160 is specified by
// OpenTelemetry--yields a maximum relative error of less than 5% for
// data with contrast 10**5 (e.g., latencies in the range 1ms to 100s).
// See the derivation here:
// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/metrics/sdk.md#exponential-histogram-aggregation
const DefaultMaxSize int32 = 160

// MinSize is the smallest reasonable configuration, which is small
// enough to contain the entire normal floating point range at
// MinScale.
const MinSize = 2

// MaximumMaxSize is an arbitrary limit.
const MaximumMaxSize = 16384

// Config contains configuration for exponential histogram creation.
type Config struct {
	maxSize int32
}

// Option is the interface that applies a configuration option.
type Option interface {
	// apply sets the Option value of a config.
	apply(Config) Config
}

// WithMaxSize sets
func WithMaxSize(size int32) Option {
	return maxSize(size)
}

// maxSize is an option to set the maximum histogram size.
type maxSize int32

// apply implements Option.
func (ms maxSize) apply(cfg Config) Config {
	cfg.maxSize = int32(ms)
	return cfg
}

// NewConfig returns an exponential histogram configuration with
// defaults and limits applied.
func NewConfig(opts ...Option) Config {
	var cfg Config
	for _, opt := range opts {
		cfg = opt.apply(cfg)
	}
	if cfg.maxSize == 0 {
		cfg.maxSize = DefaultMaxSize
	}
	if cfg.maxSize < MinSize {
		cfg.maxSize = MinSize
	}
	if cfg.maxSize > MaximumMaxSize {
		cfg.maxSize = MaximumMaxSize
	}
	return cfg
}

// Valid returns true for valid configurations.
func (c Config) Valid() bool {
	if c.maxSize == 0 {
		return true
	}
	return c.maxSize >= MinSize && c.maxSize <= MaximumMaxSize
}
