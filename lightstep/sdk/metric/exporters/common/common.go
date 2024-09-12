package common

import (
	"context"
	"errors"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/internal"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/data"
	"github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/internal/convert"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	metricapi "go.opentelemetry.io/otel/metric"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

const (
	selfTelemetryTracerName = "lightstep-go/sdk/metric"
	selfTelemetryMeterName  = "lightstep-go/sdk/metric"

	selfTelemetryItemsCounterName = "otelsdk.telemetry.items"
)

func ConfigureSelfTelemetry(
	selfSpans bool,
	selfMetrics bool,
	exporterSettings *exporter.Settings,
	tracer *trace.Tracer,
	telemetryItemsCounter *metricapi.Int64Counter,
) error {
	// setup logs
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}

	exporterSettings.TelemetrySettings.Logger = logger

	// setup traces
	tracerProvider := trace.TracerProvider(nooptrace.NewTracerProvider())
	if selfSpans {
		tracerProvider = otel.GetTracerProvider()
	}

	exporterSettings.TelemetrySettings.TracerProvider = tracerProvider
	*tracer = exporterSettings.TelemetrySettings.TracerProvider.Tracer(selfTelemetryTracerName)

	// setup metrics
	meterProvider := metricapi.MeterProvider(noopmetric.NewMeterProvider())
	if selfMetrics {
		meterProvider = internal.NOTelColMeterProvider(otel.GetMeterProvider())
	}

	exporterSettings.TelemetrySettings.MetricsLevel = configtelemetry.LevelNormal
	exporterSettings.TelemetrySettings.MeterProvider = meterProvider
	exporterSettings.TelemetrySettings.LeveledMeterProvider = func(level configtelemetry.Level) metricapi.MeterProvider {
		return meterProvider
	}

	// setup the telemetry item counter metric
	meter := exporterSettings.TelemetrySettings.LeveledMeterProvider(configtelemetry.LevelNormal).Meter(selfTelemetryMeterName)
	*telemetryItemsCounter, err = meter.Int64Counter(selfTelemetryItemsCounterName)

	return err
}

func ExportMetrics(
	ctx context.Context,
	data data.Metrics,
	tracer trace.Tracer,
	telemetryItemsCounter metricapi.Int64Counter,
	resourceMap *internal.ResourceMap,
	exporter exporter.Metrics,
) error {
	ctx, span := tracer.Start(
		ctx,
		"otelsdk_export_metrics",
	)
	defer span.End()

	converted := convert.MetricToPMetric(resourceMap, data)
	points := int64(converted.DataPointCount())

	err := exporter.ConsumeMetrics(ctx, converted)
	success := err == nil
	var state string
	if success {
		state = "ok"
	} else if errors.Is(err, context.Canceled) {
		state = "canceled"
	} else if errors.Is(err, context.DeadlineExceeded) {
		state = "timeout"
	} else {
		state = "error"
	}

	var attrs = []attribute.KeyValue{
		attribute.Bool("success", success),
		attribute.String("state", state),
	}
	telemetryItemsCounter.Add(ctx, points, metricapi.WithAttributes(attrs...))
	span.SetAttributes(append(attrs, attribute.Int64("num_points", points))...)
	if err == nil {
		span.SetStatus(otelcodes.Ok, state)
	} else {
		span.SetStatus(otelcodes.Error, state)
	}
	return err
}
