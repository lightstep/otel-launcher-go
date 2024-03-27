module github.com/lightstep/otel-launcher-go

go 1.21
toolchain go1.22.1

require (
	github.com/lightstep/otel-launcher-go/pipelines v1.27.0
	github.com/sethvargo/go-envconfig v1.0.1
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/otel v1.24.0
	go.opentelemetry.io/otel/metric v1.24.0
	go.opentelemetry.io/otel/sdk v1.24.0
	go.opentelemetry.io/otel/trace v1.24.0
)

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/apache/arrow/go/v14 v14.0.2 // indirect
	github.com/axiomhq/hyperloglog v0.0.0-20230201085229-3ddf4bad03dc // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dgryski/go-metro v0.0.0-20180109044635-280f6062b5bc // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/flatbuffers v23.5.26+incompatible // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.0 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.7 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.0 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/lightstep/otel-launcher-go/lightstep/instrumentation v1.27.0 // indirect
	github.com/lightstep/otel-launcher-go/lightstep/sdk/internal v1.27.0 // indirect
	github.com/lightstep/otel-launcher-go/lightstep/sdk/metric v1.27.0 // indirect
	github.com/lightstep/otel-launcher-go/lightstep/sdk/trace v1.25.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.2 // indirect
	github.com/open-telemetry/otel-arrow v0.20.0 // indirect
	github.com/open-telemetry/otel-arrow/collector v0.20.0 // indirect
	github.com/open-telemetry/otel-arrow/collector/exporter/otelarrowexporter v0.20.0 // indirect
	github.com/open-telemetry/otel-arrow/collector/processor/concurrentbatchprocessor v0.20.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/shirou/gopsutil/v3 v3.24.2 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/collector v0.97.0 // indirect
	go.opentelemetry.io/collector/component v0.97.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.97.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.4.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.97.0 // indirect
	go.opentelemetry.io/collector/config/confignet v0.97.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.4.0 // indirect
	go.opentelemetry.io/collector/config/configretry v0.97.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.97.0 // indirect
	go.opentelemetry.io/collector/config/configtls v0.97.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.97.0 // indirect
	go.opentelemetry.io/collector/confmap v0.97.0 // indirect
	go.opentelemetry.io/collector/consumer v0.97.0 // indirect
	go.opentelemetry.io/collector/exporter v0.97.0 // indirect
	go.opentelemetry.io/collector/extension v0.97.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.97.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.4.0 // indirect
	go.opentelemetry.io/collector/pdata v1.4.0 // indirect
	go.opentelemetry.io/collector/processor v0.97.0 // indirect
	go.opentelemetry.io/collector/receiver v0.97.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.49.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.21.1 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.20.0 // indirect
	go.opentelemetry.io/proto/otlp v1.1.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.15.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240123012728-ef4313101c80 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240221002015-b0ce06bbee7c // indirect
	google.golang.org/grpc v1.62.1 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/lightstep/otel-launcher-go/pipelines => ./pipelines

replace github.com/lightstep/otel-launcher-go/lightstep/sdk/metric => ./lightstep/sdk/metric

replace github.com/lightstep/otel-launcher-go/lightstep/sdk/trace => ./lightstep/sdk/trace

replace github.com/lightstep/otel-launcher-go/lightstep/sdk/internal => ./lightstep/sdk/internal

replace github.com/lightstep/otel-launcher-go/lightstep/instrumentation => ./lightstep/instrumentation

// The 1.10.0 release included an unneccessary breaking change of
// default temporality preference; use 1.10.1 instead or consider
// upgrading to a 2.x release.
retract v1.10.0

// 1.18.2 has a broken lightstep/sdk/internal dependency.
retract v1.18.2

// ambiguous import: found package cloud.google.com/go/compute/metadata in multiple modules
replace cloud.google.com/go => cloud.google.com/go v0.110.2
