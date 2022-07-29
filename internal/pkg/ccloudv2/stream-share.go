package ccloudv2

import (
	"context"
	"net/http"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/cdx/v1"
	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"
	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newCdxClient(baseURL, userAgent string, isTest bool) *cdxv1.APIClient {
	cfg := cdxv1.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = cdxv1.ServerConfigurations{{URL: getServerUrl(baseURL, isTest), Description: "Confluent Cloud Stream Sharing"}}
	cfg.UserAgent = userAgent

	return cdxv1.NewAPIClient(cfg)
}

func (c *Client) cdxApiContext() context.Context {
	return context.WithValue(context.Background(), cliv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) DeleteProviderShare(shareId string) (*http.Response, error) {
	request := c.StreamShareClient.ProviderSharesCdxV1Api.DeleteCdxV1ProviderShare(c.cdxApiContext(), shareId)
	return c.StreamShareClient.ProviderSharesCdxV1Api.DeleteCdxV1ProviderShareExecute(request)
}

func (c *Client) DescribeProvideShare(shareId string) (cdxv1.CdxV1ProviderShare, *http.Response, error) {
	request := c.StreamShareClient.ProviderSharesCdxV1Api.GetCdxV1ProviderShare(c.cdxApiContext(), shareId)
	return c.StreamShareClient.ProviderSharesCdxV1Api.GetCdxV1ProviderShareExecute(request)
}

func (c *Client) ListProviderShares(sharedResource string) ([]cdxv1.CdxV1ProviderShare, error) {
	providerShares := make([]cdxv1.CdxV1ProviderShare, 0)

	collectedAllShares := false
	pageToken := ""
	for !collectedAllShares {
		sharesList, _, err := c.executeListProviderShares(sharedResource, pageToken)
		if err != nil {
			return nil, err
		}
		providerShares = append(providerShares, sharesList.GetData()...)

		// nextPageUrlStringNullable is nil for the last page
		nextPageUrlStringNullable := sharesList.GetMetadata().Next
		pageToken, collectedAllShares, err = extractCdxNextPagePageToken(nextPageUrlStringNullable)
		if err != nil {
			return nil, err
		}
	}

	return providerShares, nil
}

func (c *Client) executeListProviderShares(sharedResource, pageToken string) (cdxv1.CdxV1ProviderShareList, *http.Response, error) {
	request := c.StreamShareClient.ProviderSharesCdxV1Api.ListCdxV1ProviderShares(c.cdxApiContext()).
		SharedResource(sharedResource).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		request = request.PageToken(pageToken)
	}
	return c.StreamShareClient.ProviderSharesCdxV1Api.ListCdxV1ProviderSharesExecute(request)
}
