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

package sdkinstrument

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestTruncateAttrs(t *testing.T) {
	const Limit = 16

	type test struct {
		Name   string
		Attrs  []attribute.KeyValue
		Expect []attribute.KeyValue
	}

	mkAttrs := func(attrs ...attribute.KeyValue) []attribute.KeyValue {
		return attrs
	}

	for _, test := range []test{
		{
			Name:   "simple",
			Attrs:  mkAttrs(attribute.String(strings.Repeat("X", 1024), strings.Repeat("Y", 1024))),
			Expect: mkAttrs(attribute.String(strings.Repeat("X", Limit), strings.Repeat("Y", Limit))),
		},
		{
			Name: "several",
			Attrs: mkAttrs(
				attribute.String("small", "unmodified"),
				attribute.String(strings.Repeat("X", 1024), strings.Repeat("Y", 1024)),
				attribute.String("small", "unmodified"),
			),
			Expect: mkAttrs(
				attribute.String("small", "unmodified"),
				attribute.String(strings.Repeat("X", Limit), strings.Repeat("Y", Limit)),
				attribute.String("small", "unmodified"),
			),
		},
		{
			Name: "slice",
			Attrs: mkAttrs(
				attribute.StringSlice("slice", []string{
					"simple",
					strings.Repeat("X", 1024), strings.Repeat("Y", 1024),
					"extra",
				}),
			),
			Expect: mkAttrs(
				attribute.StringSlice("slice", []string{
					"simple",
					strings.Repeat("X", Limit), strings.Repeat("Y", Limit),
					"extra",
				}),
			),
		},
		{
			Name: "mixed",
			Attrs: mkAttrs(
				attribute.Int("int", 1),
				attribute.Float64("float", 1),
				attribute.Bool("bool", true),
				attribute.String("str", strings.Repeat("Y", 1024)),
			),
			Expect: mkAttrs(
				attribute.Int("int", 1),
				attribute.Float64("float", 1),
				attribute.Bool("bool", true),
				attribute.String("str", strings.Repeat("Y", Limit)),
			),
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			perf := Performance{}.Validate()
			perf.AttributeSizeLimit = Limit
			require.Equal(t,
				test.Expect,
				perf.TruncateAttributes(test.Attrs))
		})
	}
}
