## Lightstep implementation of the OpenTelemetry-Go Metrics SDK

This implementation began as a prototype implementation of the
OpenTelemetry SDK specification and has been used as a reference point
for the OpenTelemetry-Go community metrics SDK.  Lightstep's internal
code base is largely written in Go and has a number of specific
requirements which made a direct upgrade to the OpenTelemetry-Go
community Metrics SDK difficult.

Instead of waiting for the OpenTelemetry-Go community SDK to surpass
the current OpenTelemetry specification, to support our requirements,
we decided to enable the experimental features in an SDK we control.

This implementation of the OpenTelemetry SDK has been made public,
however it is not covered by any stability guarantee.

Differences from the OpenTelemetry metrics SDK specification:

1. [ExponentialHistogram](./aggregator/histogram/structure/README.md) is the
   default aggregation for Histogram instruments.  The
   explicit-boundary histogram aggregation is not supported.
2. [MinMaxSumCount](./aggregator/minmaxsumcount/README.md) is an
   optional aggregation for Histogram instruments that encodes a
   [zero-bucket explicit-boundary histogram data
   point](https://opentelemetry.io/docs/reference/specification/metrics/datamodel/#histogram).
   Note that this aggregation only encodes the `.Min` and `.Max`
   fields when configured with delta temporality.  [Consider using the
   "lowmemory" temporality preference in the launcher.](../../../README.md#temporality-settings).
3. The OTLP exporter is the only provided exporter.  The OTLP exporter
   is based on the OTel-Core `batchprocessor` and the OTel-Arrow
   `otlpexporter` collector components.

These differences aside, this SDK features a complete implementation
of the OpenTelemetry SDK specification with support for multiple
readers.  It is possible, for example, to configure multiple OTLP
exporters with different views and destinations.

Lightstep expects to continue maintaining this implementation until
the community SDK supports configuring the behaviors listed above.
Moreover, Lightstep expects to make several optimizations in this SDK
to further optimize the synchronous instrument fast path and continue
improving memory performance.

### Using the Lightstep Metrics SDK directly

See the example in
[examples/example_test.go](./examples/example_test.go) for an example
using this SDK with minimum configuration.

### Metric instrument "Hints" API

There is a standing feature request in OpenTelemetry for a "Hints" API
to inform the SDK of recommended aggregations in the source, when
registering instruments.  This SDK implements an experimental form of
Hints API, described as follows.

The Views implementation attempts to parse the Description of each
metric instrument as the JSON-encoded form of a
`(github.com/lightstep/otel-launcher-go/lightstep/sdk/metric/view).Hint`
structure.  If successfully parsed, the embedded aggrgation kind and
configuration will be used, and the embedded Description field
replaces the original hint.

For example, to set the number of exponential histogram buckets, use a
description like this:

```
{
  "description": "measurement of ...",
  "config": {
    "histogram": {
      "max_size": 320
    }
  }
}
```

To set the MinMaxSumCount aggregation for a specific histogram instrument:

```
{
  "description": "measurement of ...",
  "aggregation": "minmaxsumcount"
}
```

To override the Temporality selected for a specific instrument:

```
{
  "description": "measurement of ...",
  "temporality": "delta"
}
```

### Synchronous Gauge instrument 

[OpenTelemetry metrics API does not support a synchronous Gauge
instrument, however the desired semantics are fairly
clear.](https://github.com/open-telemetry/opentelemetry-specification/issues/2318)
This SDK supports the intended behavior of a synchronous Gauge
instrument by distinguishing two possible behaviors as a function of
the configured temporality.

Although the Gauge data point does not have the concept of temporality
itself, the decision to report or not report a Gauge data point has
traditionally been made in one of two ways:

- When the output system is generally expecting cumulative Counters
(as in Prometheus), it is traditional to report the latest Gauge value
indefinitely, even when the instrument and attribute set are not used
again.
- When the output system is generally expecting delta Counters (as in
Statsd), it is traditional to report Gauge values at most once.

Therefore, when the Temporality selector for the instrument returns
Delta and the aggregation is a Gauge (which is only possible with a
Hint, at this time), the resulting instrument will be a synchronous
Gauge instrument.

For example, to configure a synchronous Gauge:

```
    gauge, _ := meter.SyncUpDownCounter(
	    "some_gauge",
	    instrument.WithDescription(`{"aggregation": "gauge"}`),
	)
```

Note that the API hint for Gauge aggregation can be combined with the
API hint for temporality, allowing control over Gauge behavior
independent of the default Temporality choice for UpDownCounter
instruments.

### Performance settings

The `WithPerformance()` option supports control over performance
settings.

#### IgnoreCollisions

With `IgnoreCollisions` set to true, the SDK will ignore fingerprint
collisions and bypass a safety mechanism that ensures correctness in
spite of fingerprint collisions in the fast path for synchronous
instruments.

#### InactiveCollectionPeriods

The `InactiveCollectionPeriods` setting controls how many `Collect()`
cycles with no updates are allowed before a record is removed from
intermediate state.  Larger values indicate more memory will be used
and that callers are less likely to block while creating new
aggregators.

Setting this field to 1 means records will be removed from memory
after one inactive collection cycle.

Setting this field to 0 causes the default value 10 to be used.

#### InstrumentCardinalityLimit

Synchronous instruments are implemented using a map of intermediate
state.  When this map grows to `InstrumentCardinalityLimit`, new
attribute sets will be replaced by the overflow attribute set, which
is `{ otel.metric.overflow=true }`.  This limit is applied to all
instruments regardless of view configuration before attribute filters
are applied.

For instruments configured with Delta temporality, where it is
possible for the map to shrink, note that the size of this map
includes records maintained due to `InactiveCollectionPeriods`.  The
inactivity period should be taken into account when setting
`InstrumentCardinalityLimit` to avoid overflow.

#### AggregatorCardinalityLimit

All views maintain a configurable cardinality limit, calculated after
attribute filters are applied.

When the aggregator's output grows to `AggregatorCardinalityLimit`,
new attribute sets will be replaced by the overflow attribute set,
which is `{ otel.metric.overflow=true }`.

#### MeasurementProcessor

The `MeasurementProcessor` interface that makes it possible to extend
the set of attributes from synchronous instrument events, which allows
metric attributes to be generated from the OpenTelemetry request
context and/or W3C Tracecontext baggage.

This hook also supports removing attributes from metric events based
on attribute value before they are aggregated, for example to
dynamically configure allowed cardinality values.

#### AttributeSizeLimit

This limit is used to truncate attribute key and string values to a
reasonable size.  The default limit is 8kB.  Zero is not a valid
limit.

#### Exemplars

**Status**: Experimental

Exemplars are sample measurements associated with synchronous metric
instruments.  When OpenTelemetry tracing is used in conjunction with
this Metrics SDK, exemplars will be annotated with the TraceID and
SpanID of the traced context.

Collection of metric exemplars are off by default.  The
`sdkinstrument.Performance.ExemplarsEnabled` field can be used to
enable exemplars by default.  This field may be set to a number of
exemplars to collect by default for all Counter and Histogram
instruments.

Exemplars can also be configured using the `aggregator.Config.Exemplar`
structure, or with a hint like:

```
{
  "description": "measurement of ...",
  "config": {
    "exemplar": {
      "size": 10,
	  "filter": "trace_based",
    }
  }
}
```

Like the OpenTelemetry specification, the supported filters are
"always_off", "always_on", and "trace_based".  Unlike the
OpenTelemetry specification, this SDK has two reservoir
implementations:

- "Last": always chooses the last metric event as the exemplar.  This
  method is automatically selected when the size is 1.
- "Weighted": uses a weighted sampling technique.

With weighted sampling, an unbiased sampler is used such that the
distribution of values can be estimated from the exemplars.  Each
exemplar includes a `sample.weight` attribute indicating its
contribution to the aggregate value.

Note that this weighted sampling property does not apply to
UpDownCounter instruments, because they allow negative measurements.
however these instruments can still generate exemplars.

As a simple example, consider a counter instrument with two input
attribute values counting blue and yellow items, i.e.,

```
counter := meter.Int64Counter("...")

for _ := range BLUE {
  ctx, span := tracer.Start(...)
  counter.Add(ctx, 1, attribute.String("color", "yellow")
  span.End()
}

for _ := range YELLOW {
  ctx, span := tracer.Start(...)
  counter.Add(ctx, 1, attribute.String("color", "blue")
  span.End()
}
```

Suppose the attribute is filtered, so that a single timeseries is
generated.  The aggregate sum equals `len(YELLOW) + len(BLUE)`.

Each exemplar will have one of the filtered attributes, `color=yellow`
or `color=blue`, Each exemplar will have a value of 1, in this case
(i.e., the original measurement).  The ratio of exemplars with
`color=yellow` or `color=blue` will match the ratio of counts
associated with each.

If the number of BLUE items is 3000, and the number of YELLOW items is
7000, and the number of exemplars is 1000, then:

- We expect 300 BLUE examplars
- We expect 700 YELLOW examplars
- Each exemplar has a `sample.weight` of 10, the ratio of total count to exemplar count.

Note that we expect the sum of `sample.weight` for the exemplars to
equal the total number of input events (i.e., 3000 BLUE, 7000 YELLOW).

