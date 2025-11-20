module github.com/lightstep/otel-launcher-go

go 1.24.0

toolchain go1.24.7

require (
	github.com/lightstep/otel-launcher-go/pipelines v1.34.0
	github.com/sethvargo/go-envconfig v1.3.0
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/metric v1.38.0
	go.opentelemetry.io/otel/sdk v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
)

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/apache/arrow-go/v18 v18.4.1 // indirect
	github.com/axiomhq/hyperloglog v0.2.5 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-farm v0.0.0-20240924180020-3414d57e47da // indirect
	github.com/dgryski/go-metro v0.0.0-20250106013310-edb8663e5e33 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20250903184740-5d135037bd4d // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/google/go-tpm v0.9.5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kamstrup/intmap v0.5.1 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.2.2 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/lightstep/otel-launcher-go/lightstep/instrumentation v1.34.0 // indirect
	github.com/lightstep/otel-launcher-go/lightstep/sdk/internal v1.34.0 // indirect
	github.com/lightstep/otel-launcher-go/lightstep/sdk/metric v1.34.0 // indirect
	github.com/lightstep/otel-launcher-go/lightstep/sdk/trace v1.31.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20250827001030-24949be3fa54 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/otelarrowexporter v0.135.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/grpcutil v0.135.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/otelarrow v0.135.0 // indirect
	github.com/open-telemetry/otel-arrow/collector/processor/concurrentbatchprocessor v0.35.0 // indirect
	github.com/open-telemetry/otel-arrow/go v0.42.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/shirou/gopsutil/v3 v3.24.5 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector/client v1.41.0 // indirect
	go.opentelemetry.io/collector/component v1.41.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.135.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.41.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.135.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v1.41.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.41.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.41.0 // indirect
	go.opentelemetry.io/collector/config/configoptional v0.135.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.41.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.41.0 // indirect
	go.opentelemetry.io/collector/confmap v1.41.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.135.0 // indirect
	go.opentelemetry.io/collector/consumer v1.41.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.135.0 // indirect
	go.opentelemetry.io/collector/exporter v0.135.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper v0.135.0 // indirect
	go.opentelemetry.io/collector/extension v1.41.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.41.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.135.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.135.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.41.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.135.0 // indirect
	go.opentelemetry.io/collector/pdata v1.41.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.135.0 // indirect
	go.opentelemetry.io/collector/pdata/xpdata v0.135.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.41.0 // indirect
	go.opentelemetry.io/collector/processor v1.41.0 // indirect
	go.opentelemetry.io/collector/receiver v1.41.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.13.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.35.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.35.0 // indirect
	go.opentelemetry.io/otel/log v0.14.0 // indirect
	go.opentelemetry.io/proto/otlp v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/exp v0.0.0-20250819193227-8b4c13bb791b // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/telemetry v0.0.0-20251008203120-078029d740a8 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250908214217-97024824d090 // indirect
	google.golang.org/grpc v1.75.0 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// The 1.10.0 release included an unneccessary breaking change of
// default temporality preference; use 1.10.1 instead or consider
// upgrading to a 2.x release.
retract v1.10.0

// 1.18.2 has a broken lightstep/sdk/internal dependency.
retract v1.18.2

// ambiguous import: found package cloud.google.com/go/compute/metadata in multiple modules
replace cloud.google.com/go => cloud.google.com/go v0.110.2

replace github.com/lightstep/otel-launcher-go/lightstep/sdk/internal => ./lightstep/sdk/internal

replace github.com/lightstep/otel-launcher-go/lightstep/sdk/metric => ./lightstep/sdk/metric

replace github.com/lightstep/otel-launcher-go/lightstep/sdk/trace => ./lightstep/sdk/trace

replace github.com/lightstep/otel-launcher-go/lightstep/instrumentation => ./lightstep/instrumentation

tool golang.org/x/tools/cmd/stringer
