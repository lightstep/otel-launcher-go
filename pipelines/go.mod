module github.com/lightstep/otel-launcher-go/pipelines

go 1.16

require (
	go.opentelemetry.io/contrib/instrumentation/host v0.32.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.32.0
	go.opentelemetry.io/contrib/propagators/b3 v1.7.0
	go.opentelemetry.io/contrib/propagators/ot v1.7.0
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.30.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.30.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.7.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.7.0
	go.opentelemetry.io/otel/metric v0.30.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/sdk/metric v0.30.0
	google.golang.org/grpc v1.46.0
)

require (
	go.opentelemetry.io/proto/otlp v0.16.0
	golang.org/x/net v0.0.0-20220111093109-d55c255bac03 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220112215332-a9c7c0acf9f2 // indirect
)
