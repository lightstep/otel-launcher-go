![build status](https://github.com/lightstep/otel-go/workflows/build/badge.svg)
[![Docs](https://godoc.org/github.com/lightstep/otel-go/locl?status.svg)](https://pkg.go.dev/github.com/lightstep/otel-go/locl)
[![Go Report Card](https://goreportcard.com/badge/github.com/lightstep/otel-go/locl)](https://goreportcard.com/report/github.com/lightstep/otel-go/locl)

# Lightstep OpenTelemetry Configuration Layer for Go

*NOTE: the code in this repo is currently in alpha and will likely change*

This is the Lightstep package for configuring OpenTelemetry

### Install

```bash
go get github.com/lightstep/otel-go/locl
```

### Configure

Minimal setup

```go
import "github.com/lightstep/otel-go/locl"

func main() {
    lightstepOtel := locl.ConfigureOpentelemetry(
        locl.WithServiceName("service-name"),
        locl.WithAccessToken("access-token"),
    )
    defer lightstepOtel.Shutdown()
}
```

Additional options

### Configuration Options

|Config Option     |Env Variable      |Required|Default|
|------------------|------------------|--------|-------|
|WithServiceName    |LS_SERVICE_NAME                    |y       |-                               |
|WithServiceVersion |LS_SERVICE_VERSION                 |n       |unknown                         |
|WithTraceEndpoint  |OTEL_EXPORTER_OTLP_SPAN_ENDPOINT   |n       |ingest.lightstep.com:443        |
|WithSpanExporterEndpointInsecure  |OTEL_EXPORTER_OTLP_SPAN_INSECURE   |n       |false                           |
|WithMetricExporterEndpoint |OTEL_EXPORTER_OTLP_METRIC_ENDPOINT |n       |ingest.lightstep.com:443/metrics|
|WithMetricInsecure |OTEL_EXPORTER_OTLP_METRIC_INSECURE |n       |false                           |
|WithAccessToken    |LS_ACCESS_TOKEN                    |n       |-                               |
|WithLogLevel       |OTEL_LOG_LEVEL                     |n       |info                            |
|WithPropagators    |OTEL_PROPAGATORS                   |n       |b3                              |
|WithResourceLabels |OTEL_RESOURCE_LABELS               |n       |-                               |