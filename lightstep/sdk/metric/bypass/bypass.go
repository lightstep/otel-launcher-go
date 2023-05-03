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

package bypass // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/bypass"

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
)

// FastInt64Adder is implemented by int64 Counter and UpDownCounter
// instruments retuned by this SDK and offers a fast-path for updating
// metrics with a list of attributes.
type FastInt64Adder interface {
	AddWithKeyValues(ctx context.Context, value int64, attrs ...attribute.KeyValue)
}

// FastFloat64Adder is implemented by float64 Counter and
// UpDownCounter instruments retuned by this SDK and offers a
// fast-path for updating metrics with a list of attributes.
type FastFloat64Adder interface {
	AddWithKeyValues(ctx context.Context, value float64, attrs ...attribute.KeyValue)
}

// FastInt64Recorder is implemented by float64 Histogram instruments
// retuned by this SDK and offers a fast-path for updating metrics
// with a list of attributes.
type FastInt64Recorder interface {
	RecordWithKeyValues(ctx context.Context, value int64, attrs ...attribute.KeyValue)
}

// FastFloat64Recorder is implemented by float64 Histogram instruments
// retuned by this SDK and offers a fast-path for updating metrics
// with a list of attributes.
type FastFloat64Recorder interface {
	RecordWithKeyValues(ctx context.Context, value float64, attrs ...attribute.KeyValue)
}
