package ccloudv2

import (
	"context"
	"fmt"
	"net/http"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newIamClient(url, userAgent string, unsafeTrace bool) *iamv2.APIClient {
	cfg := iamv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClient(unsafeTrace)
	cfg.Servers = iamv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return iamv2.NewAPIClient(cfg)
}

func (c *Client) iamApiContext() context.Context {
	return context.WithValue(context.Background(), iamv2.ContextAccessToken, c.AuthToken)
}

// iam service-account api calls

func (c *Client) CreateIamServiceAccount(serviceAccount iamv2.IamV2ServiceAccount) (iamv2.IamV2ServiceAccount, *http.Response, error) {
	return c.IamClient.ServiceAccountsIamV2Api.CreateIamV2ServiceAccount(c.iamApiContext()).IamV2ServiceAccount(serviceAccount).Execute()
}

func (c *Client) DeleteIamServiceAccount(id string) error {
	httpResp, err := c.IamClient.ServiceAccountsIamV2Api.DeleteIamV2ServiceAccount(c.iamApiContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIamServiceAccount(id string) (iamv2.IamV2ServiceAccount, *http.Response, error) {
	return c.IamClient.ServiceAccountsIamV2Api.GetIamV2ServiceAccount(c.iamApiContext(), id).Execute()
}

func (c *Client) UpdateIamServiceAccount(id string, update iamv2.IamV2ServiceAccountUpdate) (iamv2.IamV2ServiceAccount, *http.Response, error) {
	return c.IamClient.ServiceAccountsIamV2Api.UpdateIamV2ServiceAccount(c.iamApiContext(), id).IamV2ServiceAccountUpdate(update).Execute()
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

		pageToken, done, err = extractNextPageToken(page.GetMetadata().Next)
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
	return req.Execute()
}

// iam user api calls

func (c *Client) DeleteIamUser(id string) error {
	httpResp, err := c.IamClient.UsersIamV2Api.DeleteIamV2User(c.iamApiContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateIamUser(id string, update iamv2.IamV2UserUpdate) (iamv2.IamV2User, error) {
	resp, httpResp, err := c.IamClient.UsersIamV2Api.UpdateIamV2User(c.iamApiContext(), id).IamV2UserUpdate(update).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIamUserById(id string) (iamv2.IamV2User, error) {
	resp, httpResp, err := c.IamClient.UsersIamV2Api.GetIamV2User(c.iamApiContext(), id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIamUserByEmail(email string) (iamv2.IamV2User, error) {
	resp, httpResp, err := c.IamClient.UsersIamV2Api.ListIamV2Users(c.iamApiContext()).Execute()
	if err != nil {
		return iamv2.IamV2User{}, errors.CatchCCloudV2Error(err, httpResp)
	}
	for _, user := range resp.Data {
		if email == user.GetEmail() {
			return user, nil
		}
	}
	return iamv2.IamV2User{}, fmt.Errorf(errors.InvalidEmailErrorMsg, email)
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

		pageToken, done, err = extractNextPageToken(page.GetMetadata().Next)
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
	return req.Execute()
}

// iam user invitation api calls

func (c *Client) CreateIamInvitation(invitation iamv2.IamV2Invitation) (iamv2.IamV2Invitation, error) {
	resp, httpResp, err := c.IamClient.InvitationsIamV2Api.CreateIamV2Invitation(c.iamApiContext()).IamV2Invitation(invitation).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteIamInvitation(id string) error {
	httpResp, err := c.IamClient.InvitationsIamV2Api.DeleteIamV2Invitation(c.iamApiContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetIamInvitation(id string) (iamv2.IamV2Invitation, error) {
	resp, httpResp, err := c.IamClient.InvitationsIamV2Api.GetIamV2Invitation(c.iamApiContext(), id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListIamInvitations() ([]iamv2.IamV2Invitation, error) {
	var list []iamv2.IamV2Invitation

	done := false
	pageToken := ""
	for !done {
		page, _, err := c.executeListInvitations(pageToken)
		if err != nil {
			return nil, err
		}
		list = append(list, page.GetData()...)

		pageToken, done, err = extractNextPageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *Client) executeListInvitations(pageToken string) (iamv2.IamV2InvitationList, *http.Response, error) {
	req := c.IamClient.InvitationsIamV2Api.ListIamV2Invitations(c.iamApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
