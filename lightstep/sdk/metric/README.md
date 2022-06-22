## Alternate implementation of the OpenTelemetry-Go Metrics SDK

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

1. The exponential histogram is enabled by default; the
   explicit-boundary histogram has been removed
2. The OTLP exporter is the only provided exporter
3. The OTLP exporter supports a "stateless" temporality preference,
   which uses delta temporality for only Counter and Histogram
   instruments
4. There is an alternate aggregation for Histogram instruments,
   "MinMaxSumCount", which encodes as a zero-bucket explicit-boundary
   histogram.

Lightstep expects to continue maintaining this implementation until
the community SDK supports configuring the behaviors listed above.
Moreover, Lightstep expects to make several optimizations in this SDK
to further optimize the synchronous instrument fast path and continue
improving memory performance.