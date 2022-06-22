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
	"testing"

	metricService "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	traceService "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
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
MIIEPzCCAiegAwIBAgIRAKuNLsqy7M0h4nD5bnmS6sowDQYJKoZIhvcNAQELBQAw
HTEbMBkGA1UEAxMSVGVzdE9UZWxMYXVuY2hlckdvMB4XDTIyMDYyMTIzNTEzOFoX
DTIzMTIyMTE4MTYzMVowEzERMA8GA1UEAxMIVGVzdENlcnQwggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQCg7SFl18ctsqE9Gl9Hms/o482OzW1a4W1aGAaP
COrCsbASKZCQOY+IvFO0sgMcYB8lirZLdx0BGk+9fM/uHKzqVqLvv8K9cKdSfDJI
IPtcWvcOwq8+ChjqufBUVVDRnlBLtIg6MOiiHD2PvAjjz1MMcqBzA61j618eJ1cu
1UTU2tPSCa7Lbmj8oAW/RMzSAbmid+6jSDHbQcftp6Z4/x9GjMCgW5bAxmCzxCMY
mNKXFjSZViKr7k/fFdE4A6a42lQOEM4pl8yVsQd85+S2+AQptFZWbfrXkUuvwv76
6b62fShOYXvlMZiw6Ygvd61EZiQqhxI8XKf3zKFpIrI6pqojAgMBAAGjgYMwgYAw
DgYDVR0PAQH/BAQDAgO4MB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAd
BgNVHQ4EFgQU96hmBUCLr1jfQG+JvMUnjb5cwvAwHwYDVR0jBBgwFoAUxgbXWOda
CbEVUC+9GbmLLbLwD0kwDwYDVR0RBAgwBocEAAAAADANBgkqhkiG9w0BAQsFAAOC
AgEAWPhNZO7wt2Uwo7g1Ngq/3ihBG1PfKfKyc2AUKdY/cxzvckIF1cQltsaGcGkh
hXk1H8CNxIsrit3ssWKrJDvuqZjpE/k9EpHX53tv3UTWzaHB75az/w0KfQMdtJ+9
7WxqWoof92u7EKIRtD/5jxcF2837ZA6x3BMoZBCm9Sc21JArtYwAYSstoyyMkDbb
QKwULmZpemFVJVGb0XQ//NWC7gFYpOjfN0DOJSREkwfE84MyXl+qOQ9NJcvHCrOQ
eBlTK6k+qiqruxiA7FSJxCkybLziEsmmwiIM4MOBx3fWgGiWJXQwata6Rn+tzDih
DHtaA0f8J5WXTIPwyt2s4aKJpxvijKRaA+K8nLCW9GLURR7RqSdwzkp6wx5ipQW9
tEyOQv9M8qlvafF2jaU3RBBcliLWn765jOM+u0OnaAHmklrSxywrZTDgbbn47b8P
3medxiXCBMNMrGzkQqfshITC4VyhkDE/WzRXQ6lV3smCAYNkQhzm92Lo7X/pGRG1
vl9aq+XxHjKNaxsc67tNtE+yK6jJSAsLBfIWI7X42tgNzNbKX1d7BUCY5kHnptQg
rPKBqtv4KD14DqoSzF6JmLIY0AWEq0HHgObVz0skEy+w9aHvUp7z5kzYuAFa/RI3
t/TKz4OySwlajE5xlTPgrkBiZXuHfY7pSOG9Dxira6Xe1mc=
-----END CERTIFICATE-----
`
	TestServerPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAoO0hZdfHLbKhPRpfR5rP6OPNjs1tWuFtWhgGjwjqwrGwEimQ
kDmPiLxTtLIDHGAfJYq2S3cdARpPvXzP7hys6lai77/CvXCnUnwySCD7XFr3DsKv
PgoY6rnwVFVQ0Z5QS7SIOjDoohw9j7wI489TDHKgcwOtY+tfHidXLtVE1NrT0gmu
y25o/KAFv0TM0gG5onfuo0gx20HH7aemeP8fRozAoFuWwMZgs8QjGJjSlxY0mVYi
q+5P3xXROAOmuNpUDhDOKZfMlbEHfOfktvgEKbRWVm3615FLr8L++um+tn0oTmF7
5TGYsOmIL3etRGYkKocSPFyn98yhaSKyOqaqIwIDAQABAoIBABmXXDZL6DrWK1oC
NaC3d9S7VVceSAOp0bAAHhrk+iBYDX316i1lWfQOrukPXftHNezMcEzz9kLUolWp
4Y8mNEFX4bVqs1dY8OLnKT/bLt3zAVLxltiw1mGNjDB9GMsc9/vyC6/lUzlkcE3J
Q0inEbfrCqT/srUvGwM6kly6QaXvni05BQN+4uS4p6jjoH6nkewpx0XaYRQ4F9LY
wZWjVEWSi8eDlZVWu8koNrO70i6qNMj7lpKU0hZXZ1IFGWTP4MfwEf6Z4n9ocfCb
gUl6dUpocM3A9o4Kfbk5uzLz8lxda2V660n6hllv6scjw9YEP4EV9VLDwRb5pN1i
bRzXkrECgYEAybvQ8mWR7RFZsxi1QR36ru9zRAeYNHkTTIz2+jrodBvJ05VU1Eoe
zneQKxzoQl2kO9IE7rXv6274tjK/3CZQDKoSxJmNEeGdBcsSnKlsU65JK+qw2zOi
U3j0G48JNIH8BM4TsVEK9Bjsvh4onIs0l/66d2chfpavEJD+RzHKyv0CgYEAzDco
BukhctcYO4XorUyQlTHF2Nh3RkXno9XLVo49ugfDVbsc1vnQwtMFvmhlAiM69pZ5
14ABhSC8uhuCS4hW84kcf2bHQEv0iEbAD64lyAqt2abLgps2W09V+g1UM4mljCsc
NBTFBRRrUcd1NoHf6RYZLvtTuogBhK9JQENaI58CgYEAi1k3ThknIdD4WyRYH/Dr
dudkgbuVQbnYwOomuFb0ty9yzLq8bB//A7PHXGCNdzpj9gZu7c2zOrffCUwpB5NX
fEgGytMehRmJc7UA2EKX133ugW2OWPxjxrEoPdkiDKk1QsRvCe7nWBHXhsQiXXAz
FkMY3t3YXy8LIrBlVRxp7qkCgYEApwaNxGk1JFpsxXJWxjcTIhOdgCg8FcvjE4sv
TlH0ho0G5L2vbtzQNCioT/3Ob5slBL46VVmq5JnMAmOxg9m1VGbWWhVT7nCxRiyn
tat31090tcnINcCBCtmutl/keGqibixsWuSJ6Ae1ZyO96KD85AVg/54r8yp+I2nC
fb8YoH0CgYEAuFXii84Llf/aB+8ijSecj02vkROM2EFENPnJH7Q5XLr0u90eenid
JZjhLMtJrPoEC6qbEJGeVP5lnvZ9XGnkDQE8lK/sSftBJlhTC4iuY2wWpjGZ8lyI
5YMUz9AKT49KMOdp04CDmMLSARuUSj0wRvmgTXS3lEwtS/lfYIwu6OA=
-----END RSA PRIVATE KEY-----`

	// ServerName is encoded in the above certificates.
	ServerName = "0.0.0.0"

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
		listener, err := net.Listen("tcp", fmt.Sprint(ServerName, ":0"))
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
