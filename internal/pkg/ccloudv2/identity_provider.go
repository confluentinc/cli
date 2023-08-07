package ccloudv2

import (
	"context"
	"net/http"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newIdentityProviderClient(url, userAgent string, unsafeTrace bool) *identityproviderv2.APIClient {
	cfg := identityproviderv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClient(unsafeTrace)
	cfg.Servers = identityproviderv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return identityproviderv2.NewAPIClient(cfg)
}

func (c *Client) identityProviderApiContext() context.Context {
	return context.WithValue(context.Background(), identityproviderv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) identityPoolApiContext() context.Context {
	return context.WithValue(context.Background(), identityproviderv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateIdentityProvider(identityProvider identityproviderv2.IamV2IdentityProvider) (identityproviderv2.IamV2IdentityProvider, error) {
	resp, httpResp, err := c.IdentityProviderClient.IdentityProvidersIamV2Api.CreateIamV2IdentityProvider(c.identityProviderApiContext()).IamV2IdentityProvider(identityProvider).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteIdentityProvider(id string) error {
	httpResp, err := c.IdentityProviderClient.IdentityProvidersIamV2Api.DeleteIamV2IdentityProvider(c.identityProviderApiContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIdentityProvider(id string) (identityproviderv2.IamV2IdentityProvider, error) {
	resp, httpResp, err := c.IdentityProviderClient.IdentityProvidersIamV2Api.GetIamV2IdentityProvider(c.identityProviderApiContext(), id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateIdentityProvider(update identityproviderv2.IamV2IdentityProviderUpdate) (identityproviderv2.IamV2IdentityProvider, error) {
	resp, httpResp, err := c.IdentityProviderClient.IdentityProvidersIamV2Api.UpdateIamV2IdentityProvider(c.identityProviderApiContext(), *update.Id).IamV2IdentityProviderUpdate(update).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListIdentityProviders() ([]identityproviderv2.IamV2IdentityProvider, error) {
	var list []identityproviderv2.IamV2IdentityProvider

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListIdentityProviders(pageToken)
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

func (c *Client) executeListIdentityProviders(pageToken string) (identityproviderv2.IamV2IdentityProviderList, *http.Response, error) {
	req := c.IdentityProviderClient.IdentityProvidersIamV2Api.ListIamV2IdentityProviders(c.identityProviderApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) CreateIdentityPool(identityPool identityproviderv2.IamV2IdentityPool, providerId string) (identityproviderv2.IamV2IdentityPool, error) {
	resp, httpResp, err := c.IdentityProviderClient.IdentityPoolsIamV2Api.CreateIamV2IdentityPool(c.identityPoolApiContext(), providerId).IamV2IdentityPool(identityPool).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteIdentityPool(id, providerId string) error {
	httpResp, err := c.IdentityProviderClient.IdentityPoolsIamV2Api.DeleteIamV2IdentityPool(c.identityPoolApiContext(), providerId, id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIdentityPool(id, providerId string) (identityproviderv2.IamV2IdentityPool, error) {
	resp, httpResp, err := c.IdentityProviderClient.IdentityPoolsIamV2Api.GetIamV2IdentityPool(c.identityPoolApiContext(), providerId, id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateIdentityPool(identityPool identityproviderv2.IamV2IdentityPool, providerId string) (identityproviderv2.IamV2IdentityPool, error) {
	resp, httpResp, err := c.IdentityProviderClient.IdentityPoolsIamV2Api.UpdateIamV2IdentityPool(c.identityPoolApiContext(), providerId, *identityPool.Id).IamV2IdentityPool(identityPool).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListIdentityPools(providerId string) ([]identityproviderv2.IamV2IdentityPool, error) {
	var list []identityproviderv2.IamV2IdentityPool

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListIdentityPools(providerId, pageToken)
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

func (c *Client) executeListIdentityPools(providerId, pageToken string) (identityproviderv2.IamV2IdentityPoolList, *http.Response, error) {
	req := c.IdentityProviderClient.IdentityPoolsIamV2Api.ListIamV2IdentityPools(c.identityPoolApiContext(), providerId).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
