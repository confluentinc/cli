package cmk

import (
	"context"
	"fmt"
	"net/http"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
)

func cmkApiContext(authToken string) context.Context {
	auth := context.WithValue(context.Background(), cmkv2.ContextAccessToken, authToken)
	return auth
}

func CreateKafkaCluster(client *cmkv2.APIClient, cluster cmkv2.CmkV2Cluster, authToken string) (cmkv2.CmkV2Cluster, *http.Response, error) {
	req := client.ClustersCmkV2Api.CreateCmkV2Cluster(cmkApiContext(authToken)).CmkV2Cluster(cluster)
	return client.ClustersCmkV2Api.CreateCmkV2ClusterExecute(req)
}

func DescribeKafkaCluster(client *cmkv2.APIClient, clusterId string, environment string, authToken string) (cmkv2.CmkV2Cluster, *http.Response, error) {
	req := client.ClustersCmkV2Api.GetCmkV2Cluster(cmkApiContext(authToken), clusterId).Environment(environment)
	return client.ClustersCmkV2Api.GetCmkV2ClusterExecute(req)
}

func ListKafkaClusters(client *cmkv2.APIClient, environment string, authToken string) (cmkv2.CmkV2ClusterList, *http.Response, error) {
	req := client.ClustersCmkV2Api.ListCmkV2Clusters(cmkApiContext(authToken)).Environment(environment)
	fmt.Println("created the list request")
	return client.ClustersCmkV2Api.ListCmkV2ClustersExecute(req)
}

func UpdateKafkaCluster(client *cmkv2.APIClient, clusterId string, update cmkv2.CmkV2ClusterUpdate, authToken string) (cmkv2.CmkV2Cluster, *http.Response, error) {
	req := client.ClustersCmkV2Api.UpdateCmkV2Cluster(cmkApiContext(authToken), clusterId).CmkV2ClusterUpdate(update)
	return client.ClustersCmkV2Api.UpdateCmkV2ClusterExecute(req)
}

func DeleteKafkaCluster(client *cmkv2.APIClient, clusterId string, environment string, authToken string) (*http.Response, error) {
	req := client.ClustersCmkV2Api.DeleteCmkV2Cluster(cmkApiContext(authToken), clusterId).Environment(environment)
	return client.ClustersCmkV2Api.DeleteCmkV2ClusterExecute(req)
}
