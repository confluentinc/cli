package ccloudv2

import (
	"context"

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

func (c *Client) ListSchemaRegistryCloudRegions(cloud, packageType string) ([]srcm.SrcmV2Region, error) {
	regionListRequest := c.SchemaRegistryClient.RegionsSrcmV2Api.ListSrcmV2Regions(c.SchemaRegistryApiContext())

	if cloud != "" {
		regionListRequest = regionListRequest.SpecCloud(cloud)
	}

	if packageType != "" {
		packageSpec := []string{packageType}
		regionListRequest = regionListRequest.SpecPackages(packageSpec)
	}

	var regionList []srcm.SrcmV2Region
	done := false
	pageToken := ""
	for !done {
		regionListRequest = regionListRequest.PageToken(pageToken)
		regionPage, _, err := c.SchemaRegistryClient.RegionsSrcmV2Api.ListSrcmV2RegionsExecute(regionListRequest)
		if err != nil {
			return nil, err
		}
		regionList = append(regionList, regionPage.GetData()...)

		pageToken, done, err = extractNextPageToken(regionPage.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return regionList, nil
}
