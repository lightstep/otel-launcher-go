module github.com/lightstep/otel-launcher-go

go 1.14

require (
	github.com/sethvargo/go-envconfig v0.3.0
	github.com/stretchr/testify v1.6.1
	go.opentelemetry.io/collector v0.7.0
	go.opentelemetry.io/contrib/instrumentation/host v0.14.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.14.0
	go.opentelemetry.io/contrib/propagators v0.14.0
	go.opentelemetry.io/otel v0.14.0
	go.opentelemetry.io/otel/exporters/otlp v0.14.0
	go.opentelemetry.io/otel/sdk v0.14.0
	google.golang.org/grpc v1.32.0
)
