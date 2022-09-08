![build status](https://github.com/lightstep/otel-launcher-go/workflows/build/badge.svg)
[![Docs](https://godoc.org/github.com/lightstep/otel-launcher-go/launcher?status.svg)](https://pkg.go.dev/github.com/lightstep/otel-launcher-go/launcher)
[![Go Report Card](https://goreportcard.com/badge/github.com/lightstep/otel-launcher-go/launcher)](https://goreportcard.com/report/github.com/lightstep/otel-launcher-go/launcher)

# Launcher, a Lightstep Distro for OpenTelemetry 🚀

### What is Launcher?

Launcher is a configuration layer that chooses default values for configuration options that many OpenTelemetry users want. It provides a single function in each language to simplify discovery of the options and components available to users. The goal of Launcher is to help users that aren't familiar with OpenTelemetry quickly ramp up on what they need to get going and instrument.

### Getting started

```bash
go get github.com/lightstep/otel-launcher-go/launcher
```

### Configure

Minimal setup

```go
import "github.com/lightstep/otel-launcher-go/launcher"

func main() {
    otel := launcher.ConfigureOpentelemetry(
        launcher.WithServiceName("service-name"),
        launcher.WithAccessToken("access-token"),
    )
    defer otel.Shutdown()
}
```

For non-lightstep providers, you can set headers directly instead.

```go
import "github.com/lightstep/otel-launcher-go/launcher"

func main() {
    otel := launcher.ConfigureOpentelemetry(
        launcher.WithServiceName("service-name"),
        launcher.WithHeaders(map[string]string{
            "service-auth-key": "value",
            "service-useful-field": "testing",
        }),
    )
    defer otel.Shutdown()
}
```


Additional options

### Configuration Options

| Config Option                           | Env Variable                                     | Required | Default                       |
|-----------------------------------------|--------------------------------------------------|----------|-------------------------------|
| WithServiceName                         | LS_SERVICE_NAME                                  | y        | -                             |
| WithServiceVersion                      | LS_SERVICE_VERSION                               | n        | unknown                       |
| WithHeaders                             | OTEL_EXPORTER_OTLP_HEADERS                       | n        | {}                            |
| WithSpanExporterEndpoint                | OTEL_EXPORTER_OTLP_SPAN_ENDPOINT                 | n        | ingest.lightstep.com:443      |
| WithSpanExporterInsecure                | OTEL_EXPORTER_OTLP_SPAN_INSECURE                 | n        | false                         |
| WithMetricExporterEndpoint              | OTEL_EXPORTER_OTLP_METRIC_ENDPOINT               | n        | ingest.lightstep.com:443      |
| WithMetricExporterInsecure              | OTEL_EXPORTER_OTLP_METRIC_INSECURE               | n        | false                         |
| WithMetricExporterTemporalityPreference | OTEL_EXPORTER_OTLP_METRIC_TEMPORALITY_PREFERENCE | n        | cumulative                    |
| WithAccessToken                         | LS_ACCESS_TOKEN                                  | n        | -                             |
| WithLogLevel                            | OTEL_LOG_LEVEL                                   | n        | info                          |
| WithPropagators                         | OTEL_PROPAGATORS                                 | n        | b3                            |
| WithResourceAttributes                  | OTEL_RESOURCE_ATTRIBUTES                         | n        | -                             |
| WithMetricReportingPeriod               | OTEL_EXPORTER_OTLP_METRIC_PERIOD                 | n        | 30s                           |
| WithMetricsEnabled                      | LS_METRICS_ENABLED                               | n        | true                          |
| WithMetricsBuiltinsEnabled              | LS_METRICS_BUILTINS_ENABLED                      | n        | true                          |
| WithMetricsBuiltinLibraries             | LS_METRICS_BUILTIN_LIBRARIES                     | n        | all:stable |
| WithLightstepMetricsSDK                 | LS_METRICS_SDK                                   | n        | true                          |

### Principles behind Launcher

#### 100% interoperability with OpenTelemetry

One of the key principles behind putting together Launcher is to make lives of OpenTelemetry users easier, this means that there is no special configuration that **requires** users to install Launcher in order to use OpenTelemetry. It also means that any users of Launcher can leverage the flexibility of configuring OpenTelemetry as they need.

#### Validation

Another decision we made with launcher is to provide end users with a layer of validation of their configuration. This provides us the ability to give feedback to our users faster, so they can start collecting telemetry sooner.

Start using it today in [Go](https://github.com/lightstep/otel-launcher-go), [Java](https://github.com/lightstep/otel-launcher-java), [Javascript](https://github.com/lightstep/otel-launcher-node) and [Python](https://github.com/lightstep/otel-launcher-python) and let us know what you think!

### OpenTelemetry Metrics support

### Builtin metrics libraries

The Launcher optionally starts selected libraries of builtin
instrumentation unless explicitly disabled.  Use
`WithMetricsBuiltinsEnabled(false)` or set
`LS_METRICS_BUILTINS_ENABLED=false` to disable all builtin metrics
instrumentation.

The set of builtin metrics libraries can be configured explicitly.
Due to the evolving nature of OpenTelemetry, the metrics produced in
these libraries may vary.  Use `WithMetricsBuiltinLibraries(...)` or
set `LS_METRICS_BUILTIN_LIBRARIES=...` to configure which builtin
libraries are used.  The argument is a comma-separated list of short
library names, optionally `short_name:version` to select a specific
version.  When no version is specified, the "stable" instrumentation
version is selected.

In Launcher version 1.11.0, the default builtin libraries changed to
more closely-track the OpenTelemetry metrics specification.  The
default setting ("all:stable") selects three libraries, described
below:

| Short name | Metrics produced                                                                          |
|------------|-------------------------------------------------------------------------------------------|
| host       | system.cpu.time, system.memory.usage, system.memory.utilization, system.network.io        |
| cputime    | process.cpu.time, process.uptime, process.runtime.go.gc.cpu.time                          |
| runtime    | process.runtime.go.* generated from [runtime/metrics](https://pkg.go.dev/runtime/metrics) |

The available instrumentation library and versions are listed below:

| Short name | Version   | Instrumentation                                                                                                                                  |
|------------|-----------|--------------------------------------------------------------------------------------------------------------------------------------------------|
| all        | stable    | same as cputime:stable,host:stable,runtime:stable                                                                                                |
| cputime    | stable    | `./lightstep/instrumentation/cputime`                                                                                                            |
| cputime    | prestable | _none_                                                                                                                                           |
| runtime    | stable    | `./lightstep/instrumentation/runtime`                                                                                                            |
| runtime    | prestable | [opentelemetry-go-contrib/instrumentation/runtime](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/runtime) |
| host       | stable    | `./lightstep/instrumentation/host`                                                                                                               |
| host       | prestable | [opentelemetry-go-contrib/instrumentation/host](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/host)       |

When running inside a container where host metrics are already being
collected (e.g., using the [OpenTelemetry Collector `hostmetrics`
receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver)),
users may wish to configure only "cputime" and "runtime" instrumentation.

Note that the "cputime" library is a placeholder for metrics that
overlap with or that we expect to be produced in future versions of
the "runtime" instrumentation.  For example, [we expect the Go-1.20
`runtime/metrics` package to include GC and user CPU time
metrics](https://go-review.googlesource.com/c/go/+/404307/).  See the
documentation in each of the instrumentation packages for more
details.

To configure the behavior in Launcher versions 1.10.x and prior, use
`WithMetricsBuiltinLibraries("all:prestable")` or set
`LS_METRICS_BUILTIN_LIBRARIES=all:prestable`.

### Lightstep Metrics SDK

**NOTE**: The Lightstep Metrics SDK is enabled by default.

The Launcher contains an alternative to the [OTel-Go community Metrics
SDK](https://github.com/open-telemetry/opentelemetry-go) being
maintained by Lightstep as a way to quickly validate newer
OpenTelemetry features, such as the OpenTelemetry exponential
histogram.

The OTel-Go community SDK is not enabled by default.  This option will
return when the OTel-Go community SDK reaches a stable release.

To select the OTel-Go community Metrics SDK, use
`WithLightstepMetricsSDK(false)` or set `LS_METRICS_SDK=false`.

The differences between the OpenTelemetry Metrics SDK specification
and the alternative SDK are documented in its
[README](./lightstep/sdk/metric/README.md).

### Metrics Temporality settings

OpenTelemetry metrics SDKs give the user control over "temporality",
which is the selection of "delta" or "cumulative" policies for
aggregating Counter and Histogram instruments.  These settings determine
both memory usage and reliability of metrics reporting.

Delta temporality requires less memory than cumulative temporality for
synchronous instruments, while Cumulative requires less memory than
delta temporality for asynchronous instruments.  When reporting is
intermittent, cumulative series will average out the missing reports,
whereas delta series will have gaps.

Note that Lightstep considers a change of temporality to be a breaking
change.  Once a temporality preference has been set, the setting has
to be maintained.  The temporality preference is configured by calling
`WithMetricExporterTemporalityPreference()` or using the
`OTEL_EXPORTER_OTLP_METRIC_TEMPORALITY_PREFERENCE` environment
variable.

The launcher supports a "stateless" temporality preference.  This
selection configures the ideal behavior for Lightstep by mixing
temporality setings.

The 1.x launcher release series configures the "cumulative"
temporality preference by default.  The next major release of launcher
will configure the "stateless" temporality preference.

Lightstep users are recommended to select either the "cumulative" or
"stateless" preference.  The OpenTelemetry-specified "delta"
temporality preference is not recommended for Lightstep users.

### Metrics validation errors

Lightstep performs a number of validation steps over metrics data
before accepting it.  When an OTLP Metrics export request is
successful but data is completely or partially rejected for any
reason, the outcome is detailed using Lightstep-specific response
headers. 

These headers predate [work in OpenTelemetry on returning partial
success](https://github.com/open-telemetry/opentelemetry-proto/pull/390).
Lightstep expects to use standard OTLP fields to convey these
partially-successful outcomes in the future.

Validation errors are generally repetitive.  Lightstep limits the size
of each partial-success response to lower the overhead associated with
these responses using random selection.

The launcher contains special code to interpret these headers and
direct them to the standard OpenTelemetry-Go error handler.

------

*Made with*
![:heart:](https://a.slack-edge.com/production-standard-emoji-assets/10.2/apple-medium/2764-fe0f.png) *@ [Lightstep](http://lightstep.com/)*
