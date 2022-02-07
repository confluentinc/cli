package mock

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/confluentinc/ccloud-sdk-go-v1/mock"
	cmk "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	org "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
)

func NewClientMock() *ccloud.Client {
	return &ccloud.Client{
		Params:         nil,
		Auth:           &mock.Auth{},
		Account:        &mock.Account{},
		Kafka:          &mock.Kafka{},
		SchemaRegistry: &mock.SchemaRegistry{},
		Connect:        &mock.Connect{},
		User:           &mock.User{},
		APIKey:         &mock.APIKey{},
		KSQL:           &mock.KSQL{},
		MetricsApi:     &mock.MetricsApi{},
		UsageLimits:    &mock.UsageLimits{},
	}
}

type ClustersCmkV2ApiService struct {
	ListFunc func(ctx context.Context, cluster *cmk.CmkV2Cluster) ([]*cmk.CmkV2Cluster, error)
}

func NewCmkClientMock() *cmk.APIClient {
	server := cmk.ServerConfigurations{
		{URL: "mock.server", Description: "Confluent mock test"},
	}
	cfg := &cmk.Configuration{
		DefaultHeader:    make(map[string]string),
		UserAgent:        "OpenAPI-Generator/1.0.0/go",
		Debug:            false,
		Servers:          server,
		OperationServers: map[string]cmk.ServerConfigurations{},
	}
	return cmk.NewAPIClient(cfg)
}

func NewOrgClientMock() *org.APIClient {
	server := org.ServerConfigurations{
		{URL: "mock.server", Description: "Confluent mock test"},
	}
	cfg := &org.Configuration{
		DefaultHeader:    make(map[string]string),
		UserAgent:        "OpenAPI-Generator/1.0.0/go",
		Debug:            false,
		Servers:          server,
		OperationServers: map[string]org.ServerConfigurations{},
	}
	return org.NewAPIClient(cfg)
}
