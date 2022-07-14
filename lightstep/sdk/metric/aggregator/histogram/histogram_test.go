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

package histogram // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram"

import (
	"testing"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/histogram/structure"
	"github.com/stretchr/testify/require"
)

type Float64 = structure.Float64
type Int64 = structure.Int64

var NewFloat64 = structure.NewFloat64
var NewInt64 = structure.NewInt64
var NewConfig = structure.NewConfig

func TestAggregatorCopyMove(t *testing.T) {
	var mf Float64Methods

	h1 := NewFloat64(NewConfig(), 1, 3, 5, 7, 9)
	h2 := NewFloat64(NewConfig())
	h3 := NewFloat64(NewConfig())

	mf.Move(h1, h2)
	mf.Copy(h2, h3)

	structure.RequireEqualValues(t, h2, h3)
}

func TestAggregatorToFrom(t *testing.T) {
	var mi Int64Methods
	var mf Float64Methods
	var hi Int64

	hs, ok := mi.ToStorage(mi.ToAggregation(&hi))
	require.Equal(t, &hi, hs)
	require.True(t, ok)

	_, ok = mf.ToStorage(mi.ToAggregation(&hi))
	require.False(t, ok)
}
