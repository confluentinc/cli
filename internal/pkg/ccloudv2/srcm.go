package ccloudv2

import (
	"context"
	"net/http"

	srcmv2 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newSrcmClient(url, userAgent string, unsafeTrace bool) *srcmv2.APIClient {
	cfg := srcmv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = srcmv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return srcmv2.NewAPIClient(cfg)
}

func (c *Client) srcmApiContext() context.Context {
	return context.WithValue(context.Background(), srcmv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateSchemaRegistryCluster(srcmV2Cluster srcmv2.SrcmV2Cluster) (srcmv2.SrcmV2Cluster, *http.Response, error) {
	req := c.SchemaRegistryClient.ClustersSrcmV2Api.CreateSrcmV2Cluster(c.srcmApiContext()).SrcmV2Cluster(srcmV2Cluster)
	return c.SchemaRegistryClient.ClustersSrcmV2Api.CreateSrcmV2ClusterExecute(req)
}

func (c *Client) DeleteSchemaRegistryCluster(id, environment string) (*http.Response, error) {
	req := c.SchemaRegistryClient.ClustersSrcmV2Api.DeleteSrcmV2Cluster(c.srcmApiContext(), id).Environment(environment)
	return c.SchemaRegistryClient.ClustersSrcmV2Api.DeleteSrcmV2ClusterExecute(req)
}

func (c *Client) GetSchemaRegistryCluster(id, environment string) (srcmv2.SrcmV2Cluster, *http.Response, error) {
	req := c.SchemaRegistryClient.ClustersSrcmV2Api.GetSrcmV2Cluster(c.srcmApiContext(), id).Environment(environment)
	return c.SchemaRegistryClient.ClustersSrcmV2Api.GetSrcmV2ClusterExecute(req)
}

func (c *Client) UpdateSchemaRegistryCluster(id string, srcmV2ClusterUpdate srcmv2.SrcmV2ClusterUpdate) (srcmv2.SrcmV2Cluster, *http.Response, error) {
	req := c.SchemaRegistryClient.ClustersSrcmV2Api.UpdateSrcmV2Cluster(c.srcmApiContext(), id).SrcmV2ClusterUpdate(srcmV2ClusterUpdate)
	return c.SchemaRegistryClient.ClustersSrcmV2Api.UpdateSrcmV2ClusterExecute(req)
}

func (c *Client) ListSchemaRegistryClusters(environment string) ([]srcmv2.SrcmV2Cluster, error) {
	var list []srcmv2.SrcmV2Cluster

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListSchemaRegistryClusters(environment, pageToken)
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

func (c *Client) executeListSchemaRegistryClusters(environment, pageToken string) (srcmv2.SrcmV2ClusterList, *http.Response, error) {
	req := c.SchemaRegistryClient.ClustersSrcmV2Api.ListSrcmV2Clusters(c.srcmApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.SchemaRegistryClient.ClustersSrcmV2Api.ListSrcmV2ClustersExecute(req)
}
