package cmk

import (
	"context"
	"net/http"

	cmk "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
)

func CmkApiContext(authToken string) context.Context {
	auth := context.WithValue(context.Background(), cmk.ContextAccessToken, authToken)
	return auth
}

func CreateKafkaCluster(client *cmk.APIClient, cluster cmk.CmkV2Cluster, authToken string) (cmk.CmkV2Cluster, *http.Response, error) {
	return client.ClustersCmkV2Api.CreateCmkV2Cluster(CmkApiContext(authToken)).CmkV2Cluster(cluster).Execute()
}

func DescribeKafkaCluster(client *cmk.APIClient, clusterId string, environment string, authToken string) (cmk.CmkV2Cluster, *http.Response, error) {
	return client.ClustersCmkV2Api.GetCmkV2Cluster(CmkApiContext(authToken), clusterId).Environment(environment).Execute()
}

func ListKafkaClusters(client *cmk.APIClient, environment string, authToken string) (cmk.CmkV2ClusterList, *http.Response, error) {
	return client.ClustersCmkV2Api.ListCmkV2Clusters(CmkApiContext(authToken)).Environment(environment).Execute()
}

func UpdateKafkaCluster(client *cmk.APIClient, clusterId string, update cmk.CmkV2ClusterUpdate, authToken string) (cmk.CmkV2Cluster, *http.Response, error) {
	return client.ClustersCmkV2Api.UpdateCmkV2Cluster(CmkApiContext(authToken), clusterId).CmkV2ClusterUpdate(update).Execute()
}

func DeleteKafkaCluster(client *cmk.APIClient, clusterId string, environment string, authToken string) (*http.Response, error) {
	return client.ClustersCmkV2Api.DeleteCmkV2Cluster(CmkApiContext(authToken), clusterId).Environment(environment).Execute()
}
