package ccloudv2

import (
	"context"
	"net/http"

	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2/byok/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newByokV1Client(url, userAgent string, unsafeTrace bool) *byokv1.APIClient {
	cfg := byokv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = byokv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return byokv1.NewAPIClient(cfg)
}

func (c *Client) byokApiContext() context.Context {
	return context.WithValue(context.Background(), byokv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateByokKey(key byokv1.ByokV1Key) (byokv1.ByokV1Key, *http.Response, error) {
	return c.ByokClient.KeysByokV1Api.CreateByokV1Key(c.byokApiContext()).ByokV1Key(key).Execute()
}

func (c *Client) GetByokKey(keyId string) (byokv1.ByokV1Key, *http.Response, error) {
	return c.ByokClient.KeysByokV1Api.GetByokV1Key(c.byokApiContext(), keyId).Execute()
}

func (c *Client) DeleteByokKey(keyId string) (*http.Response, error) {
	return c.ByokClient.KeysByokV1Api.DeleteByokV1Key(c.byokApiContext(), keyId).Execute()
}

func (c *Client) ListByokKeys(provider string, state string) ([]byokv1.ByokV1Key, error) {
	var list []byokv1.ByokV1Key

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListByokKeys(pageToken, provider, state)
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

func (c *Client) executeListByokKeys(pageToken string, provider string, state string) (byokv1.ByokV1KeyList, *http.Response, error) {
	req := c.ByokClient.KeysByokV1Api.ListByokV1Keys(c.byokApiContext()).PageSize(ccloudV2ListPageSize)
	if provider != "" {
		req = req.Provider(provider)
	}
	if state != "" {
		req = req.State(state)
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
