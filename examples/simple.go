package main

import (
	"context"
	"fmt"

	"github.com/lightstep/opentelemetry-go/ls"
	"go.opentelemetry.io/otel/api/global"
)

func main() {
	ls.ConfigureOpentelemetry(
		ls.WithAccessToken("<access token>"),
		ls.WithServiceName("sdk-go"),
		ls.WithSatelliteURL("ingest.lightstep.com:443"),
		ls.WithDebug(true),
	)
	tracer := global.Tracer("ex.com/basic")

	tracer.WithSpan(context.Background(), "foo",
		func(ctx context.Context) error {
			tracer.WithSpan(ctx, "bar",
				func(ctx context.Context) error {
					tracer.WithSpan(ctx, "baz",
						func(ctx context.Context) error {
							return nil
						},
					)
					return nil
				},
			)
			return nil
		},
	)
	fmt.Println("OpenTelemetry example")
}
