// Copyright Lightstep Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"sync"

	metricService "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	traceService "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcMetadata "google.golang.org/grpc/metadata"
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
		lock sync.Mutex

		metricsRequests []*metricService.ExportMetricsServiceRequest
		metricsMDs      []grpcMetadata.MD
		traceRequests   []*traceService.ExportTraceServiceRequest
		traceMDs        []grpcMetadata.MD

		InsecureMetricsPort int
		SecureMetricsPort   int
		InsecureTracePort   int
		SecureTracePort     int
	}
)

var (
	// The certificates and keys used in these tests expire in 18
	// months from generation, because macOS doesn't like
	// long-duration self-signed certificates (see
	// https://myupbeat.wordpress.com/2022/09/09/self-signed-certificates-not-standards-compliant/).
	// These were generated using the github.com/square/certstrap
	// utility.
	//
	//   certstrap init --common-name TestCA
	//   certstrap request-cert --common-name TestServer --ip 127.0.0.1
	//   certstrap sign TestServer --CA TestCA

	// From ./out/TestCA.crt. Also copy this file to ./testdata/caroot.crt
	TestCARootCertificate = `-----BEGIN CERTIFICATE-----
MIIE4jCCAsqgAwIBAgIBATANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDEwZUZXN0
Q0EwHhcNMjQwMTA0MjEwMzM3WhcNMjUwNzA0MjExMzMzWjARMQ8wDQYDVQQDEwZU
ZXN0Q0EwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQC1wpcKdXn3d0rd
/2fM0obaPV9Dan08uqWh/DxUOQ6E5207vuUE16BQOQ7jhmRFBaLQYh36WlasBjbi
GZt7FPxleLZGCik8EkeQ97XMLMzUeCjI/bTCoxa6Te5bd61V8NcT7abzSteUfet0
vGTru/VskdXE11gb05sLCdL2/SbDBS6PExKVvqeugKaY9kI4WK+qFK/bO1HAvV9y
1yh9tUpoOxU99chP3sfLAuaMQ5mnusy9CtdhRkeaLkbPdCH8Q6A/aeY1oN+X4MOu
1ZmffCdCpGJKldiRq0+tLoCEAkUO843+ixjm0BOekgiFTX+xz5VIM0p+f6AAGZbi
XU7gF/TcksIRdJxqHZyajxFwrYVNQh+Quy79hQD2UEbW7+nQ/6l9mcqckeAX74kO
2Sa9C0ObSK4Eagr37LiTSvDrO9t8KMB9wIeQCYnG8yZtOEkr2DmtEMSm4cT57Kr3
QqWYXm+TsCTVHAQsKzqMc4udxI+UPjCCnm4qTRNrVtVfepsB2Vse4AuqNiFyLHzc
hRLC1solw7cQAHvqd4PfieyY6WP4GYEB1cMUkn3Zci/8JTeqUSvdI1Wwa5FYwCHg
iDeJ0q7QxwIQNwdI73l9oTXYOffJBwQpdBDY+Kif+iwZB43N1LMFZ11qkNzwV23R
hDb70bOH5U6i/ZMR0dtuoVEQAAqIjwIDAQABo0UwQzAOBgNVHQ8BAf8EBAMCAQYw
EgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQUsPUEbbXShAII4KrUlmSpcZpM
wGowDQYJKoZIhvcNAQELBQADggIBAGErXaFxWKAv6GjgFPM7mt+TGaea/WZMT4Hc
4sfRQhxvn41bEcNchNZaXwSC9kPUMl1il+s2CuIBHeWusObF6S+Ez5VBPn89F8BV
/YXNp2MJSTgZ+AunYOw8voP+d8/CkpFCQ+K72LAY/zsVNT7etpBBwuReLgiKqYnU
BuHTwVbPLYuSdJZh498bYuRf0h9cWQ+UhC3lKQQ6ucl179r5wsNLf8YNt6xICeXQ
LxEx9NVdLuQd342uYGdpRrSHswy5Nr5v9TQ6y65sZwyVn4H/JjJf6h41HzX2OrW7
x7hEcqnV2ZO+7VAEUqyBN0XmMQdjcaQaTrX+Pd57tLJheBQdL8Ze2hcqx7+6u1Os
+ufTYv1J6EZTGsIBWGxBArh3A/VvqpFWdywAYhnJ4yjDyXwZQCbMKAtgOBMOmXke
/8IZ/DBzMnMWI9XeyoT6PSBOKdKiDcbs+vxv1mmaK8M4rgIlRVfNfXd+kXLG+HVH
hPydKOyxAHWNjMfo/FUrlAr2GX7zkacjBxTRro9PCNWDVjApxpWCmyicWkZbJdOJ
th55vgMIA4xgwwgWP4WhDsvIymyMTSc6wSOYvErN59UCvyep2ehjLts3T0PHFxes
V9TLPF9CESK0loKQ2kWIcJPiYIwD3iaPA+822SAGpRpGAXFnHHzMrCaTeucxIX0S
yHREhbzT
-----END CERTIFICATE-----`

	// From ./out/TestServer.crt.  Also copy this file to ./testdata/testserver.crt.
	TestServerPublicCertificate = `-----BEGIN CERTIFICATE-----
MIIENTCCAh2gAwIBAgIRAN9ApcH3IJHTLQoQsmOYF50wDQYJKoZIhvcNAQELBQAw
ETEPMA0GA1UEAxMGVGVzdENBMB4XDTI0MDEwNDIxMDM0OFoXDTI1MDcwNDIxMTMz
M1owFTETMBEGA1UEAxMKVGVzdFNlcnZlcjCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAPAhQ0gfE06x0InoLKlk4o6xtWk21a79ljJyhkSAoKoZxni8drW3
iVVuHTM67enQ1jH2btkOTOkgNr11ZQ3jSzCfluZaTnGDPBPauLpi5HQtpID6aTb7
QxcEMPhJ3/09OSb9Tj29b9xA4ydJ3ZazYPJ2BX2A6DErDFnNIhykmUAsAbdMNgDj
jLHZkdZJjBUawk2ks1cGU5RoKwBzilSfmdb5bcrriMYc36IwCDF54Sia1mf4Q7RO
eDShKbUzNvaq0+DRBJLeQtgv4cls49ScixgNDTaxzHViLUCc9fL9jRT2shRWVUGl
eCeGHv8nlLp28tbV0wT4kZkZgNm0z8y6vz8CAwEAAaOBgzCBgDAOBgNVHQ8BAf8E
BAMCA7gwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMB0GA1UdDgQWBBS6
6BcWwzODodb1lSo2m2zhR3Ac7zAfBgNVHSMEGDAWgBSw9QRttdKEAgjgqtSWZKlx
mkzAajAPBgNVHREECDAGhwR/AAABMA0GCSqGSIb3DQEBCwUAA4ICAQByNW91LeIT
Qmcw8gUn5nE8M1KuACqwjj7hAN+C8UWNUZCdSdC0eky0twkQ9tb6gEC46ZfLH3AN
QYXbIyDaTRT4rwLNgyFjnXdeLkRZMfTEYxZueAyGKhdCeZoTDfsLoh/8ErSidAli
NTgg0RbJB+J8IMkO9SS8R1fSHBUOQoqn6jEGW0J+/FUuG5bP+PDcHWO/fMV/KzbJ
84c778bsd1cxdXT9zc5rMFc/0SDvgoD/80T5baIDc0s/7VEWZAv0N3aJW653jwE+
QHec2eXlU8Sfq9FjpxIx6G8ssY29ppJLCWqaXZXzE1Bl+rVabWam3PYhzppcoTFw
Mm0ba+1rTk/bVUW87wEpWMphkQRDv0nfOy3TlFdF0/3riELadLlkADly2qS8Heok
ng+FfxAJyOKiKNh65iszjfi7Ap2VxnwJDmA8K4QjcHmUrjYjYb2s0xDafrpeK6vs
lfqQl+LBGgt3DE6TUUJNeyElASYjVCT6GhtQ00W9E6Dv1ElQy9P6lx/ppr3COSYE
dCmUten2Pyj29KC76N7/HUDQ4ChwZOatjC9cT5Y4xpVL0yQgLvahLuSBozL9DAeN
q4Q9OQKV1DlT9rSVsPdICdLq/d+sjVvYKwSifKeddhwb/NJKkmeNAmc7ICSpjK7R
WXWUELD+mIE4KekmcP0NDVVDYAKniqIprg==
-----END CERTIFICATE-----
`

	// From ./out/TestServer.key. Also copy this file to ./testdata/testserver.key.
	TestServerPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA8CFDSB8TTrHQiegsqWTijrG1aTbVrv2WMnKGRICgqhnGeLx2
tbeJVW4dMzrt6dDWMfZu2Q5M6SA2vXVlDeNLMJ+W5lpOcYM8E9q4umLkdC2kgPpp
NvtDFwQw+Enf/T05Jv1OPb1v3EDjJ0ndlrNg8nYFfYDoMSsMWc0iHKSZQCwBt0w2
AOOMsdmR1kmMFRrCTaSzVwZTlGgrAHOKVJ+Z1vltyuuIxhzfojAIMXnhKJrWZ/hD
tE54NKEptTM29qrT4NEEkt5C2C/hyWzj1JyLGA0NNrHMdWItQJz18v2NFPayFFZV
QaV4J4Ye/yeUunby1tXTBPiRmRmA2bTPzLq/PwIDAQABAoIBAQDkJvaCQ+RYVOJK
5Wnp2IzZ/0baHNuSVCas79taoswEUlEczhQMO8IkhWcBEfCSw3WAKyDO4qN4rL7V
7ACD3X3HSRpa61q0x3gBdUMm9GcTa3ptgX8OWlU4PSc6ARbsyYrP3MTGLINnxc8N
uUTstqpaNICq6huy+6/Ucu8CP/HL4Szmxwih99zIkhQEqAedKWgX/82DRO6ulH5d
GrYOxUDwandbkssevKE344K4/GlhJhAgN6PVFyYlMFpsVHqjDzzKkYXzcUtk8DaY
2+G2kLEuWiZSXxxVFEKY3HVRXXUN8hffkd+yWdsMFipg56lEq+Al9Xwk3sLSwV8Y
F8ZUki5hAoGBAPYZlVOxtDX1VNNNbURh5/RlmatraQIxk/CdIjsgHa1s/niVCRBF
vltmZzn7LLGt6dw4di9CubCmHdD75SfAynFaEQTOos19YhTC2g2idw4XdLIikDKK
OeppaKNUk+rRXXcdiLyhtDjef07HOvmKLp6+7F1IPmo5k8/yr7Pj598jAoGBAPnK
MtD43ON0cdBItUmrIt0pYTJebHwL52YsrslWG4mwAXiU4PrF1YeNpRN+OCBHEpb/
Az0xBalUTcUtFSGEBAFk/si8ySdYv/j2YafLWuLqi4rVORtGRt3dC9evQeSGCj/o
Pffbau9OFjIen5XLEzUnGbesEqVV/++VvJv7Z481AoGBAOhLy8U9dvJ7yX7OlfY3
SEBL6tqAv5T/gTpcyCPxM7IwsJ7XZr/CZWVW6tcy/MQWeimR7hS8MhTJKFnMe0ij
1TNbpbbY6Zl34a3hIvw9v41AnLlMoLnj+bkHmGqbeifrSgMWkKwlIr2PX7HXoxZK
1aioZOnEOI4CHUDrPehaltLrAoGATHEsi/cc4h7Ilc0qbZkJ2lTHgfqTiIK8FfCm
rMbFNqW+TYCCOTxB1HHsisKduoMFlWAFRbyy1tcN1cGuLcuQzjxyHExp4riuRypf
SFJbRgYxHhOSnl4rYco7zY28xIqgqF4SWL+1QfbLpBrrC5RSFHoazLLEIgTnhhJ0
3edaEeECgYEA8+3ootFn6Hn/h8rjMW7Znwge2sWYHgN4bFv5EZJqlXzCw2HS4xz6
suIHe/ttW0xJoIt8HmmTXmqUdF3iBvWlTbJjxMSaokpmHIMF3iOZexUy7sFaZCFN
0gX7Y2eWK6gxZ4kfoFoi0X5FCk3mn0+wcRmKO8xfpmmoLq5JwbCB1r4=
-----END RSA PRIVATE KEY-----`

	// ServerName is encoded in the above certificates.
	ServerName = "127.0.0.1"

	ErrUnsupported = fmt.Errorf("unsupported method")
)

func NewServer() *Server {
	certificate, err := tls.X509KeyPair([]byte(TestServerPublicCertificate), []byte(TestServerPrivateKey))
	if err != nil {
		log.Fatalf("test certificates: %v", err)
	}

	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM([]byte(TestCARootCertificate))
	if !ok {
		log.Fatalf("failed to append client certs")
	}

	tlsConfig := &tls.Config{
		ClientAuth:   tls.NoClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	}

	newListener := func() (net.Listener, int) {
		listener, err := net.Listen("tcp", fmt.Sprint(ServerName, ":0"))
		if err != nil {
			log.Fatal(err)
		}
		port := listener.Addr().(*net.TCPAddr).Port
		return listener, port
	}

	stop := make(chan struct{})
	server := &Server{
		stop: stop,
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

		go func() {
			_ = grpcServer.Serve(listener)
		}()

		defer grpcServer.Stop()
		<-stop
	}(insecureMetrics)

	go func(listener net.Listener) {
		grpcServer := grpc.NewServer()
		traceService.RegisterTraceServiceServer(grpcServer, &traceServer{Server: server})

		go func() {
			_ = grpcServer.Serve(listener)
		}()

		defer grpcServer.Stop()
		<-stop
	}(insecureTrace)

	go func(listener net.Listener) {
		serverOption := grpc.Creds(credentials.NewTLS(tlsConfig))
		grpcServer := grpc.NewServer(serverOption)
		metricService.RegisterMetricsServiceServer(grpcServer, &metricsServer{Server: server})

		go func() {
			_ = grpcServer.Serve(listener)
		}()

		defer grpcServer.Stop()
		<-stop
	}(secureMetrics)

	go func(listener net.Listener) {
		serverOption := grpc.Creds(credentials.NewTLS(tlsConfig))
		grpcServer := grpc.NewServer(serverOption)
		traceService.RegisterTraceServiceServer(grpcServer, &traceServer{Server: server})

		go func() {
			_ = grpcServer.Serve(listener)
		}()

		defer grpcServer.Stop()
		<-stop
	}(secureTrace)

	return server
}

func (s *Server) TraceRequests() []*traceService.ExportTraceServiceRequest {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.traceRequests
}

func (s *Server) TraceMDs() []grpcMetadata.MD {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.traceMDs
}

func (s *Server) MetricsRequests() []*metricService.ExportMetricsServiceRequest {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.metricsRequests
}

func (s *Server) MetricsMDs() []grpcMetadata.MD {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.metricsMDs
}

func (s *Server) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()
	close(s.stop)
	s.stop = nil
}

func (s *metricsServer) Export(ctx context.Context, req *metricService.ExportMetricsServiceRequest) (*metricService.ExportMetricsServiceResponse, error) {
	var emptyValue = metricService.ExportMetricsServiceResponse{}

	md, _ := grpcMetadata.FromIncomingContext(ctx)
	s.lock.Lock()
	defer s.lock.Unlock()
	s.metricsRequests = append(s.metricsRequests, req)
	s.metricsMDs = append(s.metricsMDs, md)

	return &emptyValue, nil
}

func (s *traceServer) Export(ctx context.Context, req *traceService.ExportTraceServiceRequest) (*traceService.ExportTraceServiceResponse, error) {
	var emptyValue = traceService.ExportTraceServiceResponse{}
	md, _ := grpcMetadata.FromIncomingContext(ctx)

	s.lock.Lock()
	defer s.lock.Unlock()
	s.traceRequests = append(s.traceRequests, req)
	s.traceMDs = append(s.traceMDs, md)

	return &emptyValue, nil
}
