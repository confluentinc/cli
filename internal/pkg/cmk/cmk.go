package cmk

import (
	"context"
	"net/http"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	// cmkmock "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2/mock"
)

func CmkApiContext(authToken string) context.Context {
	auth := context.WithValue(context.Background(), cmkv2.ContextAccessToken, authToken)
	return auth
}

func CreateKafkaCluster(client *cmkv2.APIClient, cluster cmkv2.CmkV2Cluster, authToken string) (cmkv2.CmkV2Cluster, *http.Response, error) {
	return client.ClustersCmkV2Api.CreateCmkV2Cluster(CmkApiContext(authToken)).CmkV2Cluster(cluster).Execute()
}

func DescribeKafkaCluster(client *cmkv2.APIClient, clusterId string, environment string, authToken string) (cmkv2.CmkV2Cluster, *http.Response, error) {
	return client.ClustersCmkV2Api.GetCmkV2Cluster(CmkApiContext(authToken), clusterId).Environment(environment).Execute()
}

func ListKafkaClusters(client *cmkv2.APIClient, environment string, authToken string) (cmkv2.CmkV2ClusterList, *http.Response, error) {
	return client.ClustersCmkV2Api.ListCmkV2Clusters(CmkApiContext(authToken)).Environment(environment).Execute()
}

func UpdateKafkaCluster(client *cmkv2.APIClient, clusterId string, update cmkv2.CmkV2ClusterUpdate, authToken string) (cmkv2.CmkV2Cluster, *http.Response, error) {
	return client.ClustersCmkV2Api.UpdateCmkV2Cluster(CmkApiContext(authToken), clusterId).CmkV2ClusterUpdate(update).Execute()
}

func DeleteKafkaCluster(client *cmkv2.APIClient, clusterId string, environment string, authToken string) (*http.Response, error) {
	return client.ClustersCmkV2Api.DeleteCmkV2Cluster(CmkApiContext(authToken), clusterId).Environment(environment).Execute()
}
