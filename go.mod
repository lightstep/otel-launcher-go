module github.com/lightstep/otel-launcher-go

go 1.23.0

toolchain go1.24.2

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
	github.com/apache/arrow-go/v18 v18.3.0 // indirect
	github.com/apache/arrow/go/v16 v16.1.0 // indirect
	github.com/axiomhq/hyperloglog v0.0.0-20230201085229-3ddf4bad03dc // indirect
	github.com/cenkalti/backoff/v5 v5.0.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-farm v0.0.0-20240924180020-3414d57e47da // indirect
	github.com/dgryski/go-metro v0.0.0-20180109044635-280f6062b5bc // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20250323135004-b31fac66206e // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/google/go-tpm v0.9.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.2.0 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/lightstep/otel-launcher-go/lightstep/instrumentation v1.34.0 // indirect
	github.com/lightstep/otel-launcher-go/lightstep/sdk/internal v1.34.0 // indirect
	github.com/lightstep/otel-launcher-go/lightstep/sdk/metric v1.34.0 // indirect
	github.com/lightstep/otel-launcher-go/lightstep/sdk/trace v1.31.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/otelarrowexporter v0.125.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/grpcutil v0.125.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/otelarrow v0.125.0 // indirect
	github.com/open-telemetry/otel-arrow v0.35.0 // indirect
	github.com/open-telemetry/otel-arrow/collector/processor/concurrentbatchprocessor v0.35.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/shirou/gopsutil/v3 v3.24.5 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector/client v1.32.0 // indirect
	go.opentelemetry.io/collector/component v1.32.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.126.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.126.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.32.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.126.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v0.126.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.32.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.32.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.32.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.32.0 // indirect
	go.opentelemetry.io/collector/confmap v1.32.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.126.0 // indirect
	go.opentelemetry.io/collector/consumer v1.32.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.126.0 // indirect
	go.opentelemetry.io/collector/exporter v0.126.0 // indirect
	go.opentelemetry.io/collector/extension v1.32.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.32.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.126.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.126.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.32.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.126.0 // indirect
	go.opentelemetry.io/collector/pdata v1.32.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.126.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.126.0 // indirect
	go.opentelemetry.io/collector/processor v1.31.0 // indirect
	go.opentelemetry.io/collector/receiver v1.32.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.126.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.10.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.60.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.35.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.35.0 // indirect
	go.opentelemetry.io/otel/log v0.11.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/exp v0.0.0-20240909161429-701f63a606c0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/grpc v1.72.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

// The 1.10.0 release included an unneccessary breaking change of
// default temporality preference; use 1.10.1 instead or consider
// upgrading to a 2.x release.
retract v1.10.0

// 1.18.2 has a broken lightstep/sdk/internal dependency.
retract v1.18.2

// ambiguous import: found package cloud.google.com/go/compute/metadata in multiple modules
replace cloud.google.com/go => cloud.google.com/go v0.110.2

tool golang.org/x/tools/cmd/stringer
