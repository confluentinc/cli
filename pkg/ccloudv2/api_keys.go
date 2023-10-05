package ccloudv2

import (
	"context"
	"fmt"
	"net/http"

	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newApiKeysClient(url, userAgent string, unsafeTrace bool) *apikeysv2.APIClient {
	cfg := apikeysv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClient(unsafeTrace)
	cfg.Servers = apikeysv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return apikeysv2.NewAPIClient(cfg)
}

func (c *Client) apiKeysApiContext() context.Context {
	return context.WithValue(context.Background(), apikeysv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateApiKey(iamV2ApiKey apikeysv2.IamV2ApiKey) (apikeysv2.IamV2ApiKey, *http.Response, error) {
	return c.ApiKeysClient.APIKeysIamV2Api.CreateIamV2ApiKey(c.apiKeysApiContext()).IamV2ApiKey(iamV2ApiKey).Execute()
}

func (c *Client) DeleteApiKey(id string) (*http.Response, error) {
	return c.ApiKeysClient.APIKeysIamV2Api.DeleteIamV2ApiKey(c.apiKeysApiContext(), id).Execute()
}

func (c *Client) GetApiKey(id string) (apikeysv2.IamV2ApiKey, *http.Response, error) {
	apiKey, httpResp, err := c.ApiKeysClient.APIKeysIamV2Api.GetIamV2ApiKey(c.apiKeysApiContext(), id).Execute()
	if err == nil {
		return apiKey, httpResp, nil
	}

	apiKeys, err := c.ListApiKeys("", "")
	if err != nil {
		return apiKey, httpResp, err
	}

	for _, key := range apiKeys {
		if *key.Id == id {
			return apiKey, httpResp, err
		}
	}

	return apiKey, httpResp, errors.NewErrorWithSuggestions(fmt.Sprintf("unknown API key %s", id), errors.ApiKeyNotFoundSuggestions)
}

func (c *Client) UpdateApiKey(id string, iamV2ApiKeyUpdate apikeysv2.IamV2ApiKeyUpdate) (apikeysv2.IamV2ApiKey, *http.Response, error) {
	return c.ApiKeysClient.APIKeysIamV2Api.UpdateIamV2ApiKey(c.apiKeysApiContext(), id).IamV2ApiKeyUpdate(iamV2ApiKeyUpdate).Execute()
}

func (c *Client) ListApiKeys(owner, resource string) ([]apikeysv2.IamV2ApiKey, error) {
	var list []apikeysv2.IamV2ApiKey

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListApiKeys(owner, resource, pageToken)
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

func (c *Client) executeListApiKeys(owner, resource, pageToken string) (apikeysv2.IamV2ApiKeyList, *http.Response, error) {
	req := c.ApiKeysClient.APIKeysIamV2Api.ListIamV2ApiKeys(c.apiKeysApiContext()).PageSize(ccloudV2ListPageSize).SpecOwner(owner).SpecResource(resource)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
