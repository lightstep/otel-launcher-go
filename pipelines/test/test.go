package test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"testing"

	metricService "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	traceService "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
)

type (
	metricsServer struct {
		*Server
		metricService.UnimplementedMetricsServiceServer
	}

	traceServer struct {
		*Server
		traceService.UnimplementedTraceServiceServer
	}

	Server struct {
		stop chan struct{}

		InsecureMetricsPort int
		SecureMetricsPort   int
		InsecureTracePort   int
		SecureTracePort     int
	}
)

var (
	TestCARootCertificate = `-----BEGIN CERTIFICATE-----
MIIE+jCCAuKgAwIBAgIBATANBgkqhkiG9w0BAQsFADAdMRswGQYDVQQDExJUZXN0
T1RlbExhdW5jaGVyR28wHhcNMjIwNjIxMTgwNjM0WhcNMjMxMjIxMTgxNjMyWjAd
MRswGQYDVQQDExJUZXN0T1RlbExhdW5jaGVyR28wggIiMA0GCSqGSIb3DQEBAQUA
A4ICDwAwggIKAoICAQC73wwMPTeOI4fjJaIkHa4Lveupj+RX09dc2ZVDiCh9JH7B
w7H+TaUTaS+fUCg3qmeA7Nc0VdC3wOwNyRxiCPbg6a/TOzx3WuEHvN01Z+ugLr63
UKbrMKFv6TaXrcb+JbpijsgCxSNmM+kV5Olha6xKYVB1h/7LCnz2HCGrBLPRtxWO
4XRW2MMqjZgipQq2QSk1MuNv24VLiUZr5JoSoN7X792fgIEEHHLzp5BOZqsAIT8a
b8jjoer58Me/0mVqNxOl25Uefu9lh/cl1XDiwY5+zp+613ZIOweSekpHw5vOzQ32
dCvEYXt0SwgCswjnJH+f70YuclCZrCVhq62PO//cXFj+wY/wuMIq7AUDJskChaKU
bBXwB1rjtWhcByAm79bWYVTuNv7RrbrQWIUDgTfj5ce8UnrKcrDkMMT54g3HJ/HX
mPPhl/R5POY/6iy+ypw6TTNr/EgXTVBK7ewAhMKSViojERLcwW/uTeCo9vk5semv
jSvPR46QL89qVBPGdaLSTp+rwihczmMjEyaUJgYa28K8gQi6rdETKbOmM7hlTBvm
9pDQXpED7HFjnZ5WiwZJR6oz1678I6rtqTe4G7Ar3kM3ytDcFXcDHFRafwxS4E7W
CiCD5dQ6S0UNpLwSSMJp7EDoSBCKwYIkQbU2/bSttQ0nUhi5NrPGe7NmTKrcpQID
AQABo0UwQzAOBgNVHQ8BAf8EBAMCAQYwEgYDVR0TAQH/BAgwBgEB/wIBADAdBgNV
HQ4EFgQUxgbXWOdaCbEVUC+9GbmLLbLwD0kwDQYJKoZIhvcNAQELBQADggIBALlj
7fGNCmi1qz0HaY/1Z1j5mbKr8miiG0GWnPhHeMbegUOd7IayHOPmXjyUxXk5P9X3
PswUyEDeFbdRP581vED87nez0+ZKYdms11R9941+dJWGV6LfSvxU43y0bCl/BFoE
Rn3vnH2vt3HqhHLI/jgRwp16u7jLgOUijcYBPIJakoQQKHyhFx8ePb6tQugpkg9w
wbgTYp5gGBk5rTQVt9fzZN0dJ0CQMJjknyfTPOaasXJAlso2lwbpiwqOGbIAX/2i
N2t+JE3tReTkHMnaAOi7f23i6ai/AZAMMHBaZmx5scWJUZxdxG/aio6xGvHMYxUV
AWlUSJsEhStQ7HqI88RdbHinuilTqnXzYXlmNTfdd/i0owkdCo7/7LpHsOkXSBOR
E5HnA/Wb3uciB0M665ShGETU9ZyVoSw0dCDEeErBFFTLf+1hA2IKah/L+bpsh3j8
R3ej00fOqHFOI2os7TKmziz6Lx/8+kHrESxMQFmvPT0FwMq2aat4FGMxENTPmvRi
Sg7QpMhEmvdU08hTXyUBC6hiSau4Ka5hxng/lvhvyfClGsCGWMYip1yv4C3m/kGx
xGFrSfPf6GimKxvfJ06aJCP8N3MAqAHj278Wsi9gN7wIG+h5g35J4MmgzMOMLWlP
NkjHQ0jCBNZ2wekgribyNtFwVfECyJ5XQAH+SutY
-----END CERTIFICATE-----
`

	TestServerPublicCertificate = `-----BEGIN CERTIFICATE-----
MIIELDCCAhSgAwIBAgIRAKQ7lQ5Fb5wgWXa2C+SpdokwDQYJKoZIhvcNAQELBQAw
HTEbMBkGA1UEAxMSVGVzdE9UZWxMYXVuY2hlckdvMB4XDTIyMDYyMTE4MDc0MloX
DTIzMTIyMTE4MTYzMVowEzERMA8GA1UEAxMIVGVzdENlcnQwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQDFD3+pMdHiZ2UsHyepKSJKTdlvX2jtzs6Y1u24
dhWiVFNMajZQHcuYoETk5QWktb/BQ6KL9i6C5RgpCWFlBhncLCSzR04gKqMRCg+b
rdhxlxT79ycyH9jY2Zvnmmht36S86XC+bhUAvvST6R88GDHQi8KUN4TGxuxiz8/v
0yqYEthmK79Ke77DKOWoHuu5OK9lH44XJMtcRDq0FLkJMFl6sWY9gIOITNF/KU6L
Dp8T6C1hv8qBh9D2gIXX5gBmLc6t9I7x/XP4ilVM1oYznzuQS1L8P9XpWFktDwkS
x7O2VqhzPFKBX3liHylC6Q4DHVMnqNMgktHSWoKuxhYLiUCvAgMBAAGjcTBvMA4G
A1UdDwEB/wQEAwIDuDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwHQYD
VR0OBBYEFI2V2h0pHlDofX5vz4s/7xj2fbvbMB8GA1UdIwQYMBaAFMYG11jnWgmx
FVAvvRm5iy2y8A9JMA0GCSqGSIb3DQEBCwUAA4ICAQCB0XVVLnCSi8VAKiojX5pj
LDQME7o1xSkXUGcOzBvDU0XGe28N7zh0roS98ml2TJqmNqsBA9LpktEK8rkWlCyb
iIM6Sg4GdPgmC5vXURdDpFEtujzTVx5J5M1efMCb8jBqNM/nlj/lWshTVi4OvZVb
KEfNfVeV+isIkZx8LNoohOlqM3s/K6FuSU3qZa9JXYsdsjJn97xg3fukx7/SDcML
JW/0qS+VtbMZUpmN0BWo47u2JGcAS3gHSTPYfedbliQjWzGNJNcmWVskIZJn6uti
EK6KqMnAE86xFObZlr2IEyuh7IVHN3UlGqI+wj+XvlV2qT2oX7Z66El5IEcDVS5s
0VFhXk4n5wD8Y51wXUWpkrkJRegitQBKol51zfSU6c8CHIQ7ZSK3VEj5w/p6R6tl
sr7guDE5a4tEx1Sj8TRYOdlayY5wPyYdzYSpbKe7TbXzffxwOG36u6xup6YPVYtI
+yBP0l+Hpx4+t6hqZS0Saba1+VV55hhdcJXGdsv9IqK7VcVkrWeDPE6PdA4OjOxM
mtD1RGKbLrZ/2RWC/EpqWRpYhO5zO3JgUwaw7UV4trwt81XU4ROgNJ3/qkmth10v
JQp2vO1plfX+LbXSRnJsKk2FAFSZFzAYqy7gneQdm6DCn3dvPPCm7V8adX3O/ZoU
eLEBc7vRK92zDssz9H1tog==
-----END CERTIFICATE-----`

	TestServerPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAxQ9/qTHR4mdlLB8nqSkiSk3Zb19o7c7OmNbtuHYVolRTTGo2
UB3LmKBE5OUFpLW/wUOii/YuguUYKQlhZQYZ3Cwks0dOICqjEQoPm63YcZcU+/cn
Mh/Y2Nmb55pobd+kvOlwvm4VAL70k+kfPBgx0IvClDeExsbsYs/P79MqmBLYZiu/
Snu+wyjlqB7ruTivZR+OFyTLXEQ6tBS5CTBZerFmPYCDiEzRfylOiw6fE+gtYb/K
gYfQ9oCF1+YAZi3OrfSO8f1z+IpVTNaGM587kEtS/D/V6VhZLQ8JEseztlaoczxS
gV95Yh8pQukOAx1TJ6jTIJLR0lqCrsYWC4lArwIDAQABAoIBACqM0Szwc/hmCQOA
6qhtGFlg++0/dcG7oQKBji0BWmSFvsLGQFoGRPr8yEOAbDqHgBM0DnoYOyzKWPAr
dVtB+P2AjqAjamwpqLI6MOqVnCHS1JYfZNg+5izUuARHY/stij28TjrgPCrAEMGL
WdI5CzCTrP0iC8p8E3i2lJidSRoyvfGk/xKzTSZ4/CpdIZzgvtSgcAjfkMvB2zZG
pDC0TP+l5cUhUgxjBsYIUoYl3sFYPxpC63nWVV/q3nBm3q6WnTO8FLWyacObVGOt
UzwHbzU887ZPgQozBuO2a1QSm5JPmCKjJwG4QMPWiWh+b/y2acoCRSbXRx9jaFK7
oc2s/vECgYEA+orDsXIRJXmu3XijcLUFqaM/XTFE80YZEA0Qs/52l+ovHOPA32At
9z4P9HI+9bI22KnEWrkWZN3bxs6XOe8Xo6qt4JLY4PwaRD9WKe7hOnO8JFYE27Op
8Mc1RddVwGKte2GYaYzHDGWl5m/OfJrfrrUkg1H+SWON5wU6G/Mrp5kCgYEAyVp5
0vvr5RvZiezG5pCYC68u7WoWco5sT7e3Sf86rNDEwwumXY3u2vVxQx0E522vzaCE
rTkYvQLNxO3SkE/OepGk2ky00fooGjA48LtLvWeGGgV0kk18v10yn3VMzIF/2VwK
q8Z+8gXW0jXKb6EcQZMY0IT/fDWo2SGv7mvSN4cCgYEA+es99l3QmM9fDXFvp9gL
RAKiDHY/T2TXT1mZFdN5vWRPhsPx+2DXuU/hXngwMaqKZ2pBgjYrDob42sHtvE6y
CAMT23bgfN093mJHsyCk70fPn3dm9TmtBY/Rpk99LKHCZ9ccz/0r+UPUT5+sHEPp
aT8sowpBXDfAr3hZVNQm8dECgYBM2ur7HEtbJPkwyx7UbMaMVy6rUj4FNdWjy/T7
Gp+TzQ/9ftnehclw7BRyUIZJq7VZ4HYkBFIr+wD9tOUVTlD6udLZvEOcjkZ2UIe7
Y1Iylmw6THDFUyxVgsZK1SQePyPEnHw6OsbDrHTlwcBmQXGemf3zwYAfMgAj+NbF
Q4R2ywKBgQDT9H9AAjfBeDc7vvniXuKBq+70IV1aSAZrtGEy1UHTrZPl/CHSaMR+
tx1P0kRuRjgvVuQPejEpWHOJkQ1hzhCVo4fa2wl47Ot0qIroAHeMisxeTuUCiwgF
TKUq/P56Wk3pGIcwlNl5Tu+NE6WLQt16pU/yiwtfYr83RLGB6sR03w==
-----END RSA PRIVATE KEY-----
`

	ErrUnsupported = fmt.Errorf("unsupported method")
)

func NewServer(t *testing.T) *Server {
	certificate, err := tls.X509KeyPair([]byte(TestServerPublicCertificate), []byte(TestServerPrivateKey))
	if err != nil {
		t.Fatalf("test certificates: %v", err)
	}

	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM([]byte(TestCARootCertificate))
	if !ok {
		t.Fatalf("failed to append client certs")
	}

	tlsConfig := &tls.Config{
		ClientAuth:   tls.NoClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	}

	newListener := func() (net.Listener, int) {
		listener, err := net.Listen("tcp", "0.0.0.0:0")
		if err != nil {
			log.Fatal(err)
		}
		port := listener.Addr().(*net.TCPAddr).Port
		return listener, port
	}

	server := &Server{
		stop: make(chan struct{}),
	}
	var insecureMetrics, insecureTrace net.Listener
	var secureMetrics, secureTrace net.Listener

	insecureMetrics, server.InsecureMetricsPort = newListener()
	secureMetrics, server.SecureMetricsPort = newListener()
	insecureTrace, server.InsecureTracePort = newListener()
	secureTrace, server.SecureTracePort = newListener()

	go func(listener net.Listener) {
		grpcServer := grpc.NewServer()
		metricService.RegisterMetricsServiceServer(grpcServer, &metricsServer{Server: server})

		go grpcServer.Serve(listener)

		defer grpcServer.Stop()
		<-server.stop
	}(insecureMetrics)

	go func(listener net.Listener) {
		grpcServer := grpc.NewServer()
		traceService.RegisterTraceServiceServer(grpcServer, &traceServer{Server: server})

		go grpcServer.Serve(listener)

		defer grpcServer.Stop()
		<-server.stop
	}(insecureTrace)

	go func(listener net.Listener) {
		serverOption := grpc.Creds(credentials.NewTLS(tlsConfig))
		grpcServer := grpc.NewServer(serverOption)
		metricService.RegisterMetricsServiceServer(grpcServer, &metricsServer{Server: server})

		go grpcServer.Serve(listener)

		defer grpcServer.Stop()
		<-server.stop
	}(secureMetrics)

	go func(listener net.Listener) {
		serverOption := grpc.Creds(credentials.NewTLS(tlsConfig))
		grpcServer := grpc.NewServer(serverOption)
		traceService.RegisterTraceServiceServer(grpcServer, &traceServer{Server: server})

		go grpcServer.Serve(listener)

		defer grpcServer.Stop()
		<-server.stop
	}(secureTrace)

	return server
}

func (s *Server) Stop() {
	close(s.stop)
	s.stop = nil
}

func (s *metricsServer) Export(ctx context.Context, req *metricService.ExportMetricsServiceRequest) (*metricService.ExportMetricsServiceResponse, error) {
	var emptyValue = metricService.ExportMetricsServiceResponse{}
	return &emptyValue, nil
}

func (s *traceServer) Export(ctx context.Context, req *traceService.ExportTraceServiceRequest) (*traceService.ExportTraceServiceResponse, error) {
	var emptyValue = traceService.ExportTraceServiceResponse{}
	return &emptyValue, nil
}
