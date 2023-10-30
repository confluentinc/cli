package ccloudv2

import (
	"context"
	"net/http"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

const StatusProvisioning = "PROVISIONING"

func newCmkClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *cmkv2.APIClient {
	cfg := cmkv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = cmkv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return cmkv2.NewAPIClient(cfg)
}

func (c *Client) cmkApiContext() context.Context {
	return context.WithValue(context.Background(), cmkv2.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) CreateKafkaCluster(cluster cmkv2.CmkV2Cluster) (cmkv2.CmkV2Cluster, *http.Response, error) {
	return c.CmkClient.ClustersCmkV2Api.CreateCmkV2Cluster(c.cmkApiContext()).CmkV2Cluster(cluster).Execute()
}

func (c *Client) DescribeKafkaCluster(clusterId, environment string) (cmkv2.CmkV2Cluster, *http.Response, error) {
	req := c.CmkClient.ClustersCmkV2Api.GetCmkV2Cluster(c.cmkApiContext(), clusterId).Environment(environment)
	return c.CmkClient.ClustersCmkV2Api.GetCmkV2ClusterExecute(req)
}

func (c *Client) UpdateKafkaCluster(clusterId string, update cmkv2.CmkV2ClusterUpdate) (cmkv2.CmkV2Cluster, error) {
	resp, httpResp, err := c.CmkClient.ClustersCmkV2Api.UpdateCmkV2Cluster(c.cmkApiContext(), clusterId).CmkV2ClusterUpdate(update).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteKafkaCluster(clusterId, environment string) (*http.Response, error) {
	return c.CmkClient.ClustersCmkV2Api.DeleteCmkV2Cluster(c.cmkApiContext(), clusterId).Environment(environment).Execute()
}

func (c *Client) ListKafkaClusters(environment string) ([]cmkv2.CmkV2Cluster, error) {
	var list []cmkv2.CmkV2Cluster

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListClusters(pageToken, environment)
		if err != nil {
			return nil, errors.CatchCCloudV2Error(err, httpResp)
		}
		list = append(list, page.GetData()...)

		pageToken, done, err = extractNextPageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *Client) executeListClusters(pageToken, environment string) (cmkv2.CmkV2ClusterList, *http.Response, error) {
	req := c.CmkClient.ClustersCmkV2Api.ListCmkV2Clusters(c.cmkApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
