package ls

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

// TODO: might be better to use something like envconfig here
var (
	defaultSatelliteURL = "ingest.lightstep.com:443"
	lsServiceName       = os.Getenv("LS_SERVICE_NAME")
	lsServiceVersion    = os.Getenv("LS_SERVICE_VERSION")
	lsAccessToken       = os.Getenv("LS_ACCESS_TOKEN")
	lsSatelliteURL      = os.Getenv("LS_SATELLITE_URL")
	lsMetricsURL        = os.Getenv("LS_METRICS_URL")
	lsDebug, _          = strconv.ParseBool(os.Getenv("LS_DBEUG"))
	lsInsecure, _       = strconv.ParseBool(os.Getenv("LS_INSECURE"))
	Version             = "0.0.1" // overridden at build time
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

type Logger interface {
	Fatalf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}

func WithLogger(logger Logger) Option {
	return func(c *config) {
		c.logger = logger
	}
}

type DefaultLogger struct {
}

func (l *DefaultLogger) Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

func (l *DefaultLogger) Debugf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

type config struct {
	serviceName    string
	serviceVersion string
	satelliteURL   string
	metricsURL     string
	accessToken    string
	debug          bool
	insecure       bool
	logger         Logger
}

func validateConfiguration(c config) error {
	if len(c.serviceName) == 0 {
		return errors.New("invalid configuration: service name missing")
	}

	if len(c.accessToken) == 0 && c.satelliteURL == defaultSatelliteURL {
		return fmt.Errorf("invalid configuration: access token missing, must be set when reporting to %s", c.satelliteURL)
	}
	return nil
}

func newConfig(opts ...Option) config {
	satelliteURL := defaultSatelliteURL
	if len(lsSatelliteURL) > 0 {
		satelliteURL = lsSatelliteURL
	}
	c := config{
		serviceName:    lsServiceName,
		serviceVersion: lsServiceVersion,
		satelliteURL:   satelliteURL,
		metricsURL:     lsMetricsURL,
		accessToken:    lsAccessToken,
		debug:          lsDebug,
		insecure:       lsInsecure,
		logger:         &DefaultLogger{},
	}
	var defaultOpts []Option

	for _, opt := range append(defaultOpts, opts...) {
		opt(&c)
	}

	err := validateConfiguration(c)
	if err != nil {
		c.logger.Fatalf(err.Error())
	}

	return c
}

type LightstepOpentelemetry struct {
	spanExporter *otlp.Exporter
}

func ConfigureOpentelemetry(opts ...Option) LightstepOpentelemetry {
	c := newConfig(opts...)
	if c.debug {
		c.logger.Debugf("debug logging enabled")
		c.logger.Debugf("configuration")
		c.logger.Debugf("-------------")
		c.logger.Debugf("%v", c) // TODO: pretty print
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
	resources := resource.New(
		kv.String("service.name", c.serviceName),
		kv.String("service.version", c.serviceVersion),
		kv.String("library.language", "go"),
		kv.String("library.version", Version),
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
	return LightstepOpentelemetry{
		spanExporter: exporter,
	}
}

func (ls *LightstepOpentelemetry) Shutdown() {
	err := ls.spanExporter.Stop()
	if err != nil {
		log.Fatalf("failed to stop exporter: %v", err)
	}
}

// TODO:
// need to implement closing the exporter somewhere

// defer func() {
// 	log.Printf("running defered code")

// }()
