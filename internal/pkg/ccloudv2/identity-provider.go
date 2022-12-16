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
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
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
	req := c.IdentityProviderClient.IdentityProvidersIamV2Api.CreateIamV2IdentityProvider(c.identityProviderApiContext()).IamV2IdentityProvider(identityProvider)
	resp, httpResp, err := c.IdentityProviderClient.IdentityProvidersIamV2Api.CreateIamV2IdentityProviderExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteIdentityProvider(id string) error {
	req := c.IdentityProviderClient.IdentityProvidersIamV2Api.DeleteIamV2IdentityProvider(c.identityProviderApiContext(), id)
	httpResp, err := c.IdentityProviderClient.IdentityProvidersIamV2Api.DeleteIamV2IdentityProviderExecute(req)
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIdentityProvider(id string) (identityproviderv2.IamV2IdentityProvider, error) {
	req := c.IdentityProviderClient.IdentityProvidersIamV2Api.GetIamV2IdentityProvider(c.identityProviderApiContext(), id)
	resp, httpResp, err := c.IdentityProviderClient.IdentityProvidersIamV2Api.GetIamV2IdentityProviderExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateIdentityProvider(update identityproviderv2.IamV2IdentityProviderUpdate) (identityproviderv2.IamV2IdentityProvider, error) {
	req := c.IdentityProviderClient.IdentityProvidersIamV2Api.UpdateIamV2IdentityProvider(c.identityProviderApiContext(), *update.Id).IamV2IdentityProviderUpdate(update)
	resp, httpResp, err := c.IdentityProviderClient.IdentityProvidersIamV2Api.UpdateIamV2IdentityProviderExecute(req)
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
	return c.IdentityProviderClient.IdentityProvidersIamV2Api.ListIamV2IdentityProvidersExecute(req)
}

func (c *Client) CreateIdentityPool(identityPool identityproviderv2.IamV2IdentityPool, providerId string) (identityproviderv2.IamV2IdentityPool, error) {
	req := c.IdentityProviderClient.IdentityPoolsIamV2Api.CreateIamV2IdentityPool(c.identityPoolApiContext(), providerId).IamV2IdentityPool(identityPool)
	resp, httpResp, err := c.IdentityProviderClient.IdentityPoolsIamV2Api.CreateIamV2IdentityPoolExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteIdentityPool(id, providerId string) error {
	req := c.IdentityProviderClient.IdentityPoolsIamV2Api.DeleteIamV2IdentityPool(c.identityPoolApiContext(), providerId, id)
	httpResp, err := c.IdentityProviderClient.IdentityPoolsIamV2Api.DeleteIamV2IdentityPoolExecute(req)
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIdentityPool(id, providerId string) (identityproviderv2.IamV2IdentityPool, error) {
	req := c.IdentityProviderClient.IdentityPoolsIamV2Api.GetIamV2IdentityPool(c.identityPoolApiContext(), providerId, id)
	resp, httpResp, err := c.IdentityProviderClient.IdentityPoolsIamV2Api.GetIamV2IdentityPoolExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateIdentityPool(identityPool identityproviderv2.IamV2IdentityPool, providerId string) (identityproviderv2.IamV2IdentityPool, error) {
	req := c.IdentityProviderClient.IdentityPoolsIamV2Api.UpdateIamV2IdentityPool(c.identityPoolApiContext(), providerId, *identityPool.Id).IamV2IdentityPool(identityPool)
	resp, httpResp, err := c.IdentityProviderClient.IdentityPoolsIamV2Api.UpdateIamV2IdentityPoolExecute(req)
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

func (c *Client) executeListIdentityPools(providerID string, pageToken string) (identityproviderv2.IamV2IdentityPoolList, *http.Response, error) {
	req := c.IdentityProviderClient.IdentityPoolsIamV2Api.ListIamV2IdentityPools(c.identityPoolApiContext(), providerID).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.IdentityProviderClient.IdentityPoolsIamV2Api.ListIamV2IdentityPoolsExecute(req)
}
