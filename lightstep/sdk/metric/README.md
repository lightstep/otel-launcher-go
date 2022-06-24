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

1. [ExponentialHistogram](./aggregator/histogram/README.md) is the
   default aggregation for Histogram instruments.  The
   explicit-boundary histogram aggregation is not supported.
2. [MinMaxSumCount](./aggregator/minmaxsumcount/README.md) is an
   optional aggregation for Histogram instruments that encodes a
   [zero-bucket explicit-boundary histogram data
   point](https://opentelemetry.io/docs/reference/specification/metrics/datamodel/#histogram).
   Note that this aggregation only encodes the `.Min` and `.Max`
   fields when configured with delta temporality.  [Consider using the
   "stateless" temporality preference in the launcher.](../../../README.md#temporality-settings).
3. The OTLP exporter is the only provided exporter.

These differences aside, this SDK features a complete implementation
of the OpenTelemetry SDK specification with support for multiple
readers.  It is possible, for example, to configure multiple OTLP
exporters with different views and destinations.

This SDK re-uses substantial portions of the community metrics SDK,
including the OTLP gRPC client and the ExponentialHistogram mapping
functions.

Lightstep expects to continue maintaining this implementation until
the community SDK supports configuring the behaviors listed above.
Moreover, Lightstep expects to make several optimizations in this SDK
to further optimize the synchronous instrument fast path and continue
improving memory performance.

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
