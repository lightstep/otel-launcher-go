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

package sum // import "github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/sum"

import (
	"testing"

	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/aggregator/test"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/number"
	"github.com/stretchr/testify/require"
)

func TestMonotonicity(t *testing.T) {
	require.True(t, NewMonotonicInt64(1).IsMonotonic())
	require.True(t, NewMonotonicFloat64(2).IsMonotonic())
	require.False(t, NewNonMonotonicInt64(3).IsMonotonic())
	require.False(t, NewNonMonotonicFloat64(4).IsMonotonic())
}

func TestInt64MonotonicSum(t *testing.T) {
	test.GenericAggregatorTest[int64, MonotonicInt64, MonotonicInt64Methods](t, number.ToInt64)
}

func TestFloat64MonotonicSum(t *testing.T) {
	test.GenericAggregatorTest[float64, MonotonicFloat64, MonotonicFloat64Methods](t, number.ToFloat64)
}

func TestInt64NonMonotonicSum(t *testing.T) {
	test.GenericAggregatorTest[int64, NonMonotonicInt64, NonMonotonicInt64Methods](t, number.ToInt64)
}

func TestFloat64NonMonotonicSum(t *testing.T) {
	test.GenericAggregatorTest[float64, NonMonotonicFloat64, NonMonotonicFloat64Methods](t, number.ToFloat64)
}

func genericSubtractTest[N number.Any, Storage any, Methods aggregator.Methods[N, Storage]](t *testing.T) {
	var operand Storage
	var argument Storage
	var expect Storage
	var methods Methods

	methods.Init(&operand, aggregator.Config{})
	methods.Init(&argument, aggregator.Config{})
	methods.Init(&expect, aggregator.Config{})

	methods.Update(&operand, 3)
	methods.Update(&argument, 10)
	methods.Update(&expect, 7)

	methods.SubtractSwap(&operand, &argument) // operand(7) = argument(10) - operand(3)

	require.Equal(t, expect, operand)
}

func TestSubtract(t *testing.T) {
	genericSubtractTest[int64, MonotonicInt64, MonotonicInt64Methods](t)
	genericSubtractTest[float64, MonotonicFloat64, MonotonicFloat64Methods](t)
	genericSubtractTest[int64, NonMonotonicInt64, NonMonotonicInt64Methods](t)
	genericSubtractTest[float64, NonMonotonicFloat64, NonMonotonicFloat64Methods](t)
}
