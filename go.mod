module github.com/lightstep/otel-launcher-go

go 1.14

require (
	github.com/sethvargo/go-envconfig v0.3.2
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.28.0
	go.opentelemetry.io/contrib/instrumentation/host v0.23.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.23.0
	go.opentelemetry.io/contrib/propagators/b3 v0.23.0
	go.opentelemetry.io/contrib/propagators/ot v0.23.0
	go.opentelemetry.io/otel v1.0.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.23.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.23.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.0.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.0.0
	go.opentelemetry.io/otel/metric v0.23.0
	go.opentelemetry.io/otel/sdk v1.0.0
	go.opentelemetry.io/otel/sdk/metric v0.23.0
	go.opentelemetry.io/otel/trace v1.0.0
	google.golang.org/grpc v1.40.0
)
