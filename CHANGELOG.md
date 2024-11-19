# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

## [1.34.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.34.0) - 2024-11-19

- OTel-Collector v0.114.0 dependencies. [#814](https://github.com/lightstep/otel-launcher-go/pull/814)

## [1.33.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.33.0) - 2024-11-06

- OTel-Collector v0.113.0 dependencies. [#802](https://github.com/lightstep/otel-launcher-go/pull/802)
- Self-telemetry fixes. [#798](https://github.com/lightstep/otel-launcher-go/pull/798)

## [1.32.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.32.0) - 2024-10-15

- Add Prometheus exporter. [#775](https://github.com/lightstep/otel-launcher-go/pull/775)

## [1.31.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.31.0) - 2024-09-06

- SDK-internal metrics: replace "otelcol_" prefix with "otelsdk_" [#760](https://github.com/lightstep/otel-launcher-go/pull/760)
- Traces exporter: allow custom meter/trace providers [#739](https://github.com/lightstep/otel-launcher-go/pull/739)
- Use correct default retry settings. [#741](https://github.com/lightstep/otel-launcher-go/pull/741)

## [1.30.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.30.0) - 2024-07-17

- Update OpenTelemetry-Arrow components to latest, now in collector-contrib. [#736](https://github.com/lightstep/otel-launcher-go/pull/736)

## [1.29.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.29.0) - 2024-06-06

- Update to OpenTelemetry-Go 1.27.0, including new Synchronous Gauge support. [#713](https://github.com/lightstep/otel-launcher-go/pull/713)

## [1.28.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.28.0) - 2024-04-12

- Dependency update: Use OpenTelemetry Collector v0.98.0, OTel-Arrow v0.21.0. [#678](https://github.com/lightstep/otel-launcher-go/pull/678)

## [1.27.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.27.0) - 2024-03-27

- Dependency update: Use OpenTelemetry Collector v0.97.0, OTel-Arrow v0.20.0. [#663](https://github.com/lightstep/otel-launcher-go/pull/663)

## [1.26.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.26.0) - 2024-03-06

- Dependency update: Use OpenTelemetry Collector v0.96.0, OTel-Arrow v0.18.0. [#651](https://github.com/lightstep/otel-launcher-go/pull/651)

## [1.25.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.25.0) - 2024-02-14

- Dependency update: Use OpenTelemetry Collector v0.94.1, OTel-Arrow v0.17.0. [#634](https://github.com/lightstep/otel-launcher-go/pull/634)

## [1.24.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.24.0) - 2024-01-11

- Dependency update: Use OpenTelemetry Collector v0.92.0, OTel-Arrow v0.14.0. [#600](https://github.com/lightstep/otel-launcher-go/pull/600)
- Trace exporter: always use LS instrumentation packages and OTel Collector export path, remove dependency on OTel-Go OTLP trace exporter. [#595](https://github.com/lightstep/otel-launcher-go/pull/595)
- Metric exporter: always use LS metrics SDK and OTel Collector export path, remove dependency on OTel-Go OTLP metric exporter and OTel-Go-Contrib metric instrumentation. [#596](https://github.com/lightstep/otel-launcher-go/pull/596)
- Add "lowmemory" temporality naming, leaving "stateless" as an alias. [#597](https://github.com/lightstep/otel-launcher-go/pull/597)

## [1.23.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.23.0) - 2023-12-20

- Update to OTel-Arrow v0.13.0 dependencies. [#591](https://github.com/lightstep/otel-launcher-go/pull/591)
- LS metrics SDK: Record a span and a metric about the outcome of export. [#590](https://github.com/lightstep/otel-launcher-go/pull/590)
- LS Metrics SDK add an `AttributeSizeLimit` [#588](https://github.com/lightstep/otel-launcher-go/pull/588)
- Change the name of the otelcol batch processor to indicate whether traces or metrics. [#587](https://github.com/lightstep/otel-launcher-go/pull/587)

## [1.22.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.22.0) - 2023-11-27

- Use the OTel-Arrow concurrent batch processor in SDK exporters based on the OTel collector. [#569](https://github.com/lightstep/otel-launcher-go/pull/569)

## [1.21.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.21.0) - 2023-11-03

- Release based on OTel Collector v0.88 dependencies. [#539](https://github.com/lightstep/otel-launcher-go/pull/539)
- Release based on Go 1.21, including new runtime/metrics. [#537](https://github.com/lightstep/otel-launcher-go/pull/537)

## [1.20.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.20.0) - 2023-10-16

- Release based on OTel Collector v0.87 dependencies. [#526](https://github.com/lightstep/otel-launcher-go/pull/526)

## [1.19.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.19.0) - 2023-08-17

- Release based on new [open-telemetry/otel-arrow](https://github.com/open-telemetry/otel-arrow) repository. [#511](https://github.com/lightstep/otel-launcher-go/pull/511)

## [1.18.7](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.18.7) - 2023-08-14

- Add WithRenameFunction() to metrics SDK: a more flexible way to rename metrics when multiple instruments are matching on the same regular expression. [#504](https://github.com/lightstep/otel-launcher-go/pull/504)

## [1.18.6](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.18.6) - 2023-06-27

- Support experimental `MeasurementProcessor` in the LS metrics SDK. [#493](https://github.com/lightstep/otel-launcher-go/pull/493)

## [1.18.5](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.18.5) - 2023-06-26

- Correct the translation of Span StatusCode. [#495](https://github.com/lightstep/otel-launcher-go/pull/495)

## [1.18.4](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.18.4) - 2023-06-26

- Disable self-tracing, enable self-metrics, and make these optional behaviors. Bugfix in span event timestamps. [#494](https://github.com/lightstep/otel-launcher-go/pull/494)

## [1.18.3](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.18.3) - 2023-06-21

- Re-release. [#492](https://github.com/lightstep/otel-launcher-go/pull/492)

## [1.18.2](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.18.2) - 2023-06-21

- Support for multi-resource export pipelines as a special case. [#490](https://github.com/lightstep/otel-launcher-go/pull/490)

## [1.18.1](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.18.1) - 2023-06-16

- Modifies the exporter names used in the two OTel-Collector-based exporters. [#480](https://github.com/lightstep/otel-launcher-go/pull/480)

## [1.18.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.18.0) - 2023-06-12

- Adds support for collector based exporting (using batch processor and otlp exporter) for otel-go traces. [#464](https://github.com/lightstep/otel-launcher-go/pull/464)
- Updated dependencies on go.opentelemetry.io/otel to `v1.16.0` [#471](https://github.com/lightstep/otel-launcher-go/pull/471)

## [1.17.1](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.17.1) - 2023-06-01

- Lightstep metrics SDK would improperly set the `Insecure` TLS setting when not configured for insecure transport. [#465](https://github.com/lightstep/otel-launcher-go/pull/465)

## [1.17.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.17.0) - 2023-05-22

- Update to OTel-Go API v0.38.1. [#447](https://github.com/lightstep/otel-launcher-go/pull/447)
- Exporter now based on OTel collector batchprocessor and otlpexporter. [#445](https://github.com/lightstep/otel-launcher-go/pull/445)

## [1.16.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.16.0) - 2023-04-27

- LS Metrics SDK supports an API hint for controlling Temporality on a per-instrument basis. [#426](https://github.com/lightstep/otel-launcher-go/pull/426)

## [1.15.1](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.15.1) - 2023-04-03

- Correct race conditions related to sharing of the input attributes slice. [#422](https://github.com/lightstep/otel-launcher-go/pull/422)

## [1.15.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.15.0) - 2023-04-03

- Remove access token length check. [#412](https://github.com/lightstep/otel-launcher-go/pull/412)
- Update all OTel dependencies (removal of upstream `metric/unit` package). [#421](https://github.com/lightstep/otel-launcher-go/pull/421)

## [1.14.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.14.0) - 2023-03-17

- Support cardinality limits. Synchronous instruments support an
  instrument-level cardinality limit; Synchronous and Asynchronous
  aggregators support per-view cardinality limits. Performance settings
  determine the view-configured defaults. [#385](https://github.com/lightstep/otel-launcher-go/pull/385)

## [1.13.4](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.13.3) - 2023-03-02

- Minor performance improvement, one less allocation under the lock
  when fingerprint collisions are being checked. [#407](https://github.com/lightstep/otel-launcher-go/pull/407)

## [1.13.3](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.13.3) - 2023-03-01

- Fixes performance regressions introduced in version 1.13.1 and 1.12.1.
  [#404](https://github.com/lightstep/otel-launcher-go/pull/404)

## [1.13.2](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.13.2) - 2023-02-28

- Adds performance setting `InactiveCollectionPeriods` allowing
  control over how fast records are removed from memory. [#396](https://github.com/lightstep/otel-launcher-go/pull/396)

## [1.13.1](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.13.1) - 2023-02-15

- Adds performance improvements for concurrent use of synchronous
  instruments. [#384](https://github.com/lightstep/otel-launcher-go/pull/384)
- Adds `WithPerformance()` and `IgnoreCollisions` setting which offers
  around 10% faster operations in exchange for safety and correctness. This
  setting is off by default. [#384](https://github.com/lightstep/otel-launcher-go/pull/384)

## [1.13.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.13.0) - 2023-02-15

- Updates OTel-Go version dependencies to `trace@v1.12.0`, `metrics@v0.35.0`,
  `contrib-trace@v1.13.0`, `contrib-metrics@v0.38.0`:
  - Note the corresponding metrics API changes are 🛑 [BREAKING]
    between releases `v0.35.0` and `v0.36.0`. Because this release
    depends on metrics API `v0.35.0` it continues to support the
    deprecated APIs. The next minor version of this repository will
    update the dependency to `v0.36.0` or later. [#381](https://github.com/lightstep/otel-launcher-go/pull/381)

## [1.12.1](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.12.1) - 2023-02-13

- Replace a RWMutex with Mutex. [#378](https://github.com/lightstep/otel-launcher-go/pull/378)
- Eliminate redundant GC cpu-time metric from lightstep/instrumentation/cputime as
  it is now included in lightstep/instrumentation/runtime from the builtin
  runtime/metrics package. [#355](https://github.com/lightstep/otel-launcher-go/pull/355)

## [1.12.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.12.0) - 2022-12-21)

### Changed

- Update to OTel-Go 1.11.x and OTel-Go metrics 0.34.
  [#333](https://github.com/lightstep/otel-launcher-go/pull/333)
- Remove Lightstep-specific partial-success error handling, now using upstream.
  [#250](https://github.com/lightstep/otel-launcher-go/pull/250)
- Improved runtime/metrics instrumentation package w/ support for Go-1.20
  [#301](https://github.com/lightstep/otel-launcher-go/pull/301)

## [1.11.1](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.11.0) - 2022-10-05

### Bug fixes

- Updates required for tests to pass with go-1.19.x.
  [#300](https://github.com/lightstep/otel-launcher-go/pull/300)

### Changed

- Print the metric name responsible for empty attributes
  [#298](https://github.com/lightstep/otel-launcher-go/pull/298)

## [1.11.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.11.0) - 2022-09-12

🛑 [BREAKING] The set of builtin, automatic metrics instrumentation
libraries configured by the Launcher is new and improved in this
release. To configure the behavior in Launcher versions 1.10.x and
prior, use `WithMetricsBuiltinLibraries("all:prestable")` or set
`LS_METRICS_BUILTIN_LIBRARIES=all:prestable`.

### Added

- Add `WithMetricsBuiltinsEnabled()` option and environment variable
  `LS_METRICS_BUILTINS_ENABLED`, which defaults to true.
  [#265](https://github.com/lightstep/otel-launcher-go/pull/265)
- New replacement for go-contrib instrumentation/runtime added as lightstep/instrumentation/runtime.
  [#267](https://github.com/lightstep/otel-launcher-go/pull/267)
- New "cputime" instrumentation package combines several related timing metrics,
  process.cpu.time, process.uptime, and process.runtime.go.gc.cpu.time [#269](https://github.com/lightstep/otel-launcher-go/pull/269)
- New replacement for go-contrib instrumentation/host added as
  lightstep/instrumentation/host; same code but removes process metrics
  [#268](https://github.com/lightstep/otel-launcher-go/pull/268)
- When metrics builtin libraries are enabled, the
  `WithMetricsBuiltinLibraries()` option and environment variable
  `LS_METRICS_BUILTIN_LIBRARIES` control which libraries are enabled.
  By default, the libraries in
  [`./lightstep/instrumentation`](https://github.com/lightstep/otel-launcher-go/tree/main/lightstep/instrumentation)
  are selected as stable builtin instrumentation. To configure the
  former-default builtin libraries (i.e.,
  [runtime](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/runtime)
  and
  [host](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/host))
  use `WithMetricsBuiltinLibraries("all:prestable")` or set
  `LS_METRICS_BUILTIN_LIBRARIES=all:prestable`.
  [#274](https://github.com/lightstep/otel-launcher-go/pull/274)

### Changed

- Update to OTel-Go 1.10.0.
  [#284](https://github.com/lightstep/otel-launcher-go/pull/284)
- Avoid repetitive calls to `otel.Handle()` with error conditions
  caused by out-of-range metrics input values, including NaN, Inf, and
  in some cases negative values. These will be handled no more than
  once per 30 seconds per condition.
  [#272](https://github.com/lightstep/otel-launcher-go/pull/272)

## [1.10.1](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.10.1) - 2022-08-29

### Changed

- Revert the default change of temporality to "cumulative" from #258.
  New users are recommended to configure
  `WithMetricExporterTemporalityPreference("stateless")` temporality
  preference if possible. [#261](https://github.com/lightstep/otel-launcher-go/pull/261)

## [1.10.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.10.0) - 2022-08-28

This version was retracted because of the change of temporality preference.

### 🧰 Bug fixes 🧰

- Partial error handling: avoid printing spurious messages when there is no error. [#247](https://github.com/lightstep/otel-launcher-go/pull/247)

### Changed

We have intentionally combined a number of potentially breaking
changes into this release,

- 🛑 [BREAKING] Histogram instruments now select the explicit-boundary histogram for
  the otel-go SDK, which is a breaking change for Lightstep users, but
  as the Lightstep SDK is using exopnential histograms this is the
  matching default which allows upgrade and downgrade between these
  SDKs. Users who encounter errors related to histogram instruments
  (e.g,. `process.runtime.go.gc.pause_ns`) please contact a Lightstep
  representative. [#249](https://github.com/lightstep/otel-launcher-go/pull/249)
- Exponential histogram changes boundary inclusivity from lower-inclusive to
  upper-inclusive. This is defined as a non-breaking change because implementations
  have not required exactness. With this change, exact powers-of-two are
  treated as exact special cases. [#254](https://github.com/lightstep/otel-launcher-go/pull/254)
- Lightstep metrics SDK is enabled by default. We are now confident in this
  SDK even as we have discovered new issues with the old one. [#258](https://github.com/lightstep/otel-launcher-go/pull/258)
- 🛑 [BREAKING] Stateless temporality preference is used by default; Counter and
  Histogram instruments will begin reporting with delta temporality. To avoid
  repeated breakage. This is especially important for exponential histogram
  data quality. [#258](https://github.com/lightstep/otel-launcher-go/pull/258)

## [1.9.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.9.0) - 2022-08-04

### 💡 Enhancements

- Update to OpenTelemetry 1.9.0

### 🧰 Bug fixes 🧰

- Exponential histogram correctly counts the zero-bucket when increment > 1 [#238](https://github.com/lightstep/otel-launcher-go/pull/238)

## [1.8.1](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.8.1) - 2022-07-28

### 💡 Enhancements

- Optimize the synchronous instrument fast path [#226](https://github.com/lightstep/otel-launcher-go/pull/226)
- Exponential histogram package is externally-usable [#222](https://github.com/lightstep/otel-launcher-go/pull/222)
- Add an example for stand-alone use of Lightstep metrics SDK [#219](https://github.com/lightstep/otel-launcher-go/pull/219)

## [1.8.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.8.0) - 2022-07-13

### 💡 Enhancements

- Update OpenTelemetry dependencies to OTel-Go version 1.8.0 [#216](https://github.com/lightstep/otel-launcher-go/pull/216)
- OTLP version 0.18.0: includes Min/Max support for exponential histogram [#215](https://github.com/lightstep/otel-launcher-go/pull/215)
- `sync.Map` replaced with `sync.RWLock` and plain map (saves 3 allocations) [#206](https://github.com/lightstep/otel-launcher-go/pull/206)

### 🧰 Bug fixes 🧰

- Race condition possible with multiple exporters fixed [#205](https://github.com/lightstep/otel-launcher-go/pull/205)

## [1.7.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.7.0) - 2022-06-30

### 💡 Enhancements

- Introduces the "Lightstep Metrics SDK" option, an alternative to the OpenTelemetry-Go
  community SDK. This SDK features experimental features listed below: [#196](https://github.com/lightstep/otel-launcher-go/pull/196)
  - Lightstep Metrics SDK: Exponential histogram on-by-default, no explicit-boundary Histogram
  - Lightstep Metrics SDK: Support for MinMaxSumCount aggregation (via a 0-bucket Histogram)
  - Lightstep Metrics SDK: Support for synchronous Gauge instrument (configured via API hints)
- New in both Metrics SDKs: Support for "stateless" temporality preference
- New in both Metrics SDKs: Recognize & print Lightstep's metric validation error details.

## 1.6.x

We are skipping version 1.6.x so that `otel-launcher-go` major and minor versions match the `opentelemetry-go` repository.

## [1.5.2](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.5.2) - 2022-05-12

### 🧰 Bug fixes 🧰

- `WithResourceAttributes()` fixed, had been non-functional [#165](https://github.com/lightstep/otel-launcher-go/pull/165)

## [1.5.1](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.5.1) - 2022-03-24

### 🧰 Bug fixes 🧰

- tag `github.com/lightstep/otel-launcher-go/pipelines` and update go.mod for launcher

## [1.5.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.5.0) - 2022-03-18

## 💡 Enhancements 💡

- use semconv package instead of collector package (#69)
- update OpenTelemetry SDK dependencies to 1.5.0
- update OpenTelemetry metrics to 0.27.0 (#114)
- Allow Developer Mode For Go Lightstep (#60)

## [1.0.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.0.0) - 2021-09-22

### Changed

- Update OpenTelemetry tracing dependencies to 1.0.0

## [1.0.0-RC3](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.0.0-RC3) - 2021-09-15

### Changed

- Update OpenTelemetry tracing dependencies to 1.0.0-RC3
- Update OpenTelemetry metrics dependencies to 0.23.0
- Remove support for deprecated `OTEL_RESOURCE_LABELS` environment variable, as
  it has been replaced by `OTEL_RESOURCE_ATTRIBUTES` upstream.

## [0.20.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v0.20.0) - 2021-05-10

### Changed

- Update OpenTelemetry dependencies to 0.20.0

## [0.18.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v0.18.0) - 2021-03-16

### Changed

- Update OpenTelemetry dependencies to 0.18.0

### Added

- Add support for `ottrace` propagator

## [0.16.1](https://github.com/lightstep/otel-launcher-go/releases/tag/v0.16.1) - 2021-02-01

### Changed

- Enable gzip compression by default
- Update OpenTelemetry dependencies to 0.16.0

## [0.15.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v0.15.0) - 2020-12-15

### Changed

- Update OpenTelemetry dependencies to 0.15.0
- Added context param for NewExporter as required by OpenTelemetry 0.15.0

## [0.14.1](https://github.com/lightstep/otel-launcher-go/releases/tag/v0.14.1) - 2020-11-24

### Changed

- Update default value for metrics exporter endpoint (#38)

## [0.14.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v0.14.0) - 2020-11-23

### Changed

- Update configuration for baggage propagator from `"cc"` to `"baggage"` (#30)

### Added

- Add support for `tracecontext` propagator (#32)
- Adding LS_METRICS_ENABLED env variable to control host/runtime metrics (#34)

## [0.13.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v0.13.0)

### Changed

- Update OpenTelemetry dependencies to 0.13.0 (#26)
- Use BatchSpanProcessor instead of SimpleSpanProcessor (#24)

### Added

- Add host.name tag based on os.Hostname (#19)

## [0.11.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v0.11.0)

### Changed

- Update OpenTelemetry dependencies to 0.11.0

### Added

- Add a metrics exporter

## [0.10.2](https://github.com/lightstep/otel-launcher-go/releases/tag/v0.10.2)

### Changed

- update envconfig package

## [0.10.0](https://github.com/lightstep/otel-launcher-go/releases/tag/v0.10.0)

### Changed

- Update OpenTelemetry dependencies to 0.10.0

## [0.0.1](https://github.com/lightstep/otel-launcher-go/releases/tag/v0.0.1)

### Added

- Initial release
