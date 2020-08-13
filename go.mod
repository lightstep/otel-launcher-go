module github.com/lightstep/otel-launcher-go

go 1.14

replace go.opentelemetry.io/otel => ../../../go.opentelemetry.io/

replace go.opentelemetry.io/otel/exporters/otlp => ../../../go.opentelemetry.io/exporters/otlp

replace go.opentelemetry.io/otel/sdk => ../../../go.opentelemetry.io/sdk

replace github.com/open-telemetry/opentelemetry-go-contrib/plugins/runtime => ../../../github.com/open-telemetry/opentelemetry-go-contrib/plugins/runtime

require (
	github.com/open-telemetry/opentelemetry-go-contrib/plugins/runtime v0.0.0-00010101000000-000000000000
	github.com/sethvargo/go-envconfig v0.1.1
	github.com/stretchr/testify v1.6.1
	go.opentelemetry.io/collector v0.7.0
	go.opentelemetry.io/otel v0.10.0
	go.opentelemetry.io/otel/exporters/otlp v0.10.0
	go.opentelemetry.io/otel/sdk v0.10.0
	google.golang.org/grpc v1.31.0
)
