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
Q0EwHhcNMjUwOTA5MjExNjE3WhcNMjcwMzA5MjEyNjE0WjARMQ8wDQYDVQQDEwZU
ZXN0Q0EwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQC4Vp8zAb7Zkzzg
TvcVGO1woI3le1D4DPtPGU2N7XO7DFx6AA5rR9krBwMTpRwH+Toslxoo88RQ13KJ
WGU97uhjJRFSTAzYHK0W9fMhAfRYIIeQdpEWWZ1fcQcMwJQawkxA9Laoie9qh2SH
0SAoUdGRtNLvggh+8gO2p0bdl6RbKLIFhh0evFoYCRqLn1y6v0emoPfBszbNJIuo
OmLMFT/20+xopkjbOAdj4j4NdBjnzGDzicK8PcpCCP+JFwNIwcN5n8gC47y45oaz
1sV7ruuIoEaX7ZlUlKethMuJnaezGrhWgL/ppJWNSPVl28Uw+MvuK0/TrVmH0DAO
UD5sr/Qc254Ic1qmiHFeLYcjmKc5Hg1SwwvebPD5HytjdAkhuKJBhwFJ7n1jgvVh
/sc0VpOcbsmd/uCUS+AcVoDmXljK/SxcIpfywXicQdSTZzJckzsyJPNbNFKEroFL
ka4Rv46o2s0vrrpiwggh2VKQBrUw12XHY83/XOpkiB05A033I1zxyzso6083Ytc4
N4BhoGwB73WN6bSXuNqFKKUtsil/I3ff4J39Ot7+GxwmTkktMamXPYEPykpwwKPB
tI8aSgMbN27h4BIKUhOAUAM2F0CzMJtlAgRkMODB7HVv5PGgf7UAsBybkrXPhBZF
xJkhn/iFiCZSPlemT/1NaDsTDxyqTQIDAQABo0UwQzAOBgNVHQ8BAf8EBAMCAQYw
EgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQUftMmfJTZcoiAkiViOKixXY3j
rPowDQYJKoZIhvcNAQELBQADggIBAB7A7uRFXEwkS4ggqlqPmVDC6LeWq+ZHD+Qp
A7GpV19WWovM3cngMIVrUMBP7HVyP2gYgbwukL9Qsx8HMy9RYih5gJ0tGy9msWau
UZ2X+ew8f0u+hiK03XoFQ8sJ5IcKtiJVSNlqIuylOHHsZ9eE+/Lhul2PtB5CeL+k
rFvwRNpsyjPB7DBjtvLmPJYH6HybrhDjKMpe8yOF58DOdaYndc7OgEv8XJXCIlDE
yYQJnVF9aIf1CJNu7WnGIMbo8gjoG9zTN+j4Tqr+yOaLhkHGZIis3qpe5/pnLJWY
Tdu6Ym7IagD62HD88hYnI5yE217GpxpAs9vydgTQITkdSIa06XSgjtAGV4VwDXPb
RH28cGPVQRbqp+sqV88LP3Js2bzgF5HbLScl0e9SZWP67mADP8tZotTG2jA4NPY6
OXvIUjA9WYmkbbKL6QsVzEOtH93vH0y5WSqhkSJ/IXgGRHrkaJ0VWQQrF0fMNWRX
00GHDbvyw/6XbZvdPHP7JnbAwkzUOW5V1Rcl9GOyKseYqR2Ylypl22Q7dC+8a9j+
XkmBCzm19+7IkXlEBdn/5cpKtN4qCD3ChoTcZxHcYLE00XlyyD2x6Jdi7itvFtrX
njkQPbV+Gk/CTPbEi5X9R5YX+dpuPafSnXutk9WFc6HvjXHP4s+y3Z7VQTmiTDBg
rstjOKJ9
-----END CERTIFICATE-----
`

	// From ./out/TestServer.crt.  Also copy this file to ./testdata/testserver.crt.
	TestServerPublicCertificate = `-----BEGIN CERTIFICATE-----
MIIENTCCAh2gAwIBAgIRAM5PjvZRtpOF3dqJfOJR06UwDQYJKoZIhvcNAQELBQAw
ETEPMA0GA1UEAxMGVGVzdENBMB4XDTI1MDkwOTIxMTYzMFoXDTI3MDMwOTIxMjYx
NFowFTETMBEGA1UEAxMKVGVzdFNlcnZlcjCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBANmpkP9Vs1khVSp+/RZBDu/3dHRci+0CqIW9qvf0KXTPywBZq6Qs
ldRE4BiOKPBIheLP8Ne6Z3sDrNVkrKDQ0tX1TLPbrBaMPgyUyAjLGTCl5G2epjY5
xoT8ugXFRpk6XzF1q7W6SzsD6aDyMDGgDgx4HeaKGGP6V2/AET6izQYy4AgRVP4x
d20xB0mEZXO6axr32d6D4nOn8i4pwCsI7FzxemQL2hS+9Dg0/UWunKtpDLDAu2oM
CW1qtzWJ75I6wGaGSj9fmiwLJm+HM+xsIsgCWmNfIJcEvAr4ms9bn5TvAGTjwQAT
uTYxKC3uLQM0KIfE6yvNhirxa/Wx1vlwouUCAwEAAaOBgzCBgDAOBgNVHQ8BAf8E
BAMCA7gwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMB0GA1UdDgQWBBSy
S6/BJq4mAZ8YJeQVhjSgJjFRKDAfBgNVHSMEGDAWgBR+0yZ8lNlyiICSJWI4qLFd
jeOs+jAPBgNVHREECDAGhwR/AAABMA0GCSqGSIb3DQEBCwUAA4ICAQCuAw6Kk8Z6
Oxxaz1hi5AWf21S/zgjOpLpUki+4eIG93QcJvaBPhqBUybmnUs1eXyV0cYoUU4w/
AZxim3p+2ffuqTLlCV514e1VDNJG2zMQSETueIb94Y6RvWiavBBWjAidniOdOuHD
bU5KSYyoz0rOGDdi3+n4IxljGOprQ02M3u1nxRuLkPE+R2SdZRVyV92b9H5kjNlt
lJMEOxwGtiBGPn2Sr0Q2hZH6r48VnvRCkiBqwzYDO/gDItYuFV8aj47eAbTS3FMf
XSdIujFz2MVA3+DQ/XrM6BnsL45Tqv5USE+e1+iu2bKWGuJQyBNBO8U2BykSVMZP
g7lbQXyUpMTs+Bvd8CWq+s5z3g4zpHoQ6XbOUcASTZHVSxI0x8TCbQNCRy81ZKAR
+RWGe7opmWA3wmTSek3++iaHF6vHDO3Hzmj5T+mtHYAVO7T08ignsgorDZRL5QV5
D1I9pdy2c6UDDyYmi4k4/iLOSc9sgXUSfO90Mb9bwQc21Bpm5sLDXF1zu4d2lW4e
x5zeFbXLhN7S6y0fYJ3beFVvzR0w3rkhoKxo75y4nFysPJAIVwZ2ALeVtIf0rRwx
wl+36R6f4MTt+yX1DZgrw1mlYO/YGBlqQoc7tzJlZM560JF1MijLvFMY6l1mjqiG
V0I7qb1/w3AhX1AWs+VYdrCAtVtbZLUD/g==
-----END CERTIFICATE-----
`

	// From ./out/TestServer.key. Also copy this file to ./testdata/testserver.key.
	TestServerPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEA2amQ/1WzWSFVKn79FkEO7/d0dFyL7QKohb2q9/QpdM/LAFmr
pCyV1ETgGI4o8EiF4s/w17pnewOs1WSsoNDS1fVMs9usFow+DJTICMsZMKXkbZ6m
NjnGhPy6BcVGmTpfMXWrtbpLOwPpoPIwMaAODHgd5ooYY/pXb8ARPqLNBjLgCBFU
/jF3bTEHSYRlc7prGvfZ3oPic6fyLinAKwjsXPF6ZAvaFL70ODT9Ra6cq2kMsMC7
agwJbWq3NYnvkjrAZoZKP1+aLAsmb4cz7GwiyAJaY18glwS8Cviaz1uflO8AZOPB
ABO5NjEoLe4tAzQoh8TrK82GKvFr9bHW+XCi5QIDAQABAoIBAEegf9kunBbsG6kg
4es1XJOGNJnoLJnBorBkMoNnI09P5AdI+u1LSMDyK2aZPYqY6woxEJoFOvFly7Xr
a819+quzDdswImjHHNIqEcum/jKenNWf/CVjDNuNiS/F9A4Pgez2QpvKYuBYUMUI
feiOuaDL1FcdwZoA804Qf/xDOcHcV7AKKzw8YzLnS+46rtqwFoUTtTvv05QACtjp
eZFQvnLAd/7qGSlAcxYlVDhQaL+h+rG0e+A/h636JS8WHfFupXrXcIxtbdKCwmco
YWqdH99z+z6wW4e4DiU1BFcvvaZJfyn7fzWKQXDFQmsit1vVOVwdxZsyyN03geri
h3i6lo8CgYEA6CA4fGHaABeDvutMoVDAwsLHdLJdDJ2qsjXErUC+twXtB7Fr8YfW
NvxrLri5CPt1Y8QuXRPb4YGwCX2OWXhW+hpZkN+dJ7Ox4wqL6Sfz8znhKKsI8S4H
+xdMlBKyeFiDEOKNYQikHq5Mq93gIrEVX8X6WpvH6SK4Bu+bYr5HifsCgYEA8AyH
Ikjc0skgifmD3xDbH/QGaxOCIRnx8HhB6JIN9Szd97egeKnCLLS/TaHHN4Oksetp
i7O6FVNynMDaKk6D9v7VwIP+aanAclAKAqNMK8YKUuU6kQGpX2eke3FlehcR8p0L
KBxPs9mLsZyHBCLDSrhvmalXrX8EchsodoHS0J8CgYBEBM0IhZPf2wQb+c8mpgcW
CwVvSKTDgZ/3QJI1QnegIfhm/LJowCkhS64MrsxpuWWYqm/7jkosNlhjL4t65Cx4
dSgxr5TZgWpq4ThGRhLR/u/fft7L7XUhOp6R9Tie0zD0za4n9ORCqUiGRCndgI6G
1fiafHOD+Ux7m9KoiKFl6QKBgFj6C8zlfRSUgH8kAYFZWh+J8CcYYA+s8kTUDnoK
SSorq0r6wXx4UAUKKi64XINzRES+oayqvbrR55W61iMAX2HaK5jkVBUOWssEZ/F6
Xe2Lxp/bX84H86PtYsZuzdJnYruvAken1tMvO9xlzJX33LOBkbw/TMR+ZEN3VZQ3
otC5AoGAMXLa7MH3XtTvhOZAhbgL5uuOVyywRaZHNN+bmpNF+rkottYLW4GJgHUf
v9UxAnCLoyHydvnj7xmDcnEWmMVTT+YNUsfh/yXApNheIl/j3OkKHEZbFsSo1KKj
zSEW+a8hwO/1eqEoZibdVUYlFjXpLx77IWCdWlWVVibvET9WVzg=
-----END RSA PRIVATE KEY-----
`

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
