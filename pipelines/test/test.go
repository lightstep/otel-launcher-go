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
	// The certificates and keys used in these tests expire in
	// June 2042.  They were generated using the
	// github.com/stripe/certstrap utility.

	TestCARootCertificate = `-----BEGIN CERTIFICATE-----
MIIE5DCCAsygAwIBAgIBATANBgkqhkiG9w0BAQsFADARMQ8wDQYDVQQDEwZUZXN0
Q0EwIBcNMjIwNjIyMDUzNzM2WhgPMjA1MjA2MjIwNTQ3MzFaMBExDzANBgNVBAMT
BlRlc3RDQTCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAKd9bueE77Hr
+CP8egmg7LbD+ev81KZ3p6WnkUuLnIy0h79SEd0VSOsuhzIwv8wirdWq5pn+0Enu
WwnCRSYTDMNPW1xqyG5CRihNrOAKSuB7b3wZ1Mc2xWPZaZmZ5G4EXDCFkzoYWZTA
JznasLQUzB2H6xXZ0f/5ECVHHGdPq3VX9ExsQEAN4y/S2lTndo48yiJl6/7j9vH3
+fFWbUQFVae47gh5x1piNt7c8KXVy9+j9UfnpRhEYqSxMMlCHTAHa8DAMsVLCXaa
JwryYlTuInmkHmyEa6swCtMR9/XZwrk/QxUq/fQwi7n0Hqb3gUGBadmNPPnDOdwr
i66HRcook8M33FKztgZ7ICfmU5h1ixSd1BxRWvpvrgRAIuayPV94CtWsVGGZ90ac
P8IIyp3zsQq0ZRaizr2jdyv9wNhKn0mmf/5GLZvb4jn8r5zk7XV+KhfsPs+g4sZD
NkdnaOmrT7V//FeW35GZcB0ArkmwITKQqeta5DCG1kk2gM/BX2eKZFRJfmtgKBCS
Ekn2EK3eSlQKtziG7bh8Yk8kkdzRRM/Oo1WXpzyj3T/0fmyQvgUp5xeOMPdsn5rm
x3sywxDOU3+OcrpgkO6Uvj9qygpi0UtAYQbgm+RinN0HxStTrJfpZAvtMDb2AHS4
Z7wTv+65zllui9DAtWsuO28VzNrXqp77AgMBAAGjRTBDMA4GA1UdDwEB/wQEAwIB
BjASBgNVHRMBAf8ECDAGAQH/AgEAMB0GA1UdDgQWBBStokraXYub8zBHkpsn7MwY
5LYr6TANBgkqhkiG9w0BAQsFAAOCAgEAN4zYgi90Mo/sAWu+CBnP49KkFmTbIOol
r+7yJSC6c0pwJz6yvP+y8u+iAkYn23+zztNbO/8tY2PfD5FeLk2s7esdy4YORgNQ
8eMnG1cY0WzLu13n4o4fIof1uUX2YswJGbUjQKV1a9GprOKLZYJXduKMUYjDrtYv
IRnBY6+KvCMc4Kb46Av88LKoCSRfEQ8KcMh7ZV7HMeE6kLx+gyxwsBMUtTR5Ahlh
6ibAJ60T8gBEdw+ukYILBIPqyLrSCemRQBkTWR9cRkaxU4Eipd0nNb/p8U4qTUbg
88syJrE/EfrEwmBp0tpMo9JlSdrRrtzTfwyBsQ+NYPxD8ecPIVbZpKS3ilKvDqaY
2FUS2/0ka4viwlDWFLcm35VbCJPlTIbayTonwtyPnB61RkZxW5Yk7hXuc8n+MDxI
w/tM1+gRvMGtCu1orPJvleREUG+Hvt11YPwmZg6B6bFpFueDB7NTECGXvVd6T7R5
REroLU6QwgsgYhTF17GI6mItBYe0Pnlfg4LKDNeZHJz0GX5UETPQhxB3DPFq9s/9
6zjGAatoGlUBdd9z4/P1m5kSSLaDp0SDoLOgiYjC60fFiqXmNsAEQlnmltCAaRFS
mNXYPediH+u+P9u4E8PHjq7myxXOFVxKiUzGdU7/aP4nDtgcG/sNQnYikSs/SoXl
+am5C45fOPE=
-----END CERTIFICATE-----`

	TestServerPublicCertificate = `-----BEGIN CERTIFICATE-----
MIIENzCCAh+gAwIBAgIRANV2ZXO0f9cuDMbQHenQx1owDQYJKoZIhvcNAQELBQAw
ETEPMA0GA1UEAxMGVGVzdENBMCAXDTIyMDYyMjA1NDEwMVoYDzIwNTIwNjIyMDU0
NzMxWjAVMRMwEQYDVQQDEwpUZXN0U2VydmVyMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAwj/2a+aT+oVYsVRIo8sNQOBXLaFxM8Dicyxq/21bRe9yVnaq
i+vjB+7bBORpca5+KsrmasH9YYL3n5N1RQVDOOxfKP26YhH6Blna0DlTYo8dmO6e
24XVAIKPMcaUCdjg13tGeKSfY9MySzByOXFYOfWKMMUISlUUMHJBiduBnbgRfISg
xyj5ZhcwIh3nQeQh1VXraa66Gu8ovf5iyIur2hFIsMgAS0hze1gR4ZTeffSmoRZC
BFb46iSahVs+I+p7FEX0s9NIxhd/ROW/mi6bz4bW2o/YdYmoQZgMsdEpmb9X0htS
r8FtSDq3ItEOf1fLgE3XD2Oa3KWNSdBamNIMLwIDAQABo4GDMIGAMA4GA1UdDwEB
/wQEAwIDuDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwHQYDVR0OBBYE
FI4ejbIHUEaUF2KOT3ue8O+eGsiHMB8GA1UdIwQYMBaAFK2iStpdi5vzMEeSmyfs
zBjktivpMA8GA1UdEQQIMAaHBH8AAAEwDQYJKoZIhvcNAQELBQADggIBAAEGfhuJ
SbhMKtIBejAHUpi3ZlwAzSrvMTRPkQuBHHmGrgxjs7qnacg1RFM9RsUnJTnZjrOH
+szXBFry6K3t1OnyT9BznnKz1ceYArOp+S/oqrDe92rL4iN/DQlYfI+iCt9i+Cz8
DrX5neJzizSGtH5dVvroEqsr9tSJ0tvPS2WOK1qRbhtSw+WlI27y5UsfC7PEPhos
Z3XE2CocHNyEmB3V4KRCgkeZKirKTwDxrT3bC4+pS3PDWdINYBiGnqIfqDiHhxX9
Ir498HnKl43SEJLrDLmlt8llQGhdAYbbJ7ub1SZQOpLVS0yNR3dZ0uMiQGbvixBk
ECe4JeBIoeuEhLWtAJOkwNLTV2dd6j0DZdfR/9rx4zai4eHreuX2mKb5ACtrL3gC
nY8/8aTqHR+R44+Z+mjFmTeWfVsSBnmw1REwMlUL8cSw1RwRvX5N96WMZnCZ51P/
6olro4Fv5z7+C7Sjit2OX6POn+Ny694rFDyREkMI2/WbknJ2+FM1DSaxP1tAIAHD
a2R/fF6nSXP9Qz2BUGIUA9h9jDkGyooF8UqCP2zOcTWh0YW/CV+K2JD4UxbMuTGN
KXTlg41Od1DdvVHf8zqbhFyUUxR7PsxgWYlgylxWiT8UsO03L+jyN0zoLynar25O
wmPSy4wouDxuQTO+dVQcLZwsI52nu7PZZaVx
-----END CERTIFICATE-----`

	TestServerPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAwj/2a+aT+oVYsVRIo8sNQOBXLaFxM8Dicyxq/21bRe9yVnaq
i+vjB+7bBORpca5+KsrmasH9YYL3n5N1RQVDOOxfKP26YhH6Blna0DlTYo8dmO6e
24XVAIKPMcaUCdjg13tGeKSfY9MySzByOXFYOfWKMMUISlUUMHJBiduBnbgRfISg
xyj5ZhcwIh3nQeQh1VXraa66Gu8ovf5iyIur2hFIsMgAS0hze1gR4ZTeffSmoRZC
BFb46iSahVs+I+p7FEX0s9NIxhd/ROW/mi6bz4bW2o/YdYmoQZgMsdEpmb9X0htS
r8FtSDq3ItEOf1fLgE3XD2Oa3KWNSdBamNIMLwIDAQABAoIBAHmrCJUTCpL63M/N
g+Yb88Q0AEbTfQ02fmA3bRlqDKZkUVB46V/UsxIv+L06uBT9f4ccKXCq6yMdni40
dVpy7mUEIKKTMh/lNJ6vv0926JSuIZK9u4CydfTo0foScH0ue75cN4qvSiqDiVfx
E0qJhQJgmlrrvsKYQZoKpqRLegcnwVHUjpKMgwbeMDSWr1JxrddQUZ9heZt4aKg7
yOD3EL1iQf9FiXcjnS8VujEOJ5J0aasC6XTYASg2YZmHMboMUdtcMzboKBpQt5MT
98Kpo67PJG+UYfbnXyHWMyXNu9OsO3qNjFDv2O95lUUBnoIkoyxekkT0oCYU1Tnx
KJV5jrkCgYEA2Usvcbq6TmwigGgKajzNSq7zKoc4qsGhat3M/RPXD2LVOghfF3HE
QX65lNr+Egu1dH27RdzvAO0PGYBJkKcPfD5C0xQJXpYds9bEQAsI1e5BwOL8q/Fe
qnSq2H3EDILJ4z4dpxUPFkcrse+kRWYqXgLEdIinMIZgneO4qcnqQWsCgYEA5Nn0
YqX17tBp2aMe1UCaiKOrNXuvEtqkK/62HU1hYAm6kt+eJUIIVMTz/p7ZjLZS88CU
M5Qqu6ZM6n9Xl8ojIg+670OhYlEyzX7o5bO4yfMWp2cfBEIaZvh+ChBJMqi14uVe
wpuGatcb7NSWYL01JxZJoCuaedlPRfS0i+pp3U0CgYBwc5RuCvB3vUZtpWoeaLDl
QXzeOXR+Cg77OyXmoundMIygp8xuWZXzPx3ThzGNLToOuzK7iQa3N/dkfxuTHKHK
7n2utuPSa2WbuD1/1zYPYGnu5IlWgmc3V4FC4HMg9l58l5zI5wETymk2gIpG0ASE
+nGozT+YwTInA76BP9lXWQKBgQDKs0KDHfx3SqJ24sSsnkxCOrWq6aJoUMCZN0KX
MbLOHc/jx62L0rEOZGS5YnnO6K8Qt8KM7O/sxZ/bFG/BQolb4hLxWjXXn5Qf8AjZ
bBaAyY+HNw+B9grsqaz5vPMYq9Zu4jrMpHSqrV1Op/2KMgyiUltkQzrQMmrEy7of
M8IRAQKBgHr3qpZqCFkqY97HxbT5Nc4YOVoNiUy+RJ7TIryS/pSAWhXuefyXXZEO
3cg2zN/X8j3PozlyqyInPTJbbOanKFdCKPpq0um0HOMfONLPz8CpODhIzOR7b7/l
6KqtKg5elMQerqqtPOdcitn6qAxfnOjoFyfDhLxsbV1PIlTja1Oa
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
