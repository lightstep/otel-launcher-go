module github.com/lightstep/otel-launcher-go/pipelines

go 1.18

require (
	// Host and runtime instrumentation
	go.opentelemetry.io/contrib/instrumentation/host v0.41.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.41.0

	// b3 and opentracing propagators
	go.opentelemetry.io/contrib/propagators/b3 v1.16.0
	go.opentelemetry.io/contrib/propagators/ot v1.16.0
	go.opentelemetry.io/otel v1.15.0

	// Standard trace SDK and gRPC OTLP exporter
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.15.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.15.0
	go.opentelemetry.io/otel/metric v0.38.0
	go.opentelemetry.io/otel/sdk v1.15.0
	go.opentelemetry.io/otel/sdk/metric v0.38.0

	// gRPC
	google.golang.org/grpc v1.54.0
)

require (
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/shirou/gopsutil/v3 v3.23.3 // indirect
	github.com/stretchr/testify v1.8.2
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.15.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.15.0 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/protobuf v1.30.0
)

require (
	github.com/lightstep/otel-launcher-go/lightstep/instrumentation v1.16.0
	github.com/lightstep/otel-launcher-go/lightstep/sdk/metric v1.16.0
	go.opentelemetry.io/collector v0.76.1
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.38.0
)

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/apache/arrow/go/v12 v12.0.0-20230404000714-f02d35119ae6 // indirect
	github.com/apache/thrift v0.16.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/f5/otel-arrow-adapter v0.0.0-20230407165153-7ac49c4f4aae // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/goccy/go-json v0.9.11 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/flatbuffers v2.0.8+incompatible // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/asmfmt v1.3.2 // indirect
	github.com/klauspost/compress v1.16.5 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/knadh/koanf v1.5.0 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/minio/asm2plan9s v0.0.0-20200509001527-cdd76441f9d8 // indirect
	github.com/minio/c2goasm v0.0.0-20190812172519-36a3d3bbc4f3 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.1.17 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/shoenig/go-m1cpu v0.1.4 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/collector/component v0.76.1 // indirect
	go.opentelemetry.io/collector/confmap v0.76.1 // indirect
	go.opentelemetry.io/collector/consumer v0.76.1 // indirect
	go.opentelemetry.io/collector/exporter v0.76.1 // indirect
	go.opentelemetry.io/collector/featuregate v0.76.1 // indirect
	go.opentelemetry.io/collector/pdata v1.0.0-rcv0011 // indirect
	go.opentelemetry.io/collector/processor/batchprocessor v0.76.1 // indirect
	go.opentelemetry.io/collector/receiver v0.76.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.41.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/lightstep/otel-launcher-go/lightstep/sdk/metric => ../lightstep/sdk/metric

replace github.com/lightstep/otel-launcher-go/lightstep/instrumentation => ../lightstep/instrumentation
