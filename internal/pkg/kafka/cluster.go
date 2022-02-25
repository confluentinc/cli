package kafka

import (
	"context"
	"net/http"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
)

func cmkApiContext(authToken string) context.Context {
	auth := context.WithValue(context.Background(), cmkv2.ContextAccessToken, authToken)
	return auth
}

func ListCmkKafkaClusters(client *cmkv2.APIClient, environment string, authToken string) (cmkv2.CmkV2ClusterList, *http.Response, error) {
	req := client.ClustersCmkV2Api.ListCmkV2Clusters(cmkApiContext(authToken)).Environment(environment)
	return client.ClustersCmkV2Api.ListCmkV2ClustersExecute(req)
}

func ListKafkaClusters(client *ccloud.Client, environmentId string) ([]*schedv1.KafkaCluster, error) {
	req := &schedv1.KafkaCluster{AccountId: environmentId}
	return client.Kafka.List(context.Background(), req)
}
