# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

🛑 [BREAKING] The set of builtin, automatic metrics instrumentation
libraries configured by the Launcher is new and improved in this
release  To configure the behavior in Launcher versions 1.10.x and
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
  are selected as stable builtin instrumentation.  To configure the
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
  in some cases negative values.  These will be handled no more than
  once per 30 seconds per condition.
  [#272](https://github.com/lightstep/otel-launcher-go/pull/272)

## [1.10.1](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.10.1) - 2022-08-29

### Changed

- Revert the default change of temporality to "cumulative" from #258.
  New users are recommended to configure
  `WithMetricExporterTemporalityPreference("stateless")` temporality
  preference if possible.  [#261](https://github.com/lightstep/otel-launcher-go/pull/261)

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
  SDKs.  Users who encounter errors related to histogram instruments
  (e.g,. `process.runtime.go.gc.pause_ns`) please contact a Lightstep
  representative. [#249](https://github.com/lightstep/otel-launcher-go/pull/249)
- Exponential histogram changes boundary inclusivity from lower-inclusive to
  upper-inclusive. This is defined as a non-breaking change because implementations
  have not required exactness.  With this change, exact powers-of-two are
  treated as exact special cases.  [#254](https://github.com/lightstep/otel-launcher-go/pull/254)
- Lightstep metrics SDK is enabled by default.  We are now confident in this
  SDK even as we have discovered new issues with the old one. [#258](https://github.com/lightstep/otel-launcher-go/pull/258)
- 🛑 [BREAKING] Stateless temporality preference is used by default; Counter and
  Histogram instruments will begin reporting with delta temporality.  To avoid
  repeated breakage.  This is especially important for exponential histogram
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
  community SDK.  This SDK features experimental features listed below: [#196](https://github.com/lightstep/otel-launcher-go/pull/196)
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
