#!/bin/sh

# jmacd-wonderland
export LS_ACCESS_TOKEN=xKNu3iOPZzWgG+X6fBNzuuPKwpnkQYbGwnlzFHGPgwdsJgYtbyUGP0KbVJTOF5ryuIrVtF28ikzPHzhmxOCrD8RoAH47/9Hnzh1t6R7y
export LS_SERVICE_NAME=experiment-test

# @@@ wat this is broken
#export OTEL_EXPORTER_OTLP_HEADERS=X-Scope-OrgID=jmacd

export OTEL_EXPORTER_OTLP_SPAN_ENDPOINT=127.0.0.1:4318
export OTEL_EXPORTER_OTLP_SPAN_INSECURE=true

export OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=ingest.lightstep.com
export OTEL_EXPORTER_OTLP_METRICS_INSECURE=false

export LS_METRICS_ENABLED=true

GO=gotip

${GO} version

${GO} run .
