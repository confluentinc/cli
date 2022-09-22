package ccloudv2

import (
	"context"
	"net/http"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newIamClient(url, userAgent string, unsafeTrace bool) *iamv2.APIClient {
	cfg := iamv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = iamv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return iamv2.NewAPIClient(cfg)
}

func (c *Client) iamApiContext() context.Context {
	return context.WithValue(context.Background(), iamv2.ContextAccessToken, c.AuthToken)
}

// iam service-account api calls

func (c *Client) CreateIamServiceAccount(serviceAccount iamv2.IamV2ServiceAccount) (iamv2.IamV2ServiceAccount, *http.Response, error) {
	req := c.IamClient.ServiceAccountsIamV2Api.CreateIamV2ServiceAccount(c.iamApiContext()).IamV2ServiceAccount(serviceAccount)
	return c.IamClient.ServiceAccountsIamV2Api.CreateIamV2ServiceAccountExecute(req)
}

func (c *Client) DeleteIamServiceAccount(id string) error {
	req := c.IamClient.ServiceAccountsIamV2Api.DeleteIamV2ServiceAccount(c.iamApiContext(), id)
	httpResp, err := c.IamClient.ServiceAccountsIamV2Api.DeleteIamV2ServiceAccountExecute(req)
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIamServiceAccount(id string) (iamv2.IamV2ServiceAccount, *http.Response, error) {
	req := c.IamClient.ServiceAccountsIamV2Api.GetIamV2ServiceAccount(c.iamApiContext(), id)
	return c.IamClient.ServiceAccountsIamV2Api.GetIamV2ServiceAccountExecute(req)
}

func (c *Client) UpdateIamServiceAccount(id string, update iamv2.IamV2ServiceAccountUpdate) (iamv2.IamV2ServiceAccount, *http.Response, error) {
	req := c.IamClient.ServiceAccountsIamV2Api.UpdateIamV2ServiceAccount(c.iamApiContext(), id).IamV2ServiceAccountUpdate(update)
	return c.IamClient.ServiceAccountsIamV2Api.UpdateIamV2ServiceAccountExecute(req)
}

func (c *Client) ListIamServiceAccounts() ([]iamv2.IamV2ServiceAccount, error) {
	var list []iamv2.IamV2ServiceAccount

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListServiceAccounts(pageToken)
		if err != nil {
			return nil, errors.CatchCCloudV2Error(err, httpResp)
		}
		list = append(list, page.GetData()...)

		pageToken, done, err = extractIamNextPageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *Client) executeListServiceAccounts(pageToken string) (iamv2.IamV2ServiceAccountList, *http.Response, error) {
	req := c.IamClient.ServiceAccountsIamV2Api.ListIamV2ServiceAccounts(c.iamApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.IamClient.ServiceAccountsIamV2Api.ListIamV2ServiceAccountsExecute(req)
}

// iam user api calls

func (c *Client) DeleteIamUser(id string) error {
	req := c.IamClient.UsersIamV2Api.DeleteIamV2User(c.iamApiContext(), id)
	httpResp, err := c.IamClient.UsersIamV2Api.DeleteIamV2UserExecute(req)
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateIamUser(id string, update iamv2.IamV2UserUpdate) (iamv2.IamV2User, error) {
	req := c.IamClient.UsersIamV2Api.UpdateIamV2User(c.iamApiContext(), id).IamV2UserUpdate(update)
	resp, httpResp, err := c.IamClient.UsersIamV2Api.UpdateIamV2UserExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListIamUsers() ([]iamv2.IamV2User, error) {
	var list []iamv2.IamV2User

	done := false
	pageToken := ""
	for !done {
		page, _, err := c.executeListUsers(pageToken)
		if err != nil {
			return nil, err
		}
		list = append(list, page.GetData()...)

		pageToken, done, err = extractIamNextPageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *Client) executeListUsers(pageToken string) (iamv2.IamV2UserList, *http.Response, error) {
	req := c.IamClient.UsersIamV2Api.ListIamV2Users(c.iamApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.IamClient.UsersIamV2Api.ListIamV2UsersExecute(req)
}

func extractIamNextPageToken(nextPageUrlStringNullable iamv2.NullableString) (string, bool, error) {
	if !nextPageUrlStringNullable.IsSet() {
		return "", true, nil
	}
	nextPageUrlString := *nextPageUrlStringNullable.Get()
	pageToken, err := extractPageToken(nextPageUrlString)
	return pageToken, false, err
}
