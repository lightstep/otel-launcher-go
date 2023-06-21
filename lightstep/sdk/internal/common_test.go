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

package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

func TestResourceMapEqual(t *testing.T) {
	var rm ResourceMap

	r1 := rm.Get(resource.NewSchemaless(attribute.String("K", "V")))
	r2 := rm.Get(resource.NewSchemaless(attribute.String("K", "V")))

	require.Equal(t, r1, r2)
	require.False(t, rm.disabled)
}

func TestResourceMapUnequal(t *testing.T) {
	var rm ResourceMap

	r1 := rm.Get(resource.NewSchemaless(attribute.String("K", "V")))
	r2 := rm.Get(resource.NewSchemaless(attribute.String("K", "W")))

	require.NotEqual(t, r1, r2)
	require.True(t, rm.disabled)
	require.Nil(t, rm.input)

	r3 := rm.Get(resource.NewSchemaless(attribute.String("K", "V")))
	require.Equal(t, r1, r3)
}

func BenchmarkResourceUnequal(b *testing.B) {
	b.ReportAllocs()
	var rm ResourceMap

	_ = rm.Get(resource.NewSchemaless(attribute.String("K", "V")))
	repeat := resource.NewSchemaless(attribute.Int("I", 0))

	for i := 0; i < b.N; i++ {
		_ = rm.Get(repeat)
	}
}

func BenchmarkResourceEqual(b *testing.B) {
	var rm ResourceMap

	_ = rm.Get(resource.NewSchemaless(attribute.String("K", "V")))
	repeat := resource.NewSchemaless(attribute.String("K", "V"))

	for i := 0; i < b.N; i++ {
		_ = rm.Get(repeat)
	}
}
