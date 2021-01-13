package test_server

import (
	"net/http/httptest"
	"testing"
)

// TestBackend consists of the servers for necessary mocked backend services
// Each server is instantiated with its router type (<type>_router.go) that has routes and handlers defined
type TestBackend struct {
	cloud       *httptest.Server
	kafka       *httptest.Server
	mds         *httptest.Server
	sr			*httptest.Server
}

func StartTestBackend(t *testing.T) *TestBackend {
	cloudRouter := NewCloudRouter(t)
	kafkaRouter := NewKafkaRouter(t)
	mdsRouter := NewMdsRouter(t)
	srRouter := NewSRRouter(t)
	backend := &TestBackend{
		cloud:       httptest.NewServer(cloudRouter),
		kafka:       httptest.NewServer(kafkaRouter),
		mds:         httptest.NewServer(mdsRouter),
		sr: 		 httptest.NewServer(srRouter),
	}
	cloudRouter.kafkaApiUrl = backend.kafka.URL
	cloudRouter.srApiUrl = backend.sr.URL
	return backend
}

func (b *TestBackend) Close() {
	if b.cloud != nil {
		b.cloud.Close()
	}
	if b.kafka != nil {
		b.kafka.Close()
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

func (b *TestBackend) GetKafkaUrl() string {
	return b.kafka.URL
}

func (b *TestBackend) GetMdsUrl() string {
	return b.mds.URL
}
// Creates and returns new TestBackend struct with passed CloudRouter and KafkaRouter
// Use this to spin up a backend for a ccloud cli test that requires non-default endpoint behavior or needs additional endpoints
// Define/override the endpoints on the corresponding routers
func NewCloudTestBackendFromRouters(cloudRouter *CloudRouter, kafkaRouter *KafkaRouter) *TestBackend {
	ccloud := &TestBackend{
		cloud:       httptest.NewServer(cloudRouter),
		kafka:       httptest.NewServer(kafkaRouter),
	}
	cloudRouter.kafkaApiUrl = ccloud.kafka.URL
	return ccloud
}
// Creates and returns new TestBackend struct with passed MdsRouter
// Use this to spin up a backend for a confluent cli test that requires non-default endpoint behavior or needs additional endpoints
// Define/override the endpoints on the mdsRouter
func NewConfluentTestBackendFromRouter(mdsRouter *MdsRouter) *TestBackend {
	confluent := &TestBackend{
		mds:       httptest.NewServer(mdsRouter),
	}
	return confluent
}
