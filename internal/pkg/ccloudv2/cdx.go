package ccloudv2

import (
	"context"
	"net/http"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"

	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newCdxClient(baseURL, userAgent string, isTest bool) *cdxv1.APIClient {
	cfg := cdxv1.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = cdxv1.ServerConfigurations{{URL: getServerUrl(baseURL, isTest)}}
	cfg.UserAgent = userAgent

	return cdxv1.NewAPIClient(cfg)
}

func (c *Client) cdxApiContext() context.Context {
	return context.WithValue(context.Background(), cdxv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) DeleteProviderShare(shareId string) (*http.Response, error) {
	req := c.CdxClient.ProviderSharesCdxV1Api.DeleteCdxV1ProviderShare(c.cdxApiContext(), shareId)
	return c.CdxClient.ProviderSharesCdxV1Api.DeleteCdxV1ProviderShareExecute(req)
}

func (c *Client) DescribeProvideShare(shareId string) (cdxv1.CdxV1ProviderShare, *http.Response, error) {
	req := c.CdxClient.ProviderSharesCdxV1Api.GetCdxV1ProviderShare(c.cdxApiContext(), shareId)
	return c.CdxClient.ProviderSharesCdxV1Api.GetCdxV1ProviderShareExecute(req)
}

func (c *Client) ListProviderShares(sharedResource string) ([]cdxv1.CdxV1ProviderShare, error) {
	list := make([]cdxv1.CdxV1ProviderShare, 0)

	done := false
	pageToken := ""
	for !done {
		page, _, err := c.executeListProviderShares(sharedResource, pageToken)
		if err != nil {
			return nil, err
		}
		list = append(list, page.GetData()...)

		// nextPageUrlStringNullable is nil for the last page
		nextPageUrlStringNullable := page.GetMetadata().Next
		pageToken, done, err = extractCdxNextPageToken(nextPageUrlStringNullable)
		if err != nil {
			return nil, err
		}
	}

	return list, nil
}

func (c *Client) executeListProviderShares(sharedResource, pageToken string) (cdxv1.CdxV1ProviderShareList, *http.Response, error) {
	req := c.CdxClient.ProviderSharesCdxV1Api.ListCdxV1ProviderShares(c.cdxApiContext()).SharedResource(sharedResource).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.CdxClient.ProviderSharesCdxV1Api.ListCdxV1ProviderSharesExecute(req)
}

func extractCdxNextPageToken(nextPageUrlStringNullable cdxv1.NullableString) (string, bool, error) {
	if !nextPageUrlStringNullable.IsSet() {
		return "", true, nil
	}
	nextPageUrlString := *nextPageUrlStringNullable.Get()
	pageToken, err := extractPageToken(nextPageUrlString)
	return pageToken, false, err
}
