package ccloudv2

import (
	"context"
	"net/http"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newConnectCustomPluginClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *connectcustompluginv1.APIClient {
	cfg := connectcustompluginv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = connectcustompluginv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return connectcustompluginv1.NewAPIClient(cfg)
}

func (c *Client) connectCustomPluginApiContext() context.Context {
	return context.WithValue(context.Background(), connectcustompluginv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) GetPresignedUrl(request connectcustompluginv1.ConnectV1PresignedUrlRequest) (connectcustompluginv1.ConnectV1PresignedUrl, error) {
	resp, httpResp, err := c.ConnectCustomPluginClient.PresignedUrlsConnectV1Api.PresignedUploadUrlConnectV1PresignedUrl(c.connectCustomPluginApiContext()).ConnectV1PresignedUrlRequest(request).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateCustomPlugin(createCustomPluginRequest connectcustompluginv1.ConnectV1CustomConnectorPlugin) (connectcustompluginv1.ConnectV1CustomConnectorPlugin, error) {
	resp, httpResp, err := c.ConnectCustomPluginClient.CustomConnectorPluginsConnectV1Api.CreateConnectV1CustomConnectorPlugin(c.connectCustomPluginApiContext()).ConnectV1CustomConnectorPlugin(createCustomPluginRequest).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListCustomPlugins(cloud string) ([]connectcustompluginv1.ConnectV1CustomConnectorPlugin, error) {
	var list []connectcustompluginv1.ConnectV1CustomConnectorPlugin

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListPlugins(pageToken, cloud)
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

func (c *Client) DescribeCustomPlugin(id string) (connectcustompluginv1.ConnectV1CustomConnectorPlugin, error) {
	resp, httpResp, err := c.ConnectCustomPluginClient.CustomConnectorPluginsConnectV1Api.GetConnectV1CustomConnectorPlugin(c.connectCustomPluginApiContext(), id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteCustomPlugin(id string) error {
	httpResp, err := c.ConnectCustomPluginClient.CustomConnectorPluginsConnectV1Api.DeleteConnectV1CustomConnectorPlugin(c.connectCustomPluginApiContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateCustomPlugin(id string, updateCustomPluginRequest connectcustompluginv1.ConnectV1CustomConnectorPluginUpdate) (connectcustompluginv1.ConnectV1CustomConnectorPlugin, error) {
	resp, httpResp, err := c.ConnectCustomPluginClient.CustomConnectorPluginsConnectV1Api.UpdateConnectV1CustomConnectorPlugin(c.connectCustomPluginApiContext(), id).ConnectV1CustomConnectorPluginUpdate(updateCustomPluginRequest).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) executeListPlugins(pageToken, cloud string) (connectcustompluginv1.ConnectV1CustomConnectorPluginList, *http.Response, error) {
	req := c.ConnectCustomPluginClient.CustomConnectorPluginsConnectV1Api.ListConnectV1CustomConnectorPlugins(c.connectCustomPluginApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	if cloud != "" {
		req = req.Cloud(cloud)
	}
	return req.Execute()
}

func (c *Client) CreateCustomPluginVersion(createCustomPluginVersionRequest connectcustompluginv1.ConnectV1CustomConnectorPluginVersion, id string) (connectcustompluginv1.ConnectV1CustomConnectorPluginVersion, error) {
	resp, httpResp, err := c.ConnectCustomPluginClient.CustomConnectorPluginVersionsConnectV1Api.CreateConnectV1CustomConnectorPluginVersion(c.connectCustomPluginApiContext(), id).ConnectV1CustomConnectorPluginVersion(createCustomPluginVersionRequest).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DescribeCustomPluginVersion(pluginId, versionId string) (connectcustompluginv1.ConnectV1CustomConnectorPluginVersion, error) {
	resp, httpResp, err := c.ConnectCustomPluginClient.CustomConnectorPluginVersionsConnectV1Api.GetConnectV1CustomConnectorPluginVersion(c.connectCustomPluginApiContext(), pluginId, versionId).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListCustomPluginVersions(pluginId string) (connectcustompluginv1.ConnectV1CustomConnectorPluginVersionList, error) {
	resp, httpResp, err := c.ConnectCustomPluginClient.CustomConnectorPluginVersionsConnectV1Api.ListConnectV1CustomConnectorPluginVersions(c.connectCustomPluginApiContext(), pluginId).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteCustomPluginVersion(pluginId, versionId string) error {
	httpResp, err := c.ConnectCustomPluginClient.CustomConnectorPluginVersionsConnectV1Api.DeleteConnectV1CustomConnectorPluginVersion(c.connectCustomPluginApiContext(), pluginId, versionId).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateCustomPluginVersion(pluginId, versionId string, versionUpdate connectcustompluginv1.ConnectV1CustomConnectorPluginVersion) (connectcustompluginv1.ConnectV1CustomConnectorPluginVersion, error) {
	resp, httpResp, err := c.ConnectCustomPluginClient.CustomConnectorPluginVersionsConnectV1Api.UpdateConnectV1CustomConnectorPluginVersion(c.connectCustomPluginApiContext(), pluginId, versionId).ConnectV1CustomConnectorPluginVersion(versionUpdate).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
