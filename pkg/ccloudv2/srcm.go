package ccloudv2

import (
	"context"
	"net/http"

	srcmv3 "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v3"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newSrcmClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *srcmv3.APIClient {
	cfg := srcmv3.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = srcmv3.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return srcmv3.NewAPIClient(cfg)
}

func (c *Client) srcmApiContext() context.Context {
	return context.WithValue(context.Background(), srcmv3.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) GetSchemaRegistryClusterById(clusterId, environment string) (srcmv3.SrcmV3Cluster, error) {
	cluster, httpResp, err := c.SrcmClient.ClustersSrcmV3Api.GetSrcmV3Cluster(c.srcmApiContext(), clusterId).Environment(environment).Execute()
	return cluster, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetSchemaRegistryClustersByEnvironment(environment string) ([]srcmv3.SrcmV3Cluster, error) {
	var list []srcmv3.SrcmV3Cluster

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

func (c *Client) executeListSchemaRegistryClusters(environment, pageToken string) (srcmv3.SrcmV3ClusterList, *http.Response, error) {
	req := c.SrcmClient.ClustersSrcmV3Api.ListSrcmV3Clusters(c.srcmApiContext()).Environment(environment)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.SrcmClient.ClustersSrcmV3Api.ListSrcmV3ClustersExecute(req)
}
