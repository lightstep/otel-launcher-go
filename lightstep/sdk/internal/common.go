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
	"errors"
	"fmt"
	"strings"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	metricapi "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/embedded"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
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
		panic(fmt.Errorf("unhandled case: %v", inA.Value.Type()))
	}
}

type notelMeterProvider struct {
	embedded.MeterProvider
	meterProvider metric.MeterProvider
}

type notelMeter struct {
	embedded.Meter
	meter metric.Meter
}

func (no notelMeterProvider) Meter(name string, opts ...metric.MeterOption) metric.Meter {
	return notelMeter{
		meter: no.meterProvider.Meter(name, opts...),
	}
}

func fixName(name string) string {
	if strings.HasPrefix(name, "otelcol_") {
		return "otelsdk_" + name[8:]
	}
	return name
}

func (no notelMeter) Int64Counter(name string, options ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	return no.meter.Int64Counter(fixName(name), options...)
}

func (no notelMeter) Int64UpDownCounter(name string, options ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error) {
	return no.meter.Int64UpDownCounter(fixName(name), options...)
}

func (no notelMeter) Int64Histogram(name string, options ...metric.Int64HistogramOption) (metric.Int64Histogram, error) {
	return no.meter.Int64Histogram(fixName(name), options...)
}

func (no notelMeter) Int64Gauge(name string, options ...metric.Int64GaugeOption) (metric.Int64Gauge, error) {
	return no.meter.Int64Gauge(fixName(name), options...)
}

func (no notelMeter) Int64ObservableCounter(name string, options ...metric.Int64ObservableCounterOption) (metric.Int64ObservableCounter, error) {
	return no.meter.Int64ObservableCounter(fixName(name), options...)
}

func (no notelMeter) Int64ObservableUpDownCounter(name string, options ...metric.Int64ObservableUpDownCounterOption) (metric.Int64ObservableUpDownCounter, error) {
	return no.meter.Int64ObservableUpDownCounter(fixName(name), options...)
}

func (no notelMeter) Int64ObservableGauge(name string, options ...metric.Int64ObservableGaugeOption) (metric.Int64ObservableGauge, error) {
	return no.meter.Int64ObservableGauge(fixName(name), options...)
}

func (no notelMeter) Float64Counter(name string, options ...metric.Float64CounterOption) (metric.Float64Counter, error) {
	return no.meter.Float64Counter(fixName(name), options...)
}

func (no notelMeter) Float64UpDownCounter(name string, options ...metric.Float64UpDownCounterOption) (metric.Float64UpDownCounter, error) {
	return no.meter.Float64UpDownCounter(fixName(name), options...)
}

func (no notelMeter) Float64Histogram(name string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	return no.meter.Float64Histogram(fixName(name), options...)
}

func (no notelMeter) Float64Gauge(name string, options ...metric.Float64GaugeOption) (metric.Float64Gauge, error) {
	return no.meter.Float64Gauge(fixName(name), options...)
}

func (no notelMeter) Float64ObservableCounter(name string, options ...metric.Float64ObservableCounterOption) (metric.Float64ObservableCounter, error) {
	return no.meter.Float64ObservableCounter(fixName(name), options...)
}

func (no notelMeter) Float64ObservableUpDownCounter(name string, options ...metric.Float64ObservableUpDownCounterOption) (metric.Float64ObservableUpDownCounter, error) {
	return no.meter.Float64ObservableUpDownCounter(fixName(name), options...)
}

func (no notelMeter) Float64ObservableGauge(name string, options ...metric.Float64ObservableGaugeOption) (metric.Float64ObservableGauge, error) {
	return no.meter.Float64ObservableGauge(fixName(name), options...)
}

func (no notelMeter) RegisterCallback(f metric.Callback, instruments ...metric.Observable) (metric.Registration, error) {
	return no.meter.RegisterCallback(f, instruments...)
}

func NOTelColMeterProvider(m metric.MeterProvider) metric.MeterProvider {
	return &notelMeterProvider{meterProvider: m}
}

const (
	selfTelemetryItemsCounterName = "otelsdk.telemetry.items"
)

func ConfigureSelfTelemetry(
	name string,
	tp trace.TracerProvider,
	mp metric.MeterProvider,
) (component.TelemetrySettings, trace.Tracer, metricapi.Int64Counter, error) {
	var settings component.TelemetrySettings
	// setup logs
	logger, logErr := zap.NewProduction()

	// this replaces otelcol_ w/ otelsdk_
	mp = NOTelColMeterProvider(mp)

	settings.Logger = logger
	settings.TracerProvider = tp
	settings.MeterProvider = mp
	settings.MetricsLevel = configtelemetry.LevelNormal
	settings.LeveledMeterProvider = func(level configtelemetry.Level) metricapi.MeterProvider {
		return mp
	}

	tracer := tp.Tracer(name)
	meter := mp.Meter(name)
	counter, metErr := meter.Int64Counter(selfTelemetryItemsCounterName)
	return settings, tracer, counter, errors.Join(logErr, metErr)
}
