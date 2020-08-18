module github.com/lightstep/otel-launcher-go

go 1.14

replace go.opentelemetry.io/otel/sdk => ../../../go.opentelemetry.io/sdk

replace go.opentelemetry.io/otel/exporters/otlp => ../../../go.opentelemetry.io/exporters/otlp

replace go.opentelemetry.io/otel => ../../../go.opentelemetry.io/

replace go.opentelemetry.io/contrib/instrumentation/runtime => ../../../github.com/open-telemetry/opentelemetry-go-contrib/instrumentation/runtime

replace go.opentelemetry.io/contrib/instrumentation/host => ../../../github.com/open-telemetry/opentelemetry-go-contrib/instrumentation/host

replace go.opentelemetry.io/contrib => ../../../github.com/open-telemetry/opentelemetry-go-contrib

require (
	go.opentelemetry.io/contrib/instrumentation/runtime v0.10.1
	go.opentelemetry.io/contrib/instrumentation/host v0.10.1
	go.opentelemetry.io/contrib v0.10.1
	github.com/sethvargo/go-envconfig v0.1.1
	github.com/stretchr/testify v1.6.1
	go.opentelemetry.io/collector v0.7.0
	go.opentelemetry.io/otel v0.10.0
	go.opentelemetry.io/otel/exporters/otlp v0.10.0
	go.opentelemetry.io/otel/sdk v0.10.0
	google.golang.org/grpc v1.31.0
)
