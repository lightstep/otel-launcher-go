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

|Config Option     |Env Variable      |Required|Default|
|------------------|------------------|--------|-------|
|WithServiceName            |LS_SERVICE_NAME                    |y       |-                               |
|WithServiceVersion         |LS_SERVICE_VERSION                 |n       |unknown                         |
|WithHeaders                |OTEL_EXPORTER_OTLP_HEADERS         |n       |{}                              |
|WithSpanExporterEndpoint   |OTEL_EXPORTER_OTLP_SPAN_ENDPOINT   |n       |ingest.lightstep.com:443        |
|WithSpanExporterInsecure   |OTEL_EXPORTER_OTLP_SPAN_INSECURE   |n       |false                           |
|WithMetricExporterEndpoint |OTEL_EXPORTER_OTLP_METRIC_ENDPOINT |n       |ingest.lightstep.com:443        |
|WithMetricExporterInsecure |OTEL_EXPORTER_OTLP_METRIC_INSECURE |n       |false                           |
|WithAccessToken            |LS_ACCESS_TOKEN                    |n       |-                               |
|WithLogLevel               |OTEL_LOG_LEVEL                     |n       |info                            |
|WithPropagators            |OTEL_PROPAGATORS                   |n       |b3                              |
|WithResourceAttributes     |OTEL_RESOURCE_ATTRIBUTES           |n       |-                               |
|WithMetricReportingPeriod  |OTEL_EXPORTER_OTLP_METRIC_PERIOD   |n       |30s                             |
|WithMetricsEnabled         |LS_METRICS_ENABLED                 |n       |True                            |

### Principles behind Launcher

##### 100% interoperability with OpenTelemetry

One of the key principles behind putting together Launcher is to make lives of OpenTelemetry users easier, this means that there is no special configuration that **requires** users to install Launcher in order to use OpenTelemetry. It also means that any users of Launcher can leverage the flexibility of configuring OpenTelemetry as they need.

##### Validation

Another decision we made with launcher is to provide end users with a layer of validation of their configuration. This provides us the ability to give feedback to our users faster, so they can start collecting telemetry sooner.

Start using it today in [Go](https://github.com/lightstep/otel-launcher-go), [Java](https://github.com/lightstep/otel-launcher-java), [Javascript](https://github.com/lightstep/otel-launcher-node) and [Python](https://github.com/lightstep/otel-launcher-python) and let us know what you think!

------

*Made with* ![:heart:](https://a.slack-edge.com/production-standard-emoji-assets/10.2/apple-medium/2764-fe0f.png) *@ [Lightstep](http://lightstep.com/)*
