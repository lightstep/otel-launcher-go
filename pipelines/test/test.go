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
Q0EwHhcNMjMwNTMxMjEyOTAwWhcNMjQxMjAxMjEzODU2WjARMQ8wDQYDVQQDEwZU
ZXN0Q0EwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQC6EKtFt1d/6yfA
wm91QKxr/mk6QrbuHczmtO4/XeXbt1wR2PeUxJEfOyE36Q1QZ1O/EXYQdESvmINO
xM92y4acDPbo/Bp4eIvuJfofUqZ7Ivq+6Od5TdQvji1UGo6/3kMc6Yzjea6yJsUZ
I5Rxn2b3x2m1mJDPA9ARyWGcd8kZsv3Mr5d5lLM0x73zpBmhMldmpYqIyA81UL8Q
uVaWW3ZQSQ4NAnWK+5zIquPbssnv1xJrtca0WYOXoYKzKt9Bi3djbz9tcmJrGwtG
/aAjJs6chbXclgGwdxvYrDOg0ezgeitc8CdojjI8KzZWONxZTvqCyAWHgTZnAniu
yxMg19Bd3b5vcA71WVMoslBo5sNFE8Qouk+74dIX7bPXIzR8UGCZPeOlU15SJf2E
2VfnrsuOoBlZoZAnXXWR6iDs7g4UL5+aZ69aID2M5O0m+9SEQgSKrmCV/pIThC9J
31V8lYcUp22sC2SGz0tzHO8u6wYEQ8nI/XD1JwsCaEzbdD2oR7QWwACy5XooiKzb
9H/c1b4X2AYbgYCOYVuhoNcOhD9xnFpIYRxcd1cCZEdSxjFyKhmw1rNFY/Kt+OOa
kdVHp4xbr4CTEv0mS3RIWvwzBe8XkPohZkg+A7R+NKHQyHqDWdu3ZmQayOAK4PlP
OrepmLrvTPCxJhw0Rj/8fH3TPJmT9wIDAQABo0UwQzAOBgNVHQ8BAf8EBAMCAQYw
EgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQUVwCyIfPv4BCTqHDG848rHEl8
Tl8wDQYJKoZIhvcNAQELBQADggIBACU0Gh3CUy8+wSOC2SZnE2J6fnahqaBU3qgA
YhsMZ2roLaJwDR33FCOK1ZeJ20Sz0obshJzLUhjSYJBZsuzJdyqfXZVAkI05gjJm
if+OlhQpB4yJMC6Yrp5+r2IaERMy6Fx28SsOAUB7/82qFBEFJ0ylJ8xibkL5F3VZ
Ma310xV6sOwJ/xLSxYbL2QmAWEfq7N/yKnBoGqpKaNHURS0Asf4gSZC5Jb6WQomH
9Ar00dIUKLUfpEsYHAvQofIM3dhw4tCwllXxBD31KJHa3sL23/RZDXC71NY9xhmN
vM9sqJiusdEcHGHSF9W8I5hpfi3itOpPf+Fq/CoXoo6IWw8DoLze1O8q/x8iXBWc
n59I1Gvjg+7E0L5d7sFCwDtMG9oa3qiycI5U+0NXY5TWvWpjFUribw5GdkvWFUWa
utTIaUb3lVUnOnQiCrXYc8WK8oJ5AmFQOwbwVZ2OHo1omtZj+0+wZHatu77imxHJ
DaRvw06IBR3o6O57hs33RP2v6zZkZaBRIV/YHx19aQoBncElZE3mtQjPfddMbxVW
hRb2v8CtnTgSl+1LVfqCo8BtCwg08c9M0ROHorBJcMEUwHAy+iuvXswZtaw59ggB
V99UAuFiewdswaLPWXWz8hNroHQAqWhMBDm+h5eA/XQodzF5WGgZct5mNd/mHQAX
8WS2JZrg
-----END CERTIFICATE-----`

	// From ./out/TestServer.crt.  Also copy this file to ./testdata/testserver.crt.
	TestServerPublicCertificate = `-----BEGIN CERTIFICATE-----
MIIENTCCAh2gAwIBAgIRAO0b4DlBVtR2Lx7RtKta/5AwDQYJKoZIhvcNAQELBQAw
ETEPMA0GA1UEAxMGVGVzdENBMB4XDTIzMDUzMTIxMzAyMVoXDTI0MTIwMTIxMzg1
NlowFTETMBEGA1UEAxMKVGVzdFNlcnZlcjCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAMwAxxebmqhpZiCOo22XRlpX3gRzsjZ+X2s95a5GFneiavls+4zh
dmT5pMkfJ0Bbo95pSm3BLFY0bJC+bijIcgUuF3PdNPyRCe/qF2hyWokHYz1KF6Nx
2bLLOMPDtummQoUiGCo0FWXBXcnBIw8JumCJUkFqHBgG2mgFd5uV/iaSgVFatBAo
kA4xoVs8313g2/jlQP8SzxTNltqBgnrpEPwJM3Sv7XFjD9OxzT8uAzNrAQDVbEOz
gPLDzHEvm8XQ4m0gKAAW5tK3X5o48IEoD/wWXJlJxRol3DecM6TuQbAceZi9QvW1
WGLGM9SGk3gc4NaNCOa474QJ4Ham1FHmrdECAwEAAaOBgzCBgDAOBgNVHQ8BAf8E
BAMCA7gwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMB0GA1UdDgQWBBRm
q/jiRE2ZNYHv2yX3uyNCQIdJTTAfBgNVHSMEGDAWgBRXALIh8+/gEJOocMbzjysc
SXxOXzAPBgNVHREECDAGhwR/AAABMA0GCSqGSIb3DQEBCwUAA4ICAQAnZBlM0+qk
p/fEu/kuhSv2EBVW6OfUZfkyBBnF/QIcci7YBC03oGT/P+i1gkw0yn4pdoBiPdMb
gNHuaktJP8Jk6+Pl6TPT1OSKsLbcqmHIAGw9O4I4GOO1bZRmuh/YxGcR/aOkbQVH
OGOUw12M/a+anbXOVQczd/hWSLGxrQmPPyQJvfuJm+2J8fFSCKxKHMZ1hYew9+3o
1tWp9zwUNqNrLcfXdcsMBqaA0riK8WlPU1cZnO7FIwyvdE6khaBI2HSUF6TyhcX/
GHEQYZSj8BfzwNGN/+eesuC/IWqgwjPK5gvrtN+b/O+sjP/u9gSgN3KyyFtocX6i
x09PAuY+S8LgxHAAjDajZ7JDfgNhe6DzqO7WkvRZV9rd6pnluNIzzDtnmaeFKL6D
3olLi6RF5UU32ZHCOrDMgHge4rV0fH4sO5udgE9nKgDazMV5cQ1nlim+LX3nidbB
plHkpdh8aqd2nLGHsPx+QIhjcWlV3XAumXAhAM6VRXbvSx68ZFGdXpzFYDGZTPIR
LtIAPs49m2ZmH0nsoZ3H7CtFJKM6HP7ppJUltnAQR2SRxJDEs/PNFC8kvHzcM9v+
r8WZFrdypZJi4uC58OdHzRU4wusJ6Ctns9QLCzx3n2u+4A2+onFua/+JIkBAwJ2y
sr57Fu5xq9RYzIHYIK8tQh3zyaoVdXzbeA==
-----END CERTIFICATE-----`

	// From ./out/TestServer.key. Also copy this file to ./testdata/testserver.key.
	TestServerPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAzADHF5uaqGlmII6jbZdGWlfeBHOyNn5faz3lrkYWd6Jq+Wz7
jOF2ZPmkyR8nQFuj3mlKbcEsVjRskL5uKMhyBS4Xc900/JEJ7+oXaHJaiQdjPUoX
o3HZsss4w8O26aZChSIYKjQVZcFdycEjDwm6YIlSQWocGAbaaAV3m5X+JpKBUVq0
ECiQDjGhWzzfXeDb+OVA/xLPFM2W2oGCeukQ/AkzdK/tcWMP07HNPy4DM2sBANVs
Q7OA8sPMcS+bxdDibSAoABbm0rdfmjjwgSgP/BZcmUnFGiXcN5wzpO5BsBx5mL1C
9bVYYsYz1IaTeBzg1o0I5rjvhAngdqbUUeat0QIDAQABAoIBAFl+JdjXbDthMWoq
6MrUyAot7bmqP43kVlunZbDUElsdJyyQgLHG3rdhRMbuIbv796FEM2B+k8KlcNL2
l0DQ3iQjmio76li6D8/ts6MgN1NRqFlkvYX4FfZ3bEmE/CL3ysq4sD0jaBehglFV
ptmb/OUImMsVj4mCyCFF9LiOTlYAub9IGeAg/Tgsoxwb+7Ytfsyvg4MArW+wOA8y
ar7ipHPtkBtU3Qm9GN39IGJLp1QbTg6+cIh0wsZyuIRvDUbZMx9DYJ0VFpOi6zwp
dNBnF+UW7nhzoOcLKFhouwHUu7AGObqvEM85d7YexoYPUN36861FhYH0PcFiePQQ
ev0mEHkCgYEAzz04eLEkFtOlADHSDu45bkZ3Lh0mO2BgSqlW9a5nMOtKpnuYOON+
MpHAC6e6VmoSoedEZ9XG7MEus5yAgI6DFZR+t+RZCL6/7MnmSkB6o9N3bWWoLLfU
ePJEig3/naE9AoHTCy8g1ldevA5BM/QKydNLX+T2ZRQngoI9hzuauTcCgYEA/ACi
cbikFLTA/lSKSKS2kzNTlykxXjjWcUHUtmyKrbui7dgn2r+h6n9iaSQWN9/ip5Ah
4TIg+pnJG4uVSUm7Tl1fR5t34eHyS0TSOOQFBVRjIv6vYorZmBhmyYiLiizhn0sf
eGsu9FgsUcf22Kgrpf8q6Vq3ctN36+KgH+1ztTcCgYEAmMITwkMwyvKvCXmv0Pmg
s7yVVRR/ff0IfYBdbTNlNRX7LMSl7CkkeLoeyXiVTeVaXqVOMwvNWe78McEGp7xk
u5992KclSeDxL+WTLuBghin2OllYob3PjGdoRisTZGnZwuNXYUMX/WbhmdUYEues
nCB3yvPG+7LjfLvsBqbU4fkCgYBONxA0Rb+/oX6JMcod0+nK9FJMh3+IJBIC4xDX
cgb091kRg1aTYYkq1FDCG24992JM6cJqN/nebh7qSr+SGK1nHDn3aryhwlRGolyx
Rax9Q/zlHrFm33u75k743EWbJGT+4P0qjfHr6vYOiAcIpeGuSu2RshNgM1x0PUm/
vx57rwKBgDALJHF1MAsvk9RMDlRE5SdeXdSioDN09wauX3/trXbfBmNZYWO07XQn
GkvVy1KIHbffZzLcL2rNtKNRFaCx102XfO9o8Q5zYuGIm8Zi8N/sGzPlJ4byiCZu
Pmb9AR+VWyneYMbbm/atyt99poHkHXilRfJ7HxGmrdCLwUHeFtD+
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
