package ccloudv2

import (
	"context"
	"net/http"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/cdx/v1"
	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"
	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newStreamShareClient(baseURL, userAgent string, isTest bool) *cdxv1.APIClient {
	cfg := cdxv1.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = cdxv1.ServerConfigurations{{URL: getServerUrl(baseURL, isTest), Description: "Confluent Cloud Stream Sharing"}}
	cfg.UserAgent = userAgent

	return cdxv1.NewAPIClient(cfg)
}

func (c *Client) streamSharingApiContext() context.Context {
	return context.WithValue(context.Background(), cliv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) DeleteProviderShare(shareId string) (*http.Response, error) {
	request := c.StreamShareClient.ProviderSharesCdxV1Api.DeleteCdxV1ProviderShare(c.streamSharingApiContext(), shareId)
	return c.StreamShareClient.ProviderSharesCdxV1Api.DeleteCdxV1ProviderShareExecute(request)
}

func (c *Client) DescribeProvideShare(shareId string) (cdxv1.CdxV1ProviderShare, *http.Response, error) {
	request := c.StreamShareClient.ProviderSharesCdxV1Api.GetCdxV1ProviderShare(c.streamSharingApiContext(), shareId)
	return c.StreamShareClient.ProviderSharesCdxV1Api.GetCdxV1ProviderShareExecute(request)
}

func (c *Client) ListProviderShares(sharedResource, pageToken string) (cdxv1.CdxV1ProviderShareList, *http.Response, error) {
	request := c.StreamShareClient.ProviderSharesCdxV1Api.ListCdxV1ProviderShares(c.streamSharingApiContext()).
		PageToken(pageToken).SharedResource(sharedResource)
	return c.StreamShareClient.ProviderSharesCdxV1Api.ListCdxV1ProviderSharesExecute(request)
}
