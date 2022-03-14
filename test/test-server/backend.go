package test_server

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestCloudURL is used to hardcode a specific port (1024) so tests can identify CCloud URLs
var TestCloudURL = url.URL{Scheme: "http", Host: "127.0.0.1:1024"}

// TestBackend consists of the servers for necessary mocked backend services
// Each server is instantiated with its router type (<type>_router.go) that has routes and handlers defined
type TestBackend struct {
	cloud          *httptest.Server
	kafkaApi       *httptest.Server
	kafkaRestProxy *httptest.Server
	mds            *httptest.Server
	sr             *httptest.Server
}

func StartTestBackend(t *testing.T, isAuditLogEnabled bool) *TestBackend {
	cloudRouter := NewCloudRouter(t, isAuditLogEnabled)
	kafkaRouter := NewKafkaRouter(t)
	mdsRouter := NewMdsRouter(t)
	srRouter := NewSRRouter(t)
	kafkaRPServer := configureKafkaRestServer(kafkaRouter.KafkaRP)

	backend := &TestBackend{
		cloud:          newTestCloudServer(cloudRouter),
		kafkaApi:       httptest.NewServer(kafkaRouter.KafkaApi),
		kafkaRestProxy: kafkaRPServer,
		mds:            httptest.NewServer(mdsRouter),
		sr:             httptest.NewServer(srRouter),
	}

	cloudRouter.kafkaApiUrl = backend.kafkaApi.URL
	cloudRouter.srApiUrl = backend.sr.URL
	cloudRouter.kafkaRPUrl = backend.kafkaRestProxy.URL

	return backend
}

func StartTestCloudServer(t *testing.T, isAuditLogEnabled bool) *TestBackend {
	cloudRouter := NewCloudRouter(t, isAuditLogEnabled)
	backend := &TestBackend{
		cloud: newTestCloudServer(cloudRouter),
	}
	fmt.Println("starting backend server " + backend.GetCloudUrl())
	return backend
}

//var kafkaRestPort *string // another test uses port 8090
func configureKafkaRestServer(router KafkaRestProxyRouter) *httptest.Server {
	return httptest.NewServer(router)
}

func newTestCloudServer(handler http.Handler) *httptest.Server {
	server := httptest.NewUnstartedServer(handler)

	// Stop the old listener
	if err := server.Listener.Close(); err != nil {
		panic(err)
	}

	// Create a new listener with the hardcoded port
	l, err := net.Listen("tcp", TestCloudURL.Host)
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
	if b.kafkaApi != nil {
		b.kafkaApi.Close()
	}
	if b.kafkaRestProxy != nil {
		b.kafkaRestProxy.Close()
	}
	if b.mds != nil {
		b.mds.Close()
	}
	if b.sr != nil {
		b.sr.Close()
	}
}

func (b *TestBackend) GetCloudUrl() string {
	return b.cloud.URL
}

func (b *TestBackend) GetKafkaApiUrl() string {
	return b.kafkaApi.URL
}

func (b *TestBackend) GetKafkaRestUrl() string {
	return b.kafkaRestProxy.URL + "/kafka"
}

func (b *TestBackend) GetMdsUrl() string {
	return b.mds.URL
}
