module github.com/lightstep/otel-launcher-go

go 1.16

require (
	github.com/kr/text v0.2.0 // indirect
	github.com/lightstep/otel-launcher-go/pipelines v1.5.2
	github.com/sethvargo/go-envconfig v0.7.0
	github.com/stretchr/testify v1.7.4
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/metric v0.30.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/lightstep/otel-launcher-go/pipelines => ./pipelines
