package ccloudv2

import (
	"context"
	"net/http"

	srcm "github.com/confluentinc/ccloud-sdk-go-v2/srcm/v2"

	"github.com/confluentinc/cli/internal/pkg/errors"
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

func (c *Client) ListSchemaRegistryRegions(cloud, packageType string) ([]srcm.SrcmV2Region, error) {
	var list []srcm.SrcmV2Region

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListSchemaRegistryRegions(cloud, packageType, pageToken)
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

func (c *Client) executeListSchemaRegistryRegions(cloud, packageType, pageToken string) (srcm.SrcmV2RegionList, *http.Response, error) {
	req := c.SchemaRegistryClient.RegionsSrcmV2Api.ListSrcmV2Regions(c.SchemaRegistryApiContext())
	if cloud != "" {
		req = req.SpecCloud(cloud)
	}
	if packageType != "" {
		req = req.SpecPackages([]string{packageType})
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
