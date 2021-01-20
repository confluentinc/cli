package test_server

import (
	"net/http/httptest"
	"testing"
)

// TestBackend consists of the servers for necessary mocked backend services
// Each server is instantiated with its router type (<type>_router.go) that has routes and handlers defined
type TestBackend struct {
	cloud       	*httptest.Server
	kafkaApi       	*httptest.Server
	kafkaRestProxy	*httptest.Server
	mds         	*httptest.Server
	sr				*httptest.Server
}

func StartTestBackend(t *testing.T) *TestBackend {
	cloudRouter := NewCloudRouter(t)
	kafkaRouter := NewKafkaRouter(t)
	mdsRouter := NewMdsRouter(t)
	srRouter := NewSRRouter(t)
	kafkaRPServer := configureKafkaRestServer(kafkaRouter.KafkaRP)

	backend := &TestBackend{
		cloud:      	httptest.NewServer(cloudRouter),
		kafkaApi:       httptest.NewServer(kafkaRouter.KafkaApi),
		kafkaRestProxy: kafkaRPServer,
		mds:         	httptest.NewServer(mdsRouter),
		sr: 		 	httptest.NewServer(srRouter),
	}
	cloudRouter.kafkaApiUrl = backend.kafkaApi.URL
	cloudRouter.srApiUrl = backend.sr.URL
	cloudRouter.kafkaRPUrl = backend.kafkaRestProxy.URL
	return backend
}

//var kafkaRestPort *string // another test uses port 8090
func configureKafkaRestServer(router KafkaRestProxyRouter) *httptest.Server {
	kafkaRPServer := httptest.NewUnstartedServer(router)
	kafkaRPServer.StartTLS()
	return kafkaRPServer
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

func (b *TestBackend) GetMdsUrl() string {
	return b.mds.URL
}
// Creates and returns new TestBackend struct with passed CloudRouter and KafkaRouter
// Use this to spin up a backend for a ccloud cli test that requires non-default endpoint behavior or needs additional endpoints
// Define/override the endpoints on the corresponding routers
func NewCloudTestBackendFromRouters(cloudRouter *CloudRouter, kafkaRouter *KafkaRouter) *TestBackend {
	ccloud := &TestBackend{
		cloud:       	httptest.NewServer(cloudRouter),
		kafkaApi:       httptest.NewServer(kafkaRouter.KafkaApi),
		kafkaRestProxy: configureKafkaRestServer(kafkaRouter.KafkaRP),
	}
	cloudRouter.kafkaApiUrl = ccloud.kafkaApi.URL
	cloudRouter.kafkaRPUrl = ccloud.kafkaRestProxy.URL
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
