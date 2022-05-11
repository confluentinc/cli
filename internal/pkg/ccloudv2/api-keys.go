package ccloudv2

import (
	"context"
	"net/http"

	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newApiKeysClient(baseURL, userAgent string, isTest bool) *apikeysv2.APIClient {
	apiKeysServer := getServerUrl(baseURL, isTest)
	cfg := apikeysv2.NewConfiguration()
	cfg.Servers = apikeysv2.ServerConfigurations{
		{URL: apiKeysServer, Description: "Confluent Cloud IAM"},
	}
	cfg.UserAgent = userAgent
	cfg.Debug = plog.CliLogger.GetLevel() >= plog.DEBUG
	return apikeysv2.NewAPIClient(cfg)
}

func (c *Client) apiKeysApiContext() context.Context {
	return context.WithValue(context.Background(), apikeysv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateApiKey(iamV2ApiKey apikeysv2.IamV2ApiKey) (apikeysv2.IamV2ApiKey, *http.Response, error) {
	req := c.ApiKeysClient.APIKeysIamV2Api.CreateIamV2ApiKey(c.apiKeysApiContext()).IamV2ApiKey(iamV2ApiKey)
	return c.ApiKeysClient.APIKeysIamV2Api.CreateIamV2ApiKeyExecute(req)
}

func (c *Client) DeleteApiKey(id string) (*http.Response, error) {
	req := c.ApiKeysClient.APIKeysIamV2Api.DeleteIamV2ApiKey(c.apiKeysApiContext(), id)
	return c.ApiKeysClient.APIKeysIamV2Api.DeleteIamV2ApiKeyExecute(req)
}

func (c *Client) GetApiKey(id string) (apikeysv2.IamV2ApiKey, *http.Response, error) {
	req := c.ApiKeysClient.APIKeysIamV2Api.GetIamV2ApiKey(c.apiKeysApiContext(), id)
	return c.ApiKeysClient.APIKeysIamV2Api.GetIamV2ApiKeyExecute(req)
}

func (c *Client) UpdateApiKey(id string, iamV2ApiKeyUpdate apikeysv2.IamV2ApiKeyUpdate) (apikeysv2.IamV2ApiKey, *http.Response, error) {
	req := c.ApiKeysClient.APIKeysIamV2Api.UpdateIamV2ApiKey(c.apiKeysApiContext(), id).IamV2ApiKeyUpdate(iamV2ApiKeyUpdate)
	return c.ApiKeysClient.APIKeysIamV2Api.UpdateIamV2ApiKeyExecute(req)
}

func (c *Client) ListApiKeys() ([]apikeysv2.IamV2ApiKey, error) {
	keys := make([]apikeysv2.IamV2ApiKey, 0)

	collectedAllKeys := false
	pageToken := ""
	for !collectedAllKeys {
		keyList, _, err := c.executeListApiKeys(pageToken)
		if err != nil {
			return nil, err
		}
		keys = append(keys, keyList.GetData()...)

		// nextPageUrlStringNullable is nil for the last page
		nextPageUrlStringNullable := keyList.GetMetadata().Next
		pageToken, collectedAllKeys, err = extractApiKeysNextPagePageToken(nextPageUrlStringNullable)
		if err != nil {
			return nil, err
		}
	}
	return keys, nil
}

func (c *Client) executeListApiKeys(pageToken string) (apikeysv2.IamV2ApiKeyList, *http.Response, error) {
	req := c.ApiKeysClient.APIKeysIamV2Api.ListIamV2ApiKeys(c.apiKeysApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.ApiKeysClient.APIKeysIamV2Api.ListIamV2ApiKeysExecute(req)
}
