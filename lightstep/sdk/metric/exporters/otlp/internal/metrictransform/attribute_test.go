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

package metrictransform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
)

func testAttributes(attrs []attribute.KeyValue) []*commonpb.KeyValue {
	// This copy is because attribute.NewSet() sorts in place
	cpy := make([]attribute.KeyValue, len(attrs))
	copy(cpy, attrs)
	return Attributes(attribute.NewSet(cpy...))
}

type attributeTest struct {
	attrs    []attribute.KeyValue
	expected []*commonpb.KeyValue
}

func TestAttributes(t *testing.T) {
	for _, test := range []attributeTest{
		{nil, nil},
		{
			[]attribute.KeyValue{
				attribute.Int("int to int", 123),
				attribute.Int64("int64 to int64", 1234567),
				attribute.Float64("float64 to double", 0.5),
				attribute.String("string to string", "string"),
				attribute.Bool("bool to bool", true),
			},
			[]*commonpb.KeyValue{
				{
					Key: "bool to bool",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_BoolValue{
							BoolValue: true,
						},
					},
				},
				{
					Key: "int64 to int64",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_IntValue{
							IntValue: 1234567,
						},
					},
				},
				{
					Key: "int to int",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_IntValue{
							IntValue: 123,
						},
					},
				},
				{
					Key: "float64 to double",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_DoubleValue{
							DoubleValue: 0.5,
						},
					},
				},
				{
					Key: "string to string",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{
							StringValue: "string",
						},
					},
				},
			},
		},
	} {
		got := testAttributes(test.attrs)
		if !assert.Len(t, got, len(test.expected)) {
			continue
		}
		assert.ElementsMatch(t, test.expected, got)
	}
}

func TestArrayAttributes(t *testing.T) {
	// Array KeyValue supports only arrays of primitive types:
	// "bool", "int", "int64",
	// "float64", "string",
	for _, test := range []attributeTest{
		{nil, nil},
		{
			[]attribute.KeyValue{
				{
					Key:   attribute.Key("invalid"),
					Value: attribute.Value{},
				},
			},
			[]*commonpb.KeyValue{
				{
					Key: "invalid",
					Value: &commonpb.AnyValue{
						Value: &commonpb.AnyValue_StringValue{
							StringValue: "INVALID",
						},
					},
				},
			},
		},
		{
			[]attribute.KeyValue{
				attribute.BoolSlice("bool slice to bool array", []bool{true, false}),
				attribute.IntSlice("int slice to int64 array", []int{1, 2, 3}),
				attribute.Int64Slice("int64 slice to int64 array", []int64{1, 2, 3}),
				attribute.Float64Slice("float64 slice to double array", []float64{0.5, 0.25, 0.125}),
				attribute.StringSlice("string slice to string array", []string{"foo", "bar", "baz"}),
			},
			[]*commonpb.KeyValue{
				newOTelBoolArray("bool slice to bool array", []bool{true, false}),
				newOTelIntArray("int slice to int64 array", []int64{1, 2, 3}),
				newOTelIntArray("int64 slice to int64 array", []int64{1, 2, 3}),
				newOTelDoubleArray("float64 slice to double array", []float64{0.5, 0.25, 0.125}),
				newOTelStringArray("string slice to string array", []string{"foo", "bar", "baz"}),
			},
		},
	} {
		actualArrayAttributes := testAttributes(test.attrs)
		expectedArrayAttributes := test.expected
		if !assert.Len(t, actualArrayAttributes, len(expectedArrayAttributes)) {
			continue
		}

		assert.ElementsMatch(t, test.expected, actualArrayAttributes)
	}
}

func newOTelBoolArray(key string, values []bool) *commonpb.KeyValue {
	arrayValues := []*commonpb.AnyValue{}
	for _, b := range values {
		arrayValues = append(arrayValues, &commonpb.AnyValue{
			Value: &commonpb.AnyValue_BoolValue{
				BoolValue: b,
			},
		})
	}

	return newOTelArray(key, arrayValues)
}

func newOTelIntArray(key string, values []int64) *commonpb.KeyValue {
	arrayValues := []*commonpb.AnyValue{}

	for _, i := range values {
		arrayValues = append(arrayValues, &commonpb.AnyValue{
			Value: &commonpb.AnyValue_IntValue{
				IntValue: i,
			},
		})
	}

	return newOTelArray(key, arrayValues)
}

func newOTelDoubleArray(key string, values []float64) *commonpb.KeyValue {
	arrayValues := []*commonpb.AnyValue{}

	for _, d := range values {
		arrayValues = append(arrayValues, &commonpb.AnyValue{
			Value: &commonpb.AnyValue_DoubleValue{
				DoubleValue: d,
			},
		})
	}

	return newOTelArray(key, arrayValues)
}

func newOTelStringArray(key string, values []string) *commonpb.KeyValue {
	arrayValues := []*commonpb.AnyValue{}

	for _, s := range values {
		arrayValues = append(arrayValues, &commonpb.AnyValue{
			Value: &commonpb.AnyValue_StringValue{
				StringValue: s,
			},
		})
	}

	return newOTelArray(key, arrayValues)
}

func newOTelArray(key string, arrayValues []*commonpb.AnyValue) *commonpb.KeyValue {
	return &commonpb.KeyValue{
		Key: key,
		Value: &commonpb.AnyValue{
			Value: &commonpb.AnyValue_ArrayValue{
				ArrayValue: &commonpb.ArrayValue{
					Values: arrayValues,
				},
			},
		},
	}
}
