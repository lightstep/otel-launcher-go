package main

import (
	"context"
	"fmt"

	"github.com/lightstep/opentelemetry-go/ls"
	"go.opentelemetry.io/otel/api/global"
)

func main() {
	lsOtel := ls.ConfigureOpentelemetry(ls.WithServiceName("my-service"))
	defer lsOtel.Shutdown()
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
