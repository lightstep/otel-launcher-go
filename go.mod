module github.com/lightstep/otel-launcher-go

go 1.23.0

toolchain go1.24.2

require (
<<<<<<< Updated upstream
	github.com/client9/misspell v0.3.4
	github.com/dgryski/go-farm v0.0.0-20240924180020-3414d57e47da
	github.com/go-logr/logr v1.4.2
	github.com/go-logr/stdr v1.2.2
	github.com/golang/mock v1.6.0
	github.com/golangci/golangci-lint v1.64.8
	github.com/google/go-cmp v0.7.0
	github.com/itchyny/gojq v0.12.17
	github.com/lightstep/go-expohisto v1.0.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/otelarrowexporter v0.125.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter v0.126.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry v0.126.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/otelarrowreceiver v0.125.0
	github.com/open-telemetry/otel-arrow/collector/processor/concurrentbatchprocessor v0.35.0
||||||| Stash base
	github.com/client9/misspell v0.3.4
	github.com/dgryski/go-farm v0.0.0-20240924180020-3414d57e47da
	github.com/go-logr/logr v1.4.3
	github.com/go-logr/stdr v1.2.2
	github.com/golang/mock v1.6.0
	github.com/golangci/golangci-lint v1.64.8
	github.com/google/go-cmp v0.7.0
	github.com/itchyny/gojq v0.12.17
	github.com/lightstep/go-expohisto v1.0.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/otelarrowexporter v0.127.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter v0.127.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry v0.127.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/otelarrowreceiver v0.127.0
	github.com/open-telemetry/otel-arrow/collector/processor/concurrentbatchprocessor v0.35.0
=======
	github.com/lightstep/otel-launcher-go/pipelines v1.34.0
>>>>>>> Stashed changes
	github.com/sethvargo/go-envconfig v1.3.0
<<<<<<< Updated upstream
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/stretchr/testify v1.10.0
	go.opentelemetry.io/collector/component v1.32.0
	go.opentelemetry.io/collector/component/componenttest v0.126.0
	go.opentelemetry.io/collector/config/configcompression v1.32.0
	go.opentelemetry.io/collector/config/configgrpc v0.126.0
	go.opentelemetry.io/collector/config/confighttp v0.126.0
	go.opentelemetry.io/collector/config/confignet v1.32.0
	go.opentelemetry.io/collector/config/configopaque v1.32.0
	go.opentelemetry.io/collector/config/configretry v1.32.0
	go.opentelemetry.io/collector/config/configtls v1.32.0
	go.opentelemetry.io/collector/consumer/consumertest v0.126.0
	go.opentelemetry.io/collector/exporter v0.126.0
	go.opentelemetry.io/collector/pdata v1.32.0
	go.opentelemetry.io/collector/processor v1.31.0
	go.opentelemetry.io/collector/receiver v1.32.0
	go.opentelemetry.io/collector/receiver/receivertest v0.126.0
	go.opentelemetry.io/contrib/propagators/b3 v1.35.0
	go.opentelemetry.io/contrib/propagators/ot v1.35.0
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/metric v1.35.0
	go.opentelemetry.io/otel/sdk v1.35.0
	go.opentelemetry.io/otel/sdk/metric v1.35.0
	go.opentelemetry.io/otel/trace v1.35.0
	go.opentelemetry.io/proto/otlp v1.3.1
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/sync v0.14.0
	golang.org/x/tools v0.33.0
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6
||||||| Stash base
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/stretchr/testify v1.10.0
	go.opentelemetry.io/collector/component v1.33.0
	go.opentelemetry.io/collector/component/componenttest v0.127.0
	go.opentelemetry.io/collector/config/configcompression v1.33.0
	go.opentelemetry.io/collector/config/configgrpc v0.127.0
	go.opentelemetry.io/collector/config/confighttp v0.127.0
	go.opentelemetry.io/collector/config/confignet v1.33.0
	go.opentelemetry.io/collector/config/configopaque v1.33.0
	go.opentelemetry.io/collector/config/configretry v1.33.0
	go.opentelemetry.io/collector/config/configtls v1.33.0
	go.opentelemetry.io/collector/consumer/consumertest v0.127.0
	go.opentelemetry.io/collector/exporter v0.127.0
	go.opentelemetry.io/collector/pdata v1.33.0
	go.opentelemetry.io/collector/processor v1.33.0
	go.opentelemetry.io/collector/receiver v1.33.0
	go.opentelemetry.io/collector/receiver/receivertest v0.127.0
	go.opentelemetry.io/contrib/propagators/b3 v1.36.0
	go.opentelemetry.io/contrib/propagators/ot v1.36.0
	go.opentelemetry.io/otel v1.36.0
	go.opentelemetry.io/otel/metric v1.36.0
	go.opentelemetry.io/otel/sdk v1.36.0
	go.opentelemetry.io/otel/sdk/metric v1.36.0
	go.opentelemetry.io/otel/trace v1.36.0
	go.opentelemetry.io/proto/otlp v1.7.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/sync v0.14.0
	golang.org/x/tools v0.33.0
	google.golang.org/grpc v1.72.2
	google.golang.org/protobuf v1.36.6
=======
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/metric v1.38.0
	go.opentelemetry.io/otel/sdk v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
>>>>>>> Stashed changes
)

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/apache/arrow-go/v18 v18.3.0 // indirect
	github.com/apache/arrow/go/v16 v16.1.0 // indirect
<<<<<<< Updated upstream
<<<<<<< HEAD
||||||| e0e8b9b
	github.com/apache/arrow/go/v17 v17.0.0 // indirect
=======
	github.com/ashanbrown/forbidigo v1.6.0 // indirect
	github.com/ashanbrown/makezero v1.2.0 // indirect
>>>>>>> 8b62b34bd9d1875086ac38cf2c7fbeab0a4452cf
	github.com/axiomhq/hyperloglog v0.2.5 // indirect
<<<<<<< HEAD
||||||| e0e8b9b
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
=======
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bkielbasa/cyclop v1.2.3 // indirect
	github.com/blizzy78/varnamelen v0.8.0 // indirect
	github.com/bombsimon/wsl/v4 v4.5.0 // indirect
	github.com/breml/bidichk v0.3.2 // indirect
	github.com/breml/errchkjson v0.4.0 // indirect
	github.com/butuzov/ireturn v0.3.1 // indirect
	github.com/butuzov/mirror v1.3.0 // indirect
	github.com/catenacyber/perfsprint v0.8.2 // indirect
	github.com/ccojocar/zxcvbn-go v1.0.2 // indirect
>>>>>>> 8b62b34bd9d1875086ac38cf2c7fbeab0a4452cf
||||||| Stash base
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bkielbasa/cyclop v1.2.3 // indirect
	github.com/blizzy78/varnamelen v0.8.0 // indirect
	github.com/bombsimon/wsl/v4 v4.5.0 // indirect
	github.com/breml/bidichk v0.3.2 // indirect
	github.com/breml/errchkjson v0.4.0 // indirect
	github.com/butuzov/ireturn v0.3.1 // indirect
	github.com/butuzov/mirror v1.3.0 // indirect
	github.com/catenacyber/perfsprint v0.8.2 // indirect
	github.com/ccojocar/zxcvbn-go v1.0.2 // indirect
=======
	github.com/axiomhq/hyperloglog v0.0.0-20230201085229-3ddf4bad03dc // indirect
>>>>>>> Stashed changes
	github.com/cenkalti/backoff/v5 v5.0.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-farm v0.0.0-20240924180020-3414d57e47da // indirect
	github.com/dgryski/go-metro v0.0.0-20250106013310-edb8663e5e33 // indirect
<<<<<<< Updated upstream
	github.com/ettle/strcase v0.2.0 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/firefart/nonamedreturns v1.0.5 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20250323135004-b31fac66206e // indirect
||||||| Stash base
	github.com/ettle/strcase v0.2.0 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/firefart/nonamedreturns v1.0.5 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20250520203025-c3c3a4ec1653 // indirect
=======
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20250520203025-c3c3a4ec1653 // indirect
>>>>>>> Stashed changes
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.8.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/flatbuffers v25.2.10+incompatible // indirect
	github.com/google/go-tpm v0.9.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
<<<<<<< Updated upstream
	github.com/gordonklaus/ineffassign v0.1.0 // indirect
	github.com/gostaticanalysis/analysisutil v0.7.1 // indirect
	github.com/gostaticanalysis/comment v1.5.0 // indirect
	github.com/gostaticanalysis/forcetypeassert v0.2.0 // indirect
	github.com/gostaticanalysis/nilerr v0.1.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/hashicorp/go-immutable-radix/v2 v2.1.0 // indirect
||||||| Stash base
	github.com/gordonklaus/ineffassign v0.1.0 // indirect
	github.com/gostaticanalysis/analysisutil v0.7.1 // indirect
	github.com/gostaticanalysis/comment v1.5.0 // indirect
	github.com/gostaticanalysis/forcetypeassert v0.2.0 // indirect
	github.com/gostaticanalysis/nilerr v0.1.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/hashicorp/go-immutable-radix/v2 v2.1.0 // indirect
=======
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
>>>>>>> Stashed changes
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
	github.com/lufia/plan9stats v0.0.0-20250317134145-8bc96cf8fc35 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
<<<<<<< Updated upstream
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nakabonne/nestif v0.3.1 // indirect
	github.com/nishanths/exhaustive v0.12.0 // indirect
	github.com/nishanths/predeclared v0.2.2 // indirect
	github.com/nunnatsa/ginkgolinter v0.19.1 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/grpcutil v0.125.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/otelarrow v0.125.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.125.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.126.0 // indirect
||||||| Stash base
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nakabonne/nestif v0.3.1 // indirect
	github.com/nishanths/exhaustive v0.12.0 // indirect
	github.com/nishanths/predeclared v0.2.2 // indirect
	github.com/nunnatsa/ginkgolinter v0.19.1 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/grpcutil v0.127.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/otelarrow v0.127.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.127.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.127.0 // indirect
=======
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/otelarrowexporter v0.127.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/grpcutil v0.127.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/otelarrow v0.127.0 // indirect
>>>>>>> Stashed changes
	github.com/open-telemetry/otel-arrow v0.35.0 // indirect
	github.com/open-telemetry/otel-arrow/collector/processor/concurrentbatchprocessor v0.35.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
<<<<<<< Updated upstream
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/quasilyte/go-ruleguard v0.4.3-0.20240823090925-0fe6f58b47b1 // indirect
	github.com/quasilyte/go-ruleguard/dsl v0.3.22 // indirect
	github.com/quasilyte/gogrep v0.5.0 // indirect
	github.com/quasilyte/regex/syntax v0.0.0-20210819130434-b3f0c404a727 // indirect
	github.com/quasilyte/stdinfo v0.0.0-20220114132959-f7386bf02567 // indirect
	github.com/raeperd/recvcheck v0.2.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
||||||| Stash base
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.64.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/quasilyte/go-ruleguard v0.4.3-0.20240823090925-0fe6f58b47b1 // indirect
	github.com/quasilyte/go-ruleguard/dsl v0.3.22 // indirect
	github.com/quasilyte/gogrep v0.5.0 // indirect
	github.com/quasilyte/regex/syntax v0.0.0-20210819130434-b3f0c404a727 // indirect
	github.com/quasilyte/stdinfo v0.0.0-20220114132959-f7386bf02567 // indirect
	github.com/raeperd/recvcheck v0.2.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
=======
>>>>>>> Stashed changes
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/shirou/gopsutil/v3 v3.24.5 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
<<<<<<< Updated upstream
<<<<<<< HEAD
	go.opentelemetry.io/collector/client v1.31.0 // indirect
	go.opentelemetry.io/collector/component v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v0.125.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.31.0 // indirect
	go.opentelemetry.io/collector/confmap v1.31.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.125.0 // indirect
	go.opentelemetry.io/collector/consumer v1.31.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.125.0 // indirect
	go.opentelemetry.io/collector/exporter v0.125.0 // indirect
	go.opentelemetry.io/collector/extension v1.31.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.31.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.125.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.125.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.31.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.125.0 // indirect
	go.opentelemetry.io/collector/pdata v1.31.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.125.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.125.0 // indirect
	go.opentelemetry.io/collector/processor v1.31.0 // indirect
	go.opentelemetry.io/collector/receiver v1.31.0 // indirect
||||||| e0e8b9b
	go.opentelemetry.io/collector/client v1.31.0 // indirect
	go.opentelemetry.io/collector/component v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v0.125.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.31.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.116.0 // indirect
	go.opentelemetry.io/collector/confmap v1.31.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.125.0 // indirect
	go.opentelemetry.io/collector/consumer v1.31.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.125.0 // indirect
	go.opentelemetry.io/collector/exporter v0.125.0 // indirect
	go.opentelemetry.io/collector/extension v1.31.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.121.0 // indirect
	go.opentelemetry.io/collector/extension/experimental/storage v0.117.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.31.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.125.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.125.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.31.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.125.0 // indirect
	go.opentelemetry.io/collector/pdata v1.31.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.125.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.125.0 // indirect
	go.opentelemetry.io/collector/processor v1.31.0 // indirect
	go.opentelemetry.io/collector/receiver v1.31.0 // indirect
=======
	go.opentelemetry.io/collector/client v1.32.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.126.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.126.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v0.126.0 // indirect
	go.opentelemetry.io/collector/confmap v1.32.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.126.0 // indirect
	go.opentelemetry.io/collector/consumer v1.32.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.126.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.126.0 // indirect
	go.opentelemetry.io/collector/extension v1.32.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.32.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.126.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.126.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.32.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.126.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.126.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.126.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.126.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.126.0 // indirect
>>>>>>> 8b62b34bd9d1875086ac38cf2c7fbeab0a4452cf
	go.opentelemetry.io/contrib/bridges/otelzap v0.10.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.60.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0 // indirect
	go.opentelemetry.io/otel/log v0.11.0 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
||||||| Stash base
	go.opentelemetry.io/collector/client v1.31.0 // indirect
	go.opentelemetry.io/collector/client v1.31.0 // indirect
	go.opentelemetry.io/collector/client v1.32.0 // indirect
	go.opentelemetry.io/collector/client v1.32.0 // indirect
	go.opentelemetry.io/collector/client v1.33.0 // indirect
	go.opentelemetry.io/collector/component v1.31.0 // indirect
	go.opentelemetry.io/collector/component v1.31.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.126.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.126.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.127.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.126.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.126.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.127.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v0.126.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v0.126.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v0.127.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.31.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.125.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.31.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.31.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.116.0 // indirect
	go.opentelemetry.io/collector/confmap v1.31.0 // indirect
	go.opentelemetry.io/collector/confmap v1.31.0 // indirect
	go.opentelemetry.io/collector/confmap v1.32.0 // indirect
	go.opentelemetry.io/collector/confmap v1.32.0 // indirect
	go.opentelemetry.io/collector/confmap v1.33.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.125.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.125.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.126.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.126.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.127.0 // indirect
	go.opentelemetry.io/collector/consumer v1.31.0 // indirect
	go.opentelemetry.io/collector/consumer v1.31.0 // indirect
	go.opentelemetry.io/collector/consumer v1.32.0 // indirect
	go.opentelemetry.io/collector/consumer v1.32.0 // indirect
	go.opentelemetry.io/collector/consumer v1.33.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.125.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.125.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.126.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.126.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.127.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.126.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.126.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.127.0 // indirect
	go.opentelemetry.io/collector/exporter v0.125.0 // indirect
	go.opentelemetry.io/collector/exporter v0.125.0 // indirect
	go.opentelemetry.io/collector/extension v1.31.0 // indirect
	go.opentelemetry.io/collector/extension v1.31.0 // indirect
	go.opentelemetry.io/collector/extension v1.32.0 // indirect
	go.opentelemetry.io/collector/extension v1.32.0 // indirect
	go.opentelemetry.io/collector/extension v1.33.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.121.0 // indirect
	go.opentelemetry.io/collector/extension/experimental/storage v0.117.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.31.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.31.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.32.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.32.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.33.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.125.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.125.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.126.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.126.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.127.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.125.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.125.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.126.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.126.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.127.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.31.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.31.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.32.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.32.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.33.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.125.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.125.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.126.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.126.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.127.0 // indirect
	go.opentelemetry.io/collector/pdata v1.31.0 // indirect
	go.opentelemetry.io/collector/pdata v1.31.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.125.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.125.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.126.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.126.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.127.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.125.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.125.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.126.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.126.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.127.0 // indirect
	go.opentelemetry.io/collector/processor v1.31.0 // indirect
	go.opentelemetry.io/collector/processor v1.31.0 // indirect
	go.opentelemetry.io/collector/receiver v1.31.0 // indirect
	go.opentelemetry.io/collector/receiver v1.31.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.126.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.126.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.127.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.126.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.126.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.127.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.10.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.10.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.11.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.60.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.60.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/otel/log v0.11.0 // indirect
	go.opentelemetry.io/otel/log v0.11.0 // indirect
	go.opentelemetry.io/otel/log v0.12.2 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
=======
	go.opentelemetry.io/collector/client v1.33.0 // indirect
	go.opentelemetry.io/collector/component v1.33.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.127.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.33.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.127.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v0.127.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.33.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.33.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.33.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.127.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.33.0 // indirect
	go.opentelemetry.io/collector/confmap v1.33.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.127.0 // indirect
	go.opentelemetry.io/collector/consumer v1.33.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.127.0 // indirect
	go.opentelemetry.io/collector/exporter v0.127.0 // indirect
	go.opentelemetry.io/collector/extension v1.33.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.33.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.127.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.127.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.33.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.127.0 // indirect
	go.opentelemetry.io/collector/pdata v1.33.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.127.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.127.0 // indirect
	go.opentelemetry.io/collector/processor v1.33.0 // indirect
	go.opentelemetry.io/collector/receiver v1.33.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.11.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.36.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.36.0 // indirect
	go.opentelemetry.io/otel/log v0.12.2 // indirect
	go.opentelemetry.io/proto/otlp v1.7.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
>>>>>>> Stashed changes
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/exp v0.0.0-20250506013437-ce4c2cf36ca6 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
<<<<<<< Updated upstream
	google.golang.org/genproto/googleapis/api v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250505200425-f936aa4a68b2 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
||||||| Stash base
	google.golang.org/genproto/googleapis/api v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
=======
	google.golang.org/genproto/googleapis/api v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/grpc v1.72.2 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
>>>>>>> Stashed changes
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
