![build status](https://github.com/lightstep/otel-go/workflows/build/badge.svg)
[![Docs](https://godoc.org/github.com/lightstep/otel-go/locl?status.svg)](https://pkg.go.dev/github.com/lightstep/otel-go/locl)
[![Go Report Card](https://goreportcard.com/badge/github.com/lightstep/otel-go/locl)](https://goreportcard.com/report/github.com/lightstep/otel-go/locl)

# Lightstep OpenTelemetry Configuration Layer for Go

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
|WithServiceName   |LS_SERVICE_NAME   |y       |-      |
|WithServiceVersion|LS_SERVICE_VERSION|n       |unknown|
|WithSatelliteURL  |LS_SATELLITE_URL  |n       |ingest.lightstep.com:443|
|WithMetricsURL    |LS_METRICS_URL    |n       |ingest.lightstep.com:443/metrics|
|WithAccessToken   |LS_ACCESS_TOKEN   |n       |-      |
|WithDebug         |LS_DEBUG          |n       |false  |
|WithInsecure      |LS_INSECURE       |n       |false  |