package kafka

import (
	"context"
	"net/http"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
)

func ListKafkaClusters(client *ccloud.Client, environmentId string) ([]*schedv1.KafkaCluster, error) {
	req := &schedv1.KafkaCluster{AccountId: environmentId}
	return client.Kafka.List(context.Background(), req)
}

func ListCmkKafkaClusters(client *cmkv2.APIClient, auth context.Context, environment string) (cmkv2.CmkV2ClusterList, *http.Response, error) {
	return client.ClustersCmkV2Api.ListCmkV2Clusters(auth).Environment(environment).Execute() // not sure if backgrond ctx works
}

// write cmk client mock funcs here...
func MockCmkConfiguration() *cmkv2.Configuration {
	server := cmkv2.ServerConfigurations{
		{URL: "CLI integration test", Description: "Confluent Cloud"},
	}
	cfg := &cmkv2.Configuration{
		DefaultHeader:    make(map[string]string),
		UserAgent:        "OpenAPI-Generator/1.0.0/go",
		Debug:            false,
		Servers:          server,
		OperationServers: map[string]cmkv2.ServerConfigurations{},
	}
	return cfg
}
