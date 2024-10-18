module github.com/lightstep/otel-launcher-go/pipelines

go 1.22.0

toolchain go1.22.6

require (
	github.com/lightstep/otel-launcher-go/lightstep/instrumentation v1.32.0
	github.com/lightstep/otel-launcher-go/lightstep/sdk/metric v1.32.0
	github.com/lightstep/otel-launcher-go/lightstep/sdk/trace v1.31.0

	// For otelcol-based exporter configuration.
	go.opentelemetry.io/collector/config/configtls v1.17.0

	// b3 and opentracing propagators
	go.opentelemetry.io/contrib/propagators/b3 v1.28.0
	go.opentelemetry.io/contrib/propagators/ot v1.28.0
	go.opentelemetry.io/otel v1.31.0
	go.opentelemetry.io/otel/metric v1.31.0
	go.opentelemetry.io/otel/sdk v1.31.0

	// gRPC
	google.golang.org/grpc v1.67.1
)

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/axiomhq/hyperloglog v0.0.0-20230201085229-3ddf4bad03dc // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dgryski/go-metro v0.0.0-20180109044635-280f6062b5bc // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.10 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/lightstep/otel-launcher-go/lightstep/sdk/internal v1.32.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/open-telemetry/otel-arrow v0.27.0 // indirect
	github.com/open-telemetry/otel-arrow/collector/processor/concurrentbatchprocessor v0.28.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/shirou/gopsutil/v3 v3.24.4 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/stretchr/testify v1.9.0
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/collector/component v0.111.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.111.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.17.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.111.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.17.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.17.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.111.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.111.0 // indirect
	go.opentelemetry.io/collector/consumer v0.111.0 // indirect
	go.opentelemetry.io/collector/exporter v0.111.0 // indirect
	go.opentelemetry.io/collector/extension v0.111.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.111.0 // indirect
	go.opentelemetry.io/collector/pdata v1.17.0 // indirect
	go.opentelemetry.io/collector/processor v0.110.0 // indirect
	go.opentelemetry.io/collector/receiver v0.111.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.55.0 // indirect
	go.opentelemetry.io/otel/trace v1.31.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/mod v0.19.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	golang.org/x/tools v0.23.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/protobuf v1.34.2
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/apache/arrow/go/v16 v16.1.0 // indirect
	github.com/apache/arrow/go/v17 v17.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/lightstep/varopt v1.4.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/otelarrowexporter v0.111.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/grpcutil v0.111.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/otelarrow v0.111.0 // indirect
	go.opentelemetry.io/collector/client v1.17.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.17.0 // indirect
	go.opentelemetry.io/collector/consumer/consumerprofiles v0.111.0 // indirect
	go.opentelemetry.io/collector/extension/experimental/storage v0.111.0 // indirect
	go.opentelemetry.io/collector/internal/globalsignal v0.111.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.111.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.111.0 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
)

replace github.com/lightstep/otel-launcher-go/lightstep/sdk/internal => ../lightstep/sdk/internal

replace github.com/lightstep/otel-launcher-go/lightstep/sdk/metric => ../lightstep/sdk/metric

replace github.com/lightstep/otel-launcher-go/lightstep/sdk/trace => ../lightstep/sdk/trace

replace github.com/lightstep/otel-launcher-go/lightstep/instrumentation => ../lightstep/instrumentation

// ambiguous import: found package cloud.google.com/go/compute/metadata in multiple modules
replace cloud.google.com/go => cloud.google.com/go v0.110.2
