module github.com/lightstep/otel-launcher-go

go 1.14

require (
	github.com/sethvargo/go-envconfig v0.3.0
	github.com/stretchr/testify v1.6.1
	go.opentelemetry.io/collector v0.7.0
	go.opentelemetry.io/contrib/instrumentation/host v0.13.1-0.20201013231334-308925fe4f1f
	go.opentelemetry.io/contrib/instrumentation/runtime v0.13.1-0.20201013231334-308925fe4f1f
	go.opentelemetry.io/contrib/propagators v0.13.1-0.20201013231334-308925fe4f1f
	go.opentelemetry.io/otel v0.13.1-0.20201013220238-8ed55f598061
	go.opentelemetry.io/otel/exporters/otlp v0.13.1-0.20201013220238-8ed55f598061
	go.opentelemetry.io/otel/sdk v0.13.1-0.20201013220238-8ed55f598061
	google.golang.org/grpc v1.32.0
)

replace go.opentelemetry.io/contrib => github.com/codeboten/opentelemetry-go-contrib v0.13.1-0.20201013231334-308925fe4f1f

replace go.opentelemetry.io/contrib/instrumentation/host => github.com/codeboten/opentelemetry-go-contrib/instrumentation/host v0.13.1-0.20201013231334-308925fe4f1f

replace go.opentelemetry.io/contrib/instrumentation/runtime => github.com/codeboten/opentelemetry-go-contrib/instrumentation/runtime v0.13.1-0.20201013231334-308925fe4f1f

replace go.opentelemetry.io/contrib/propagators => github.com/codeboten/opentelemetry-go-contrib/propagators v0.13.1-0.20201013231334-308925fe4f1f
