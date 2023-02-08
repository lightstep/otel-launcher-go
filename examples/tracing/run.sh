#!/bin/sh

# jmacd-wonderland
#export LS_ACCESS_TOKEN=xKNu3iOPZzWgG+X6fBNzuuPKwpnkQYbGwnlzFHGPgwdsJgYtbyUGP0KbVJTOF5ryuIrVtF28ikzPHzhmxOCrD8RoAH47/9Hnzh1t6R7yds
#export LS_SERVICE_NAME=experiment-test

## staging dev-jmacd
export LS_SERVICE_NAME=arrow
export LS_ACCESS_TOKEN=42f288975e1b8048d49ff5214d0359b2

# local lst-dev
#export LS_ACCESS_TOKEN=DokIqMCWww9QWLxsN1i3lU7AHtK2UIkxWvTsWe8ZPAzOsvdhS6W2QTNot8W3fc0SB2s5an4zc5Adfwp2LyI=

# @@@ wat this is broken
#export OTEL_EXPORTER_OTLP_HEADERS=X-Scope-OrgID=jmacd

export OTEL_EXPORTER_OTLP_SPAN_ENDPOINT=127.0.0.1:4317
export OTEL_EXPORTER_OTLP_SPAN_INSECURE=true

#export OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=ingest.lightstep.com
#export OTEL_EXPORTER_OTLP_METRICS_INSECURE=false

#export LS_METRICS_ENABLED=true
export LS_METRICS_ENABLED=false

GO=go

${GO} version

${GO} run .
