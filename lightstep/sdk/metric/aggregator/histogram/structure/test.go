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

import (
	"testing"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/aggregation"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/stretchr/testify/require"
)

// Test helpers

func RequireEqualValues[N number.Any](t *testing.T, a, b *State[N]) {
	require.Equal(t, a.Scale(), b.Scale())
	require.Equal(t, a.Count(), b.Count())
	require.Equal(t, a.Sum(), b.Sum())
	RequireEqualBuckets(t, a.Positive(), b.Positive())
	RequireEqualBuckets(t, a.Negative(), b.Negative())
}

func RequireEqualBuckets(t *testing.T, a, b aggregation.Buckets) {
	require.Equal(t, a.Len(), b.Len())
	require.Equal(t, a.Offset(), b.Offset())
	for i := uint32(0); i < a.Len(); i++ {
		require.Equal(t, a.At(i), b.At(i))
	}
}
