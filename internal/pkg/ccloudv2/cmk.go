package ccloudv2

import (
	"context"
	"net/http"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
)

func NewV2CmkClient(baseURL string, isTest bool) *cmkv2.APIClient {
	cmkServer := getV2ServerUrl(baseURL, isTest)
	server := cmkv2.ServerConfigurations{
		{URL: cmkServer, Description: "Confluent Cloud CMK"},
	}
	cfg := &cmkv2.Configuration{
		DefaultHeader:    make(map[string]string),
		UserAgent:        "OpenAPI-Generator/1.0.0/go",
		Debug:            false,
		Servers:          server,
		OperationServers: map[string]cmkv2.ServerConfigurations{},
	}
	return cmkv2.NewAPIClient(cfg)
}

func (c *Client) cmkApiContext() context.Context {
	auth := context.WithValue(context.Background(), cmkv2.ContextAccessToken, c.AuthToken)
	return auth
}

func (c *Client) CreateKafkaCluster(cluster cmkv2.CmkV2Cluster) (cmkv2.CmkV2Cluster, *http.Response, error) {
	req := c.CmkClient.ClustersCmkV2Api.CreateCmkV2Cluster(c.cmkApiContext()).CmkV2Cluster(cluster)
	return c.CmkClient.ClustersCmkV2Api.CreateCmkV2ClusterExecute(req)
}

func (c *Client) DescribeKafkaCluster(clusterId, environment string) (cmkv2.CmkV2Cluster, *http.Response, error) {
	req := c.CmkClient.ClustersCmkV2Api.GetCmkV2Cluster(c.cmkApiContext(), clusterId).Environment(environment)
	return c.CmkClient.ClustersCmkV2Api.GetCmkV2ClusterExecute(req)
}

func (c *Client) ListKafkaClusters(environment string) (cmkv2.CmkV2ClusterList, *http.Response, error) {
	req := c.CmkClient.ClustersCmkV2Api.ListCmkV2Clusters(c.cmkApiContext()).Environment(environment)
	return c.CmkClient.ClustersCmkV2Api.ListCmkV2ClustersExecute(req)
}

func (c *Client) UpdateKafkaCluster(clusterId string, update cmkv2.CmkV2ClusterUpdate) (cmkv2.CmkV2Cluster, *http.Response, error) {
	req := c.CmkClient.ClustersCmkV2Api.UpdateCmkV2Cluster(c.cmkApiContext(), clusterId).CmkV2ClusterUpdate(update)
	return c.CmkClient.ClustersCmkV2Api.UpdateCmkV2ClusterExecute(req)
}

func (c *Client) DeleteKafkaCluster(clusterId, environment string) (*http.Response, error) {
	req := c.CmkClient.ClustersCmkV2Api.DeleteCmkV2Cluster(c.cmkApiContext(), clusterId).Environment(environment)
	return c.CmkClient.ClustersCmkV2Api.DeleteCmkV2ClusterExecute(req)
}
