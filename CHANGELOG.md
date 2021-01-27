## Unreleased

## 0.16.0
- Update OpenTelemetry dependencies to 0.16.0

## 0.15.0
- Update OpenTelemetry dependencies to 0.15.0
- Added context param for NewExporter as required by OpenTelemetry 0.15.0

## 0.14.1

- Update default value for metrics exporter endpoint (#38)

## 0.14.0

- Add support for `tracecontext` propagator (#32)
- Update configuration for baggage propagator from `"cc"` to `"baggage"` (#30)
- Adding LS_METRICS_ENABLED env variable to control host/runtime metrics (#34)

## 0.13.0

- Update OpenTelemetry dependencies to 0.13.0 (#26)
- Use BatchSpanProcessor instead of SimpleSpanProcessor (#24)
- Add host.name tag based on os.Hostname (#19)

## 0.11.0

- Update OpenTelemetry dependencies to 0.11.0
- Add a metrics exporter

## 0.10.2

- update envconfig package

## 0.10.0

- Update OpenTelemetry dependencies to 0.10.0

## 0.0.1

- Initial release
