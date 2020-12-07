package test_server

import (
	"net/http/httptest"
	"testing"
)

type TestBackend struct {
	cloud *httptest.Server
	cloudRouter *CloudRouter
	kafka *httptest.Server
	kafkaRouter *KafkaRouter
	mds *httptest.Server
	mdsRouter *MdsRouter
}

func StartTestBackend(t *testing.T) *TestBackend {
	cloudRouter := NewCCloudRouter(t)
	kafkaRouter := NewKafkaRouter(t)
	mdsRouter := NewMdsRouter(t)
	ccloud := &TestBackend{
		cloud:	httptest.NewServer(cloudRouter),
		cloudRouter: cloudRouter,
		kafka:	httptest.NewServer(kafkaRouter),
		kafkaRouter: kafkaRouter,
		mds: httptest.NewServer(mdsRouter),
		mdsRouter: mdsRouter,
	}
	cloudRouter.kafkaApiUrl = ccloud.kafka.URL
	return ccloud
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
}

func (b *TestBackend) GetCloudUrl() string{
	return b.cloud.URL
}

func (b *TestBackend) GetKafkaUrl() string{
	return b.kafka.URL
}

func (b *TestBackend) GetMdsUrl() string{
	return b.mds.URL
}
func NewSingleCloudTestBackend(cloudRouter *CloudRouter, kafkaRouter *KafkaRouter) *TestBackend {
	ccloud := &TestBackend{
		cloud:	httptest.NewServer(cloudRouter),
		cloudRouter: cloudRouter,
		kafka:	httptest.NewServer(kafkaRouter),
		kafkaRouter: kafkaRouter,
	}
	ccloud.cloudRouter.kafkaApiUrl = ccloud.kafka.URL
	return ccloud
}

func NewSingleConfluentTestBackend(mdsRouter *MdsRouter) *TestBackend {
	confluent := &TestBackend{
		mds: httptest.NewServer(mdsRouter),
		mdsRouter: mdsRouter,
	}
	return confluent
}
