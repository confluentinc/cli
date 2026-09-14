package ccloudv2

import (
	"context"
	"net/http"

	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newIdentityProviderClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *identityproviderv2.APIClient {
	cfg := identityproviderv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = identityproviderv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return identityproviderv2.NewAPIClient(cfg)
}

func (c *Client) identityProviderApiContext() context.Context {
	return context.WithValue(context.Background(), identityproviderv2.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) identityPoolApiContext() context.Context {
	return context.WithValue(context.Background(), identityproviderv2.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) CreateIamIdentityPool(identityPool identityproviderv2.IamV2IdentityPool, providerId string, assignedResourceOwner string) (identityproviderv2.IamV2IdentityPool, error) {
	resp, httpResp, err := c.IdentityProviderClient.IdentityPoolsIamV2Api.
		CreateIamV2IdentityPool(c.identityPoolApiContext(), providerId).
		AssignedResourceOwner(assignedResourceOwner).
		IamV2IdentityPool(identityPool).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteIamIdentityPool(id, providerId string) error {
	httpResp, err := c.IdentityProviderClient.IdentityPoolsIamV2Api.DeleteIamV2IdentityPool(c.identityPoolApiContext(), providerId, id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIamIdentityPool(id, providerId string) (identityproviderv2.IamV2IdentityPool, error) {
	resp, httpResp, err := c.IdentityProviderClient.IdentityPoolsIamV2Api.GetIamV2IdentityPool(c.identityPoolApiContext(), providerId, id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateIamIdentityPool(identityPool identityproviderv2.IamV2IdentityPool, providerId string) (identityproviderv2.IamV2IdentityPool, error) {
	resp, httpResp, err := c.IdentityProviderClient.IdentityPoolsIamV2Api.UpdateIamV2IdentityPool(c.identityPoolApiContext(), providerId, *identityPool.Id).IamV2IdentityPool(identityPool).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListIamIdentityPools(providerId string) ([]identityproviderv2.IamV2IdentityPool, error) {
	var list []identityproviderv2.IamV2IdentityPool

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListIamIdentityPools(providerId, pageToken)
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

func (c *Client) executeListIamIdentityPools(providerId, pageToken string) (identityproviderv2.IamV2IdentityPoolList, *http.Response, error) {
	req := c.IdentityProviderClient.IdentityPoolsIamV2Api.ListIamV2IdentityPools(c.identityPoolApiContext(), providerId).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

// ===== identity providers API calls =====

func (c *Client) CreateIamIdentityProvider(req identityproviderv2.IamV2IdentityProvider) (identityproviderv2.IamV2IdentityProvider, error) {
	createReq := c.IdentityProviderClient.IdentityProvidersIamV2Api.
		CreateIamV2IdentityProvider(c.identityProviderApiContext()).
		IamV2IdentityProvider(req)
	res, httpResp, err := createReq.Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIamIdentityProvider(id string) (identityproviderv2.IamV2IdentityProvider, error) {
	getReq := c.IdentityProviderClient.IdentityProvidersIamV2Api.
		GetIamV2IdentityProvider(c.identityProviderApiContext(), id)
	res, httpResp, err := getReq.Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateIamIdentityProvider(id string, update identityproviderv2.IamV2IdentityProvider) (identityproviderv2.IamV2IdentityProvider, error) {
	updateReq := c.IdentityProviderClient.IdentityProvidersIamV2Api.
		UpdateIamV2IdentityProvider(c.identityProviderApiContext(), id).
		IamV2IdentityProvider(update)
	res, httpResp, err := updateReq.Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteIamIdentityProvider(id string) error {
	deleteReq := c.IdentityProviderClient.IdentityProvidersIamV2Api.
		DeleteIamV2IdentityProvider(c.identityProviderApiContext(), id)
	httpResp, err := deleteReq.Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListIamIdentityProviders() ([]identityproviderv2.IamV2IdentityProvider, error) {
	var list []identityproviderv2.IamV2IdentityProvider

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListIamIdentityProviders(pageToken)
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

func (c *Client) executeListIamIdentityProviders(pageToken string) (identityproviderv2.IamV2IdentityProviderList, *http.Response, error) {
	req := c.IdentityProviderClient.IdentityProvidersIamV2Api.
		ListIamV2IdentityProviders(c.identityProviderApiContext()).
		PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
