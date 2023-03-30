package ccloudv2

import (
	"context"
	"net/http"

	srcm "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"
)

func newSchemaRegistryClient(url, userAgent string, unsafeTrace bool) *srcm.APIClient {
	cfg := srcm.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = srcm.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return srcm.NewAPIClient(cfg)
}

func (c *Client) SchemaRegistryApiContext() context.Context {
	return context.WithValue(context.Background(), srcm.ContextAccessToken, c.AuthToken)
}

func (c *Client) GetSchemaRegistryClusterByEnvironment(environment string) (srcm.SrcmV2ClusterList, *http.Response, error) {
	return c.SchemaRegistryClient.ClustersSrcmV2Api.ListSrcmV2Clusters(c.SchemaRegistryApiContext()).Environment(environment).Execute()
}

func (c *Client) GetSchemaRegistryClusterById(clusterId, environment string) (srcm.SrcmV2Cluster, *http.Response, error) {
	return c.SchemaRegistryClient.ClustersSrcmV2Api.GetSrcmV2Cluster(c.SchemaRegistryApiContext(), clusterId).Environment(environment).Execute()
}

func (c *Client) GetStreamGovernanceRegionById(regionId string) (srcm.SrcmV2Region, *http.Response, error) {
	return c.SchemaRegistryClient.RegionsSrcmV2Api.GetSrcmV2Region(c.SchemaRegistryApiContext(), regionId).Execute()
}

func (c *Client) DeleteSchemaRegistryCluster(clusterId, environment string) (*http.Response, error) {
	return c.SchemaRegistryClient.ClustersSrcmV2Api.DeleteSrcmV2Cluster(c.SchemaRegistryApiContext(), clusterId).Environment(environment).Execute()
}

func (c *Client) ExtractNextPageToken(nextPageUrlStringNullable srcm.NullableString) (string, bool, error) {
	if !nextPageUrlStringNullable.IsSet() {
		return "", true, nil
	}
	nextPageUrlString := *nextPageUrlStringNullable.Get()
	pageToken, err := extractPageToken(nextPageUrlString)
	return pageToken, false, err
}
