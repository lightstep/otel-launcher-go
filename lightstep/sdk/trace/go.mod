module github.com/lightstep/otel-launcher-go/lightstep/sdk/trace

<<<<<<< HEAD
go 1.24.0
||||||| parent of ab5b855 (Update all deps (take 2))
go 1.22.0
=======
go 1.23.0
>>>>>>> ab5b855 (Update all deps (take 2))

toolchain go1.24.2

require (
	github.com/google/go-cmp v0.7.0
	github.com/lightstep/otel-launcher-go/lightstep/sdk/internal v1.34.0
<<<<<<< HEAD
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/otelarrowexporter v0.135.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/otelarrowreceiver v0.135.0
	github.com/open-telemetry/otel-arrow/collector/processor/concurrentbatchprocessor v0.35.0
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/collector/component v1.41.0
	go.opentelemetry.io/collector/component/componenttest v0.135.0
	go.opentelemetry.io/collector/config/configcompression v1.41.0
	go.opentelemetry.io/collector/config/configgrpc v0.135.0
	go.opentelemetry.io/collector/config/confignet v1.41.0
	go.opentelemetry.io/collector/config/configopaque v1.41.0
	go.opentelemetry.io/collector/config/configretry v1.41.0
	go.opentelemetry.io/collector/config/configtls v1.41.0
	go.opentelemetry.io/collector/consumer/consumertest v0.135.0
	go.opentelemetry.io/collector/exporter v0.135.0
	go.opentelemetry.io/collector/exporter/exporterhelper v0.135.0
	go.opentelemetry.io/collector/pdata v1.41.0
	go.opentelemetry.io/collector/processor v1.41.0
	go.opentelemetry.io/collector/receiver v1.41.0
	go.opentelemetry.io/collector/receiver/receivertest v0.135.0
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/metric v1.38.0
	go.opentelemetry.io/otel/sdk v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
||||||| parent of ab5b855 (Update all deps (take 2))
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/otelarrowexporter v0.114.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/otelarrowreceiver v0.114.0
	github.com/open-telemetry/otel-arrow/collector/processor/concurrentbatchprocessor v0.31.0
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/collector/component v0.114.0
	go.opentelemetry.io/collector/component/componenttest v0.114.0
	go.opentelemetry.io/collector/config/configcompression v1.20.0
	go.opentelemetry.io/collector/config/configgrpc v0.114.0
	go.opentelemetry.io/collector/config/confignet v1.20.0
	go.opentelemetry.io/collector/config/configopaque v1.20.0
	go.opentelemetry.io/collector/config/configretry v1.20.0
	go.opentelemetry.io/collector/config/configtls v1.20.0
	go.opentelemetry.io/collector/consumer/consumertest v0.114.0
	go.opentelemetry.io/collector/exporter v0.114.0
	go.opentelemetry.io/collector/pdata v1.20.0
	go.opentelemetry.io/collector/processor v0.114.0
	go.opentelemetry.io/collector/receiver v0.114.0
	go.opentelemetry.io/collector/receiver/receivertest v0.114.0
	go.opentelemetry.io/otel v1.32.0
	go.opentelemetry.io/otel/metric v1.32.0
	go.opentelemetry.io/otel/sdk v1.32.0
	go.opentelemetry.io/otel/trace v1.32.0
=======
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
>>>>>>> ab5b855 (Update all deps (take 2))
	go.opentelemetry.io/proto/otlp v1.3.1
	go.uber.org/multierr v1.11.0
<<<<<<< HEAD
	google.golang.org/protobuf v1.36.9
||||||| parent of ab5b855 (Update all deps (take 2))
	google.golang.org/protobuf v1.35.1
=======
	google.golang.org/protobuf v1.36.6
>>>>>>> ab5b855 (Update all deps (take 2))
)

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
<<<<<<< HEAD
	github.com/apache/arrow-go/v18 v18.4.1 // indirect
	github.com/axiomhq/hyperloglog v0.2.5 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-metro v0.0.0-20250106013310-edb8663e5e33 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20250903184740-5d135037bd4d // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
||||||| parent of ab5b855 (Update all deps (take 2))
	github.com/apache/arrow/go/v16 v16.1.0 // indirect
	github.com/apache/arrow/go/v17 v17.0.0 // indirect
	github.com/axiomhq/hyperloglog v0.0.0-20230201085229-3ddf4bad03dc // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-metro v0.0.0-20180109044635-280f6062b5bc // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
=======
	github.com/apache/arrow-go/v18 v18.3.0 // indirect
	github.com/apache/arrow/go/v16 v16.1.0 // indirect
	github.com/axiomhq/hyperloglog v0.2.5 // indirect
	github.com/cenkalti/backoff/v5 v5.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-metro v0.0.0-20250106013310-edb8663e5e33 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.8.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
>>>>>>> ab5b855 (Update all deps (take 2))
	github.com/go-logr/stdr v1.2.2 // indirect
<<<<<<< HEAD
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
||||||| parent of ab5b855 (Update all deps (take 2))
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
=======
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
>>>>>>> ab5b855 (Update all deps (take 2))
	github.com/gogo/protobuf v1.3.2 // indirect
<<<<<<< HEAD
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/google/go-tpm v0.9.5 // indirect
||||||| parent of ab5b855 (Update all deps (take 2))
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
=======
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
>>>>>>> ab5b855 (Update all deps (take 2))
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
<<<<<<< HEAD
	github.com/kamstrup/intmap v0.5.1 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.2.2 // indirect
||||||| parent of ab5b855 (Update all deps (take 2))
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.2 // indirect
=======
	github.com/kamstrup/intmap v0.5.1 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.2.0 // indirect
>>>>>>> ab5b855 (Update all deps (take 2))
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
<<<<<<< HEAD
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/grpcutil v0.135.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/otelarrow v0.135.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.135.0 // indirect
	github.com/open-telemetry/otel-arrow/go v0.42.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
||||||| parent of ab5b855 (Update all deps (take 2))
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/grpcutil v0.114.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/otelarrow v0.114.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.114.0 // indirect
	github.com/open-telemetry/otel-arrow v0.30.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
=======
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/grpcutil v0.125.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/otelarrow v0.125.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.125.0 // indirect
	github.com/open-telemetry/otel-arrow v0.35.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
>>>>>>> ab5b855 (Update all deps (take 2))
	github.com/x448/float16 v0.8.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
<<<<<<< HEAD
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector/client v1.41.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.135.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.135.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v1.41.0 // indirect
	go.opentelemetry.io/collector/config/configoptional v0.135.0 // indirect
	go.opentelemetry.io/collector/confmap v1.41.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.135.0 // indirect
	go.opentelemetry.io/collector/consumer v1.41.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.135.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.135.0 // indirect
	go.opentelemetry.io/collector/extension v1.41.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.41.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.135.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.135.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.41.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.135.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.135.0 // indirect
	go.opentelemetry.io/collector/pdata/xpdata v0.135.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.41.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.135.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.135.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.13.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0 // indirect
	go.opentelemetry.io/otel/log v0.14.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
||||||| parent of ab5b855 (Update all deps (take 2))
	go.opentelemetry.io/collector/client v1.20.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.114.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.114.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.114.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.114.0 // indirect
	go.opentelemetry.io/collector/confmap v1.20.0 // indirect
	go.opentelemetry.io/collector/consumer v0.114.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.114.0 // indirect
	go.opentelemetry.io/collector/consumer/consumerprofiles v0.114.0 // indirect
	go.opentelemetry.io/collector/extension v0.114.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.114.0 // indirect
	go.opentelemetry.io/collector/extension/experimental/storage v0.114.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.114.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.114.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverprofiles v0.114.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.56.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.32.0 // indirect
=======
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
>>>>>>> ab5b855 (Update all deps (take 2))
	go.uber.org/zap v1.27.0 // indirect
<<<<<<< HEAD
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/exp v0.0.0-20250819193227-8b4c13bb791b // indirect
	golang.org/x/mod v0.28.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250908214217-97024824d090 // indirect
	google.golang.org/grpc v1.75.0 // indirect
||||||| parent of ab5b855 (Update all deps (take 2))
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/mod v0.18.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241007155032-5fefd90f89a9 // indirect
	google.golang.org/grpc v1.67.1 // indirect
=======
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
>>>>>>> ab5b855 (Update all deps (take 2))
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace github.com/lightstep/otel-launcher-go/lightstep/sdk/internal => ../internal

// ambiguous import: found package cloud.google.com/go/compute/metadata in multiple modules
replace cloud.google.com/go => cloud.google.com/go v0.110.2
