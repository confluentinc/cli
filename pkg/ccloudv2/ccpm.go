package ccloudv2

import (
	"context"
	"net/http"

	ccpmv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccpm/v1"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newCCPMClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *ccpmv1.APIClient {
	cfg := ccpmv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = ccpmv1.ServerConfigurations{ccpmv1.ServerConfiguration{URL: url}}
	cfg.UserAgent = userAgent

	return ccpmv1.NewAPIClient(cfg)
}

func (c *Client) ccpmApiContext() context.Context {
	return context.WithValue(context.Background(), ccpmv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

// CCPM API methods
func (c *Client) CreateCCPMPlugin(request ccpmv1.CcpmV1CustomConnectPlugin) (ccpmv1.CcpmV1CustomConnectPlugin, error) {
	resp, httpResp, err := c.CCPMClient.CustomConnectPluginsCcpmV1Api.CreateCcpmV1CustomConnectPlugin(c.ccpmApiContext()).CcpmV1CustomConnectPlugin(request).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListCCPMPlugins(cloud, environment string) ([]ccpmv1.CcpmV1CustomConnectPlugin, error) {
	var list []ccpmv1.CcpmV1CustomConnectPlugin

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListCCPMPlugins(pageToken, cloud, environment)
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

func (c *Client) DescribeCCPMPlugin(id string) (ccpmv1.CcpmV1CustomConnectPlugin, error) {
	resp, httpResp, err := c.CCPMClient.CustomConnectPluginsCcpmV1Api.GetCcpmV1CustomConnectPlugin(c.ccpmApiContext(), id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteCCPMPlugin(id string) error {
	httpResp, err := c.CCPMClient.CustomConnectPluginsCcpmV1Api.DeleteCcpmV1CustomConnectPlugin(c.ccpmApiContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateCCPMPlugin(id string, request ccpmv1.CcpmV1CustomConnectPluginUpdate) (ccpmv1.CcpmV1CustomConnectPlugin, error) {
	resp, httpResp, err := c.CCPMClient.CustomConnectPluginsCcpmV1Api.UpdateCcpmV1CustomConnectPlugin(c.ccpmApiContext(), id).CcpmV1CustomConnectPluginUpdate(request).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) executeListCCPMPlugins(pageToken, cloud, environment string) (ccpmv1.CcpmV1CustomConnectPluginList, *http.Response, error) {
	req := c.CCPMClient.CustomConnectPluginsCcpmV1Api.ListCcpmV1CustomConnectPlugins(c.ccpmApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	if cloud != "" {
		req = req.SpecCloud(cloud)
	}
	if environment != "" {
		req = req.Environment(environment)
	}
	return req.Execute()
}

func (c *Client) CreateCCPMPresignedUrl(request ccpmv1.CcpmV1PresignedUrl) (ccpmv1.CcpmV1PresignedUrl, error) {
	resp, httpResp, err := c.CCPMClient.PresignedUrlsCcpmV1Api.CreateCcpmV1PresignedUrl(c.ccpmApiContext()).CcpmV1PresignedUrl(request).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateCCPMPluginVersion(pluginId string, request ccpmv1.CcpmV1CustomConnectPluginVersion) (ccpmv1.CcpmV1CustomConnectPluginVersion, error) {
	resp, httpResp, err := c.CCPMClient.CustomConnectPluginVersionsCcpmV1Api.CreateCcpmV1CustomConnectPluginVersion(c.ccpmApiContext(), pluginId).CcpmV1CustomConnectPluginVersion(request).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DescribeCCPMPluginVersion(pluginId, versionId string) (ccpmv1.CcpmV1CustomConnectPluginVersion, error) {
	resp, httpResp, err := c.CCPMClient.CustomConnectPluginVersionsCcpmV1Api.GetCcpmV1CustomConnectPluginVersion(c.ccpmApiContext(), pluginId, versionId).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListCCPMPluginVersions(pluginId string) (ccpmv1.CcpmV1CustomConnectPluginVersionList, error) {
	resp, httpResp, err := c.CCPMClient.CustomConnectPluginVersionsCcpmV1Api.ListCcpmV1CustomConnectPluginVersions(c.ccpmApiContext(), pluginId).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteCCPMPluginVersion(pluginId, versionId string) error {
	httpResp, err := c.CCPMClient.CustomConnectPluginVersionsCcpmV1Api.DeleteCcpmV1CustomConnectPluginVersion(c.ccpmApiContext(), pluginId, versionId).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}
