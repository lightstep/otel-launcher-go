module github.com/lightstep/otel-launcher-go/pipelines

go 1.20

require (
	github.com/lightstep/otel-launcher-go/lightstep/instrumentation v1.20.0
	github.com/lightstep/otel-launcher-go/lightstep/sdk/metric v1.20.0
	go.opentelemetry.io/collector/config/configtls v0.87.0
	// Host and runtime instrumentation
	go.opentelemetry.io/contrib/instrumentation/host v0.45.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.45.0

	// b3 and opentracing propagators
	go.opentelemetry.io/contrib/propagators/b3 v1.20.0
	go.opentelemetry.io/contrib/propagators/ot v1.20.0
	go.opentelemetry.io/otel v1.19.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.42.0

	// Standard trace SDK and gRPC OTLP exporter
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.19.0
	go.opentelemetry.io/otel/metric v1.19.0
	go.opentelemetry.io/otel/sdk v1.19.0
	go.opentelemetry.io/otel/sdk/metric v1.19.0

	// gRPC
	google.golang.org/grpc v1.58.2
)

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/apache/arrow/go/v12 v12.0.1 // indirect
	github.com/apache/thrift v0.16.0 // indirect
	github.com/axiomhq/hyperloglog v0.0.0-20230201085229-3ddf4bad03dc // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dgryski/go-metro v0.0.0-20180109044635-280f6062b5bc // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/goccy/go-json v0.9.11 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/flatbuffers v2.0.8+incompatible // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/asmfmt v1.3.2 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/v2 v2.0.1 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/lightstep/otel-launcher-go/lightstep/sdk/internal v1.20.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/minio/asm2plan9s v0.0.0-20200509001527-cdd76441f9d8 // indirect
	github.com/minio/c2goasm v0.0.0-20190812172519-36a3d3bbc4f3 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20220423185008-bf980b35cac4 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.1 // indirect
	github.com/open-telemetry/otel-arrow v0.6.0 // indirect
	github.com/open-telemetry/otel-arrow/collector v0.6.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/shirou/gopsutil/v3 v3.23.9 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/stretchr/testify v1.8.4
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/collector v0.87.0 // indirect
	go.opentelemetry.io/collector/component v0.87.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.87.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v0.87.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.87.0 // indirect
	go.opentelemetry.io/collector/config/confignet v0.87.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v0.87.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.87.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.87.0 // indirect
	go.opentelemetry.io/collector/confmap v0.87.0 // indirect
	go.opentelemetry.io/collector/consumer v0.87.0 // indirect
	go.opentelemetry.io/collector/exporter v0.87.0 // indirect
	go.opentelemetry.io/collector/extension v0.87.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.87.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.0.0-rcv0016 // indirect
	go.opentelemetry.io/collector/pdata v1.0.0-rcv0016 // indirect
	go.opentelemetry.io/collector/processor v0.87.0 // indirect
	go.opentelemetry.io/collector/processor/batchprocessor v0.87.0 // indirect
	go.opentelemetry.io/collector/receiver v0.87.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.42.0 // indirect
	go.opentelemetry.io/otel/trace v1.19.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	golang.org/x/mod v0.11.0 // indirect
	golang.org/x/net v0.15.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/tools v0.9.1 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230711160842-782d3b101e98 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230711160842-782d3b101e98 // indirect
	google.golang.org/protobuf v1.31.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require github.com/knadh/koanf/providers/confmap v0.1.0 // indirect

replace github.com/lightstep/otel-launcher-go/lightstep/sdk/internal => ../lightstep/sdk/internal

replace github.com/lightstep/otel-launcher-go/lightstep/sdk/metric => ../lightstep/sdk/metric

replace github.com/lightstep/otel-launcher-go/lightstep/instrumentation => ../lightstep/instrumentation

// ambiguous import: found package cloud.google.com/go/compute/metadata in multiple modules
replace cloud.google.com/go => cloud.google.com/go v0.110.2
