package ls

import (
	"log"
	"os"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

var (
	lsServiceName     = os.Getenv("LS_SERVICE_NAME")
	lsSserviceVersion = os.Getenv("LS_SERVICE_VERSION")
	lsAccessToken     = os.Getenv("LS_ACCESS_TOKEN")
	lsSatelliteURL    = os.Getenv("LS_SATELLITE_URL")
	lsInsecure        = os.Getenv("LS_INSECURE")
	Version           = "0.0.1" // overridden at build time
)

type Option func(*config)

func WithAccessToken(accessToken string) Option {
	return func(c *config) {
		c.accessToken = accessToken
	}
}

func WithSatelliteURL(url string) Option {
	return func(c *config) {
		c.satelliteURL = url
	}
}

func WithServiceName(name string) Option {
	return func(c *config) {
		c.serviceName = name
	}
}

func WithDebug(debug bool) Option {
	return func(c *config) {
		c.debug = debug
	}
}

type config struct {
	serviceName    string
	serviceVersion string
	satelliteURL   string
	metricsURL     string
	accessToken    string
	debug          bool
	insecure       bool
}

func newConfig(opts ...Option) config {
	var c config
	var defaultOpts []Option

	for _, opt := range append(defaultOpts, opts...) {
		opt(&c)
	}

	return c
}

func ConfigureOpentelemetry(opts ...Option) {
	c := newConfig(opts...)
	if c.debug {
		log.Println("debug logging enabled")
		log.Println("configuration")
		log.Println("-------------")
		log.Printf("%v", c) // TODO: pretty print
	}
	if len(c.serviceName) == 0 {
		log.Fatalf("invalid configuration: service name missing")
	}

	headers := map[string]string{
		"lightstep-access-token": c.accessToken,
	}

	secureOption := otlp.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if c.insecure {
		secureOption = otlp.WithInsecure()
	}
	exporter, err := otlp.NewExporter(
		secureOption,
		otlp.WithAddress(c.satelliteURL),
		otlp.WithHeaders(headers),
	)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}
	defer func() {
		err := exporter.Stop()
		if err != nil {
			log.Fatalf("failed to stop exporter: %v", err)
		}
	}()
	resources := resource.New(
		kv.String("service.name", c.serviceName),
		kv.String("service.version", c.serviceVersion),
		kv.String("library.language", "go"),
		kv.String("library.version", "1.2.3"),
	)
	tp, err := trace.NewProvider(
		trace.WithConfig(trace.Config{DefaultSampler: trace.AlwaysSample()}),
		trace.WithSyncer(exporter),
		trace.WithResource(resources),
	)
	if err != nil {
		log.Fatal(err)
	}

	global.SetTraceProvider(tp)
}
