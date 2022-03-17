package ccloudv2

import (
	"context"
	"fmt"
	"net/http"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func NewV2IamClient(baseURL string, isTest bool) *iamv2.APIClient {
	iamServer := getServerUrl(baseURL, isTest)
	cfg := iamv2.NewConfiguration()
	cfg.Servers = iamv2.ServerConfigurations{
		{URL: iamServer, Description: "Confluent Cloud IAM"},
	}
	return iamv2.NewAPIClient(cfg)
}

func (c *Client) iamApiContext() context.Context {
	return context.WithValue(context.Background(), iamv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateIamServiceAccount(serviceAccount iamv2.IamV2ServiceAccount) (iamv2.IamV2ServiceAccount, *http.Response, error) {
	req := c.IamClient.ServiceAccountsIamV2Api.CreateIamV2ServiceAccount(c.iamApiContext()).IamV2ServiceAccount(serviceAccount)
	return c.IamClient.ServiceAccountsIamV2Api.CreateIamV2ServiceAccountExecute(req)
}

func (c *Client) DeleteIamServiceAccount(id string) (*http.Response, error) {
	req := c.IamClient.ServiceAccountsIamV2Api.DeleteIamV2ServiceAccount(c.iamApiContext(), id)
	return c.IamClient.ServiceAccountsIamV2Api.DeleteIamV2ServiceAccountExecute(req)
}

func (c *Client) GetIamServiceAccount(id string) (iamv2.IamV2ServiceAccount, *http.Response, error) {
	req := c.IamClient.ServiceAccountsIamV2Api.GetIamV2ServiceAccount(c.iamApiContext(), id)
	return c.IamClient.ServiceAccountsIamV2Api.GetIamV2ServiceAccountExecute(req)
}

func (c *Client) ListIamServiceAccounts() (iamv2.IamV2ServiceAccountList, *http.Response, error) {
	req := c.IamClient.ServiceAccountsIamV2Api.ListIamV2ServiceAccounts(c.iamApiContext()).PageSize(100)
	return c.IamClient.ServiceAccountsIamV2Api.ListIamV2ServiceAccountsExecute(req)
}

func (c *Client) UpdateIamServiceAccount(id string, update iamv2.IamV2ServiceAccountUpdate) (iamv2.IamV2ServiceAccount, *http.Response, error) {
	req := c.IamClient.ServiceAccountsIamV2Api.UpdateIamV2ServiceAccount(c.iamApiContext(), id).IamV2ServiceAccountUpdate(update)
	return c.IamClient.ServiceAccountsIamV2Api.UpdateIamV2ServiceAccountExecute(req)
}

func (c *Client) DeleteIamUser(id string) (*http.Response, error) {
	req := c.IamClient.UsersIamV2Api.DeleteIamV2User(c.iamApiContext(), id)
	return c.IamClient.UsersIamV2Api.DeleteIamV2UserExecute(req)
}

func (c *Client) GetIamUserById(id string) (iamv2.IamV2User, *http.Response, error) {
	req := c.IamClient.UsersIamV2Api.GetIamV2User(c.iamApiContext(), id)
	return c.IamClient.UsersIamV2Api.GetIamV2UserExecute(req)
}

func (c *Client) GetIamUserByEmail(email string) (iamv2.IamV2User, error) {
	resp, _, err := c.IamClient.UsersIamV2Api.ListIamV2Users(c.iamApiContext()).Execute()
	if err != nil {
		return iamv2.IamV2User{}, err
	}
	for _, user := range resp.Data {
		if email == *user.Email {
			return user, nil
		}
	}
	return iamv2.IamV2User{}, fmt.Errorf(errors.InvalidEmailErrorMsg, email)
}

func (c *Client) ListIamUsers() (iamv2.IamV2UserList, *http.Response, error) {
	req := c.IamClient.UsersIamV2Api.ListIamV2Users(c.iamApiContext()).PageSize(100)
	return c.IamClient.UsersIamV2Api.ListIamV2UsersExecute(req)
}
