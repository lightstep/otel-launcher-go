module github.com/lightstep/otel-launcher-go

go 1.14

require (
	github.com/sethvargo/go-envconfig v0.3.0
	github.com/stretchr/testify v1.6.1
	go.opentelemetry.io/collector v0.7.0
	go.opentelemetry.io/contrib/instrumentation/host v0.13.1-0.20201013231334-308925fe4f1f
	go.opentelemetry.io/contrib/instrumentation/runtime v0.13.1-0.20201013231334-308925fe4f1f
	go.opentelemetry.io/contrib/propagators v0.13.1-0.20201013231334-308925fe4f1f
	go.opentelemetry.io/otel v0.14.0
	go.opentelemetry.io/otel/exporters/otlp v0.14.0
	go.opentelemetry.io/otel/exporters/stdout v0.14.0 // indirect
	go.opentelemetry.io/otel/sdk v0.14.0
	google.golang.org/grpc v1.32.0
)

replace go.opentelemetry.io/contrib => github.com/codeboten/opentelemetry-go-contrib v0.13.1-0.20201013231334-308925fe4f1f

replace go.opentelemetry.io/contrib/instrumentation/host => github.com/codeboten/opentelemetry-go-contrib/instrumentation/host v0.13.1-0.20201120001640-d9116fb3a5b9

replace go.opentelemetry.io/contrib/instrumentation/runtime => github.com/codeboten/opentelemetry-go-contrib/instrumentation/runtime v0.13.1-0.20201120001640-d9116fb3a5b9

replace go.opentelemetry.io/contrib/propagators => github.com/codeboten/opentelemetry-go-contrib/propagators v0.13.1-0.20201120001036-8283981cbceb
