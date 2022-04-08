module github.com/lightstep/otel-launcher-go

go 1.16

require (
	github.com/kr/text v0.2.0 // indirect
	github.com/lightstep/otel-launcher-go/pipelines v1.5.1
	github.com/sethvargo/go-envconfig v0.6.0
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/otel v1.6.3
	go.opentelemetry.io/otel/metric v0.27.0
	go.opentelemetry.io/otel/sdk v1.5.0
	go.opentelemetry.io/otel/trace v1.6.3
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/lightstep/otel-launcher-go/pipelines => ./pipelines
