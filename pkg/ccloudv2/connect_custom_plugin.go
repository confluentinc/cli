package ccloudv2

import (
	"context"
	ccpv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"strings"
)

func newCcpClient(url, userAgent string, unsafeTrace bool) *ccpv1.APIClient {
	cfg := ccpv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClient(unsafeTrace)
	cfg.Servers = ccpv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return ccpv1.NewAPIClient(cfg)
}

func (c *Client) ccpApiContext() context.Context {
	return context.WithValue(context.Background(), ccpv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) GetPresignedUrl(extension string) (ccpv1.ConnectV1PresignedUrl, error) {
	presignedUrlRequest := *ccpv1.NewConnectV1PresignedUrlRequest()
	presignedUrlRequest.SetContentFormat(extension)
	resp, httpResp, err := c.CcpClient.PresignedUrlsConnectV1Api.PresignedUploadUrlConnectV1PresignedUrl(c.ccpApiContext()).ConnectV1PresignedUrlRequest(presignedUrlRequest).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateCustomPlugin(displayName string, description string, documentationLink string, connectorClass string, connectorType string, sensitivePropertiesString string, uploadId string) (ccpv1.ConnectV1CustomConnectorPlugin, error) {
	createCustomPluginRequest := ccpv1.NewConnectV1CustomConnectorPlugin()
	var sensitiveProperties []string
	if len(sensitivePropertiesString) > 0 {
		sensitiveProperties = strings.Split(sensitivePropertiesString, ",")
	}
	createCustomPluginRequest.SetDisplayName(displayName)
	createCustomPluginRequest.SetDescription(description)
	createCustomPluginRequest.SetDocumentationLink(documentationLink)
	createCustomPluginRequest.SetConnectorClass(connectorClass)
	createCustomPluginRequest.SetConnectorType(connectorType)
	createCustomPluginRequest.SetSensitiveConfigProperties(sensitiveProperties)
	createCustomPluginRequest.SetUploadSource(
		ccpv1.ConnectV1UploadSourcePresignedUrlAsConnectV1CustomConnectorPluginUploadSourceOneOf(
			ccpv1.NewConnectV1UploadSourcePresignedUrl("PRESIGNED_URL_LOCATION", uploadId)))
	resp, httpResp, err := c.CcpClient.CustomConnectorPluginsConnectV1Api.CreateConnectV1CustomConnectorPlugin(c.ccpApiContext()).ConnectV1CustomConnectorPlugin(*createCustomPluginRequest).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListCustomPlugins() (ccpv1.ConnectV1CustomConnectorPluginList, error) {
	resp, httpResp, err := c.CcpClient.CustomConnectorPluginsConnectV1Api.ListConnectV1CustomConnectorPlugins(c.ccpApiContext()).PageSize(ccloudV2ListPageSize).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DescribeCustomPlugin(id string) (ccpv1.ConnectV1CustomConnectorPlugin, error) {
	resp, httpResp, err := c.CcpClient.CustomConnectorPluginsConnectV1Api.GetConnectV1CustomConnectorPlugin(c.ccpApiContext(), id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteCustomPlugin(id string) error {
	httpResp, err := c.CcpClient.CustomConnectorPluginsConnectV1Api.DeleteConnectV1CustomConnectorPlugin(c.ccpApiContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateCustomPlugin(id string, name string, description string, documentationLink string, sensitivePropertiesString string) (ccpv1.ConnectV1CustomConnectorPlugin, error) {
	updateCustomPluginRequest := ccpv1.NewConnectV1CustomConnectorPluginUpdate()
	var sensitiveProperties []string
	if len(sensitivePropertiesString) > 0 {
		sensitiveProperties = strings.Split(sensitivePropertiesString, ",")
		updateCustomPluginRequest.SetSensitiveConfigProperties(sensitiveProperties)
	}
	updateCustomPluginRequest.SetDisplayName(name)
	updateCustomPluginRequest.SetDescription(description)
	updateCustomPluginRequest.SetDocumentationLink(documentationLink)
	resp, httpResp, err := c.CcpClient.CustomConnectorPluginsConnectV1Api.UpdateConnectV1CustomConnectorPlugin(c.ccpApiContext(), id).ConnectV1CustomConnectorPluginUpdate(*updateCustomPluginRequest).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
