package ccloudv2

import (
	"context"
	"net/http"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"

	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newIdentityProviderClient(baseURL, userAgent string, isTest bool) *identityproviderv2.APIClient {
	cfg := identityproviderv2.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = identityproviderv2.ServerConfigurations{{URL: getServerUrl(baseURL, isTest)}}
	cfg.UserAgent = userAgent

	return identityproviderv2.NewAPIClient(cfg)
}

func (c *Client) identityProviderApiContext() context.Context {
	return context.WithValue(context.Background(), identityproviderv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) identityPoolApiContext() context.Context {
	return context.WithValue(context.Background(), identityproviderv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateIdentityProvider(identityProvider identityproviderv2.IamV2IdentityProvider) (identityproviderv2.IamV2IdentityProvider, *http.Response, error) {
	req := c.IdentityProviderClient.IdentityProvidersIamV2Api.CreateIamV2IdentityProvider(c.identityProviderApiContext()).IamV2IdentityProvider(identityProvider)
	return c.IdentityProviderClient.IdentityProvidersIamV2Api.CreateIamV2IdentityProviderExecute(req)
}

func (c *Client) DeleteIdentityProvider(id string) (*http.Response, error) {
	req := c.IdentityProviderClient.IdentityProvidersIamV2Api.DeleteIamV2IdentityProvider(c.identityProviderApiContext(), id)
	return c.IdentityProviderClient.IdentityProvidersIamV2Api.DeleteIamV2IdentityProviderExecute(req)
}

func (c *Client) GetIdentityProvider(id string) (identityproviderv2.IamV2IdentityProvider, *http.Response, error) {
	req := c.IdentityProviderClient.IdentityProvidersIamV2Api.GetIamV2IdentityProvider(c.identityProviderApiContext(), id)
	return c.IdentityProviderClient.IdentityProvidersIamV2Api.GetIamV2IdentityProviderExecute(req)
}

func (c *Client) UpdateIdentityProvider(update identityproviderv2.IamV2IdentityProviderUpdate) (identityproviderv2.IamV2IdentityProvider, *http.Response, error) {
	req := c.IdentityProviderClient.IdentityProvidersIamV2Api.UpdateIamV2IdentityProvider(c.identityProviderApiContext(), *update.Id).IamV2IdentityProviderUpdate(update)
	return c.IdentityProviderClient.IdentityProvidersIamV2Api.UpdateIamV2IdentityProviderExecute(req)
}

func (c *Client) ListIdentityProviders() ([]identityproviderv2.IamV2IdentityProvider, error) {
	var identityProviders []identityproviderv2.IamV2IdentityProvider

	collectedAllIdentityProviders := false
	pageToken := ""
	for !collectedAllIdentityProviders {
		identityProviderList, _, err := c.executeListIdentityProviders(pageToken)
		if err != nil {
			return nil, err
		}
		identityProviders = append(identityProviders, identityProviderList.GetData()...)

		// nextPageUrlStringNullable is nil for the last page
		nextPageUrlStringNullable := identityProviderList.GetMetadata().Next
		pageToken, collectedAllIdentityProviders, err = extractIdentityProviderNextPagePageToken(nextPageUrlStringNullable)
		if err != nil {
			return nil, err
		}
	}
	return identityProviders, nil
}

func (c *Client) executeListIdentityProviders(pageToken string) (identityproviderv2.IamV2IdentityProviderList, *http.Response, error) {
	req := c.IdentityProviderClient.IdentityProvidersIamV2Api.ListIamV2IdentityProviders(c.identityProviderApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.IdentityProviderClient.IdentityProvidersIamV2Api.ListIamV2IdentityProvidersExecute(req)
}

func (c *Client) CreateIdentityPool(identityPool identityproviderv2.IamV2IdentityPool, providerId string) (identityproviderv2.IamV2IdentityPool, *http.Response, error) {
	req := c.IdentityProviderClient.IdentityPoolsIamV2Api.CreateIamV2IdentityPool(c.identityPoolApiContext(), providerId).IamV2IdentityPool(identityPool)
	return c.IdentityProviderClient.IdentityPoolsIamV2Api.CreateIamV2IdentityPoolExecute(req)
}

func (c *Client) DeleteIdentityPool(id, providerId string) (*http.Response, error) {
	req := c.IdentityProviderClient.IdentityPoolsIamV2Api.DeleteIamV2IdentityPool(c.identityPoolApiContext(), providerId, id)
	return c.IdentityProviderClient.IdentityPoolsIamV2Api.DeleteIamV2IdentityPoolExecute(req)
}

func (c *Client) GetIdentityPool(id, providerId string) (identityproviderv2.IamV2IdentityPool, *http.Response, error) {
	req := c.IdentityProviderClient.IdentityPoolsIamV2Api.GetIamV2IdentityPool(c.identityPoolApiContext(), providerId, id)
	return c.IdentityProviderClient.IdentityPoolsIamV2Api.GetIamV2IdentityPoolExecute(req)
}

func (c *Client) UpdateIdentityPool(identityPool identityproviderv2.IamV2IdentityPool, providerId string) (identityproviderv2.IamV2IdentityPool, *http.Response, error) {
	req := c.IdentityProviderClient.IdentityPoolsIamV2Api.UpdateIamV2IdentityPool(c.identityPoolApiContext(), providerId, *identityPool.Id).IamV2IdentityPool(identityPool)
	return c.IdentityProviderClient.IdentityPoolsIamV2Api.UpdateIamV2IdentityPoolExecute(req)
}

func (c *Client) ListIdentityPools(providerId string) ([]identityproviderv2.IamV2IdentityPool, error) {
	var identityPools []identityproviderv2.IamV2IdentityPool

	collectedAllIdentityPools := false
	pageToken := ""
	for !collectedAllIdentityPools {
		identityPoolList, _, err := c.executeListIdentityPools(providerId, pageToken)
		if err != nil {
			return nil, err
		}
		identityPools = append(identityPools, identityPoolList.GetData()...)

		// nextPageUrlStringNullable is nil for the last page
		nextPageUrlStringNullable := identityPoolList.GetMetadata().Next
		pageToken, collectedAllIdentityPools, err = extractIdentityPoolNextPagePageToken(nextPageUrlStringNullable)
		if err != nil {
			return nil, err
		}
	}
	return identityPools, nil
}

func (c *Client) executeListIdentityPools(providerID string, pageToken string) (identityproviderv2.IamV2IdentityPoolList, *http.Response, error) {
	req := c.IdentityProviderClient.IdentityPoolsIamV2Api.ListIamV2IdentityPools(c.identityPoolApiContext(), providerID).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.IdentityProviderClient.IdentityPoolsIamV2Api.ListIamV2IdentityPoolsExecute(req)
}
