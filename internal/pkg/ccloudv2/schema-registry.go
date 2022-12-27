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

func (c *Client) ExtractNextPageToken(nextPageUrlStringNullable srcm.NullableString) (string, bool, error) {
	if !nextPageUrlStringNullable.IsSet() {
		return "", true, nil
	}
	nextPageUrlString := *nextPageUrlStringNullable.Get()
	pageToken, err := extractPageToken(nextPageUrlString)
	return pageToken, false, err
}
