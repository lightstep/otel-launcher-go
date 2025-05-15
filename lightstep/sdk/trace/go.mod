module github.com/lightstep/otel-launcher-go/lightstep/sdk/trace

go 1.23.0

toolchain go1.24.2

require (
	github.com/google/go-cmp v0.7.0
	github.com/lightstep/otel-launcher-go/lightstep/sdk/internal v1.34.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/otelarrowexporter v0.125.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/otelarrowreceiver v0.125.0
	github.com/open-telemetry/otel-arrow/collector/processor/concurrentbatchprocessor v0.35.0
	github.com/stretchr/testify v1.10.0
	go.opentelemetry.io/collector/component v1.31.0
	go.opentelemetry.io/collector/component/componenttest v0.125.0
	go.opentelemetry.io/collector/config/configcompression v1.31.0
	go.opentelemetry.io/collector/config/configgrpc v0.125.0
	go.opentelemetry.io/collector/config/confignet v1.31.0
	go.opentelemetry.io/collector/config/configopaque v1.31.0
	go.opentelemetry.io/collector/config/configretry v1.31.0
	go.opentelemetry.io/collector/config/configtls v1.31.0
	go.opentelemetry.io/collector/consumer/consumertest v0.125.0
	go.opentelemetry.io/collector/exporter v0.125.0
	go.opentelemetry.io/collector/pdata v1.31.0
	go.opentelemetry.io/collector/processor v1.31.0
	go.opentelemetry.io/collector/receiver v1.31.0
	go.opentelemetry.io/collector/receiver/receivertest v0.125.0
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/metric v1.35.0
	go.opentelemetry.io/otel/sdk v1.35.0
	go.opentelemetry.io/otel/trace v1.35.0
	go.opentelemetry.io/proto/otlp v1.3.1
	go.uber.org/multierr v1.11.0
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/apache/arrow-go/v18 v18.3.0 // indirect
	github.com/apache/arrow/go/v16 v16.1.0 // indirect
	github.com/axiomhq/hyperloglog v0.2.5 // indirect
	github.com/cenkalti/backoff/v5 v5.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-metro v0.0.0-20250106013310-edb8663e5e33 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.8.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kamstrup/intmap v0.5.1 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.2.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/grpcutil v0.125.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/otelarrow v0.125.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.125.0 // indirect
	github.com/open-telemetry/otel-arrow v0.35.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector/client v1.31.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v0.125.0 // indirect
	go.opentelemetry.io/collector/confmap v1.31.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.125.0 // indirect
	go.opentelemetry.io/collector/consumer v1.31.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.125.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.125.0 // indirect
	go.opentelemetry.io/collector/extension v1.31.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.31.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.125.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.125.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.31.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.125.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.125.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.125.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.125.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.125.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.10.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.60.0 // indirect
	go.opentelemetry.io/otel/log v0.11.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.35.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/exp v0.0.0-20250506013437-ce4c2cf36ca6 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250505200425-f936aa4a68b2 // indirect
	google.golang.org/grpc v1.72.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace github.com/lightstep/otel-launcher-go/lightstep/sdk/internal => ../internal

// ambiguous import: found package cloud.google.com/go/compute/metadata in multiple modules
replace cloud.google.com/go => cloud.google.com/go v0.110.2
