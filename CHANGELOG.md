# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

## [1.0.0-RC1](https://github.com/lightstep/otel-launcher-go/releases/tag/v1.0.0-RC1) - 2021-06-21

### Changed
- Update OpenTelemetry tracing dependencies to 1.0.0-RC1
- Update OpenTelemetry metrics dependencies to 0.21.0
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
