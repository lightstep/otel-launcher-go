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

package metrictransform // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/exporters/otlp/internal/metrictransform"

import (
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

// Resource transforms a Resource into an OTLP Resource.
func Resource(r *resource.Resource) *resourcepb.Resource {
	if r == nil {
		return nil
	}
	return &resourcepb.Resource{Attributes: ResourceAttributes(r)}
}

// Library transforms a Library descriptor into an OTLP InstrumentationScope.
func Library(lib instrumentation.Library) *commonpb.InstrumentationScope {
	return &commonpb.InstrumentationScope{
		Name:    lib.Name,
		Version: lib.Version,
	}
}

// Temporality transforms a Temporality value into an OTLP AggrgationTemporality.
func Temporality(tempo aggregation.Temporality) metricspb.AggregationTemporality {
	switch tempo {
	case aggregation.CumulativeTemporality:
		return metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE
	case aggregation.DeltaTemporality:
		return metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA
	default:
		return metricspb.AggregationTemporality_AGGREGATION_TEMPORALITY_UNSPECIFIED
	}
}
