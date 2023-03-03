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

package pipeline // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/pipeline"

import "go.opentelemetry.io/otel/attribute"

// OverflowAttributes is the specified list of attributes to use when
// configured mechanisms overflow a cardinality limit.
var OverflowAttributes = []attribute.KeyValue{
	attribute.Bool("otel.metric.overflow", true),
}

// OverflowAttributeSet is the set corresponding with OverflowAttributes.
var OverflowAttributeSet = attribute.NewSet(OverflowAttributes...)
