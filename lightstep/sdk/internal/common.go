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
	"sync"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

type ResourceMap struct {
	lock sync.Mutex

	disabled bool
	input    *attribute.Set
	output   pcommon.Resource
}

func (rm *ResourceMap) Get(in *resource.Resource) pcommon.Resource {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	if !rm.disabled && rm.input != nil {
		if *rm.input == *in.Set() {
			// If the same resource happens over and over ...
			return rm.output
		}
		rm.disabled = true
		rm.input = nil
		rm.output = pcommon.Resource{}
	}

	out := pcommon.NewResource()
	CopyAttributes(
		out.Attributes(),
		*in.Set(),
	)

	if !rm.disabled {
		rm.input = in.Set()
		rm.output = out
	}

	return out

}

func CopyAttributes(dest pcommon.Map, src attribute.Set) {
	for iter := src.Iter(); iter.Next(); {
		CopyAttribute(dest, iter.Attribute())
	}
}

func CopyAttribute(dest pcommon.Map, inA attribute.KeyValue) {
	key := string(inA.Key)
	switch inA.Value.Type() {
	case attribute.BOOL:
		dest.PutBool(key, inA.Value.AsBool())
	case attribute.INT64:
		dest.PutInt(key, inA.Value.AsInt64())
	case attribute.FLOAT64:
		dest.PutDouble(key, inA.Value.AsFloat64())
	case attribute.STRING:
		dest.PutStr(key, inA.Value.AsString())
	case attribute.BOOLSLICE:
		sl := dest.PutEmptySlice(key)
		sl.EnsureCapacity(len(inA.Value.AsBoolSlice()))
		for _, v := range inA.Value.AsBoolSlice() {
			sl.AppendEmpty().SetBool(v)
		}
	case attribute.INT64SLICE:
		sl := dest.PutEmptySlice(key)
		sl.EnsureCapacity(len(inA.Value.AsInt64Slice()))
		for _, v := range inA.Value.AsInt64Slice() {
			sl.AppendEmpty().SetInt(v)
		}
	case attribute.FLOAT64SLICE:
		sl := dest.PutEmptySlice(key)
		sl.EnsureCapacity(len(inA.Value.AsFloat64Slice()))
		for _, v := range inA.Value.AsFloat64Slice() {
			sl.AppendEmpty().SetDouble(v)
		}
	case attribute.STRINGSLICE:
		sl := dest.PutEmptySlice(key)
		sl.EnsureCapacity(len(inA.Value.AsStringSlice()))
		for _, v := range inA.Value.AsStringSlice() {
			sl.AppendEmpty().SetStr(v)
		}
	default:
		panic("unhandled case")
	}
}
