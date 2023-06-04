package testserver

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	// TestCloudUrl is used to hardcode a specific port (1024) so tests can identify CCloud URLs
	TestCloudUrl          = url.URL{Scheme: "http", Host: "127.0.0.1:1024"}
	TestV2CloudUrl        = url.URL{Scheme: "http", Host: "127.0.0.1:2048"}
	TestHubUrl            = url.URL{Scheme: "http", Host: "127.0.0.1:4096"}
	TestKafkaRestProxyUrl = url.URL{Scheme: "http", Host: "127.0.0.1:1025"}
	TestFlinkGatewayUrl   = url.URL{Scheme: "http", Host: "127.0.0.1:1026"}
)

// TestBackend consists of the servers for necessary mocked backend services
// Each server is instantiated with its router type (<type>_router.go) that has routes and handlers defined
type TestBackend struct {
	cloud          *httptest.Server
	v2Api          *httptest.Server
	kafkaRestProxy *httptest.Server
	flinkGateway   *httptest.Server
	mds            *httptest.Server
	sr             *httptest.Server
	hub            *httptest.Server
}

func StartTestBackend(t *testing.T, isAuditLogEnabled bool) *TestBackend {
	cloudRouter := NewCloudRouter(t, isAuditLogEnabled)
	cloudV2Router := NewV2Router(t)

	backend := &TestBackend{
		cloud:          newTestCloudServer(cloudRouter, TestCloudUrl.Host),
		v2Api:          newTestCloudServer(cloudV2Router, TestV2CloudUrl.Host),
		kafkaRestProxy: newTestCloudServer(NewKafkaRestProxyRouter(t), TestKafkaRestProxyUrl.Host),
		flinkGateway:   newTestCloudServer(NewFlinkGatewayRouter(t), TestFlinkGatewayUrl.Host),
		mds:            httptest.NewServer(NewMdsRouter(t)),
		sr:             httptest.NewServer(NewSRRouter(t)),
		hub:            newTestCloudServer(NewHubRouter(t), TestHubUrl.Host),
	}

	cloudRouter.srApiUrl = backend.sr.URL
	cloudV2Router.srApiUrl = backend.sr.URL

	return backend
}

func StartTestCloudServer(t *testing.T, isAuditLogEnabled bool) *TestBackend {
	router := NewCloudRouter(t, isAuditLogEnabled)
	return &TestBackend{cloud: newTestCloudServer(router, TestCloudUrl.Host)}
}

func newTestCloudServer(handler http.Handler, address string) *httptest.Server {
	server := httptest.NewUnstartedServer(handler)

	// Stop the old listener
	if err := server.Listener.Close(); err != nil {
		panic(err)
	}

	// Create a new listener with the hardcoded port
	l, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	server.Listener = l

	server.Start()

	return server
}

func (b *TestBackend) Close() {
	if b.cloud != nil {
		b.cloud.Close()
	}
	if b.v2Api != nil {
		b.v2Api.Close()
	}
	if b.kafkaRestProxy != nil {
		b.kafkaRestProxy.Close()
	}
	if b.flinkGateway != nil {
		b.flinkGateway.Close()
	}
	if b.mds != nil {
		b.mds.Close()
	}
	if b.sr != nil {
		b.sr.Close()
	}
	if b.hub != nil {
		b.hub.Close()
	}
}

func (b *TestBackend) GetCloudUrl() string {
	return b.cloud.URL
}

func (b *TestBackend) GetKafkaRestUrl() string {
	return b.kafkaRestProxy.URL + "/kafka"
}

func (b *TestBackend) GetMdsUrl() string {
	return b.mds.URL
}
