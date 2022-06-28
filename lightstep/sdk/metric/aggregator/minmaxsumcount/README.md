# MinMaxSumCount aggregator

## Low-cost histogram aggregation

This aggregation offers a low-cost alternative to the
ExponentialHistogram for aggregating OpenTelemetry Histogram
instruments.  This aggregator tracks the Sum, Count, Min, and Max of
values recorded by the instrument.

### Delta temporality recommended

The OTLP exporter will only export the Min and Max fields when used
with Delta temporality.  [See the discussion of Histogram Min/Max and
Aggregation Temporality in the OpenTelemetry metrics data
model.](https://opentelemetry.io/docs/reference/specification/metrics/datamodel/#histogram)


### Configuring MinMaxSumCount

To select this aggregation, configure a View or the default
aggregation to `aggregation.MinMaxSumCountKind`, e.g.,

```
view.WithDefaultAggregationKindSelector(func(k sdkinstrument.Kind) aggregation.Kind {
	switch k {
	case sdkinstrument.SyncHistogram:
	    return aggregation.MinMaxSumCountKind
    case sdkinstrument.AsyncGauge:
	    return aggregation.GaugeKind
    default:
	    return aggregation.AnySumKind
	}
}
```

Or use an instrument hint, e.g.,

```
	histogram, err := meter.SyncFloat64().Histogram("some.measurement",
		instrument.WithDescription(`{
  "aggregation": "minmaxsumcount"
}`),
	)
```

### OTLP Encoding for MinMaxSimCount

MinMaxSumCount aggregators are encoding using zero-bucket
explicit-boundary histogram data points.  Lightstep treats these
similarly to the OTLP `SummaryDataPoint`: they are translated into
four timeseries each.

- `{metric_name}.sum`: Sum of recorded values
- `{metric_name}.count`: Count of recorded values
- `{metric_name}.min`: Minimum recorded value (with delta temporality)
- `{metric_name}.max`: Minimum recorded value (with delta temporality)
