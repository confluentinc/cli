package ccloudv2

import (
	"context"
	"net/http"

	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"

	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newMdsClient(baseURL, userAgent string, isTest bool) *mdsv2.APIClient {
	mdsServer := getServerUrl(baseURL, isTest)
	cfg := mdsv2.NewConfiguration()
	cfg.Servers = mdsv2.ServerConfigurations{
		{URL: mdsServer, Description: "Confluent Cloud IAM"},
	}
	cfg.UserAgent = userAgent
	cfg.Debug = plog.CliLogger.GetLevel() >= plog.DEBUG
	return mdsv2.NewAPIClient(cfg)
}

func (c *Client) mdsApiContext() context.Context {
	return context.WithValue(context.Background(), mdsv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateIamRoleBinding(iamV2RoleBinding mdsv2.IamV2RoleBinding) (mdsv2.IamV2RoleBinding, *http.Response, error) {
	req := c.MdsClient.RoleBindingsIamV2Api.CreateIamV2RoleBinding(c.mdsApiContext()).IamV2RoleBinding(iamV2RoleBinding)
	return c.MdsClient.RoleBindingsIamV2Api.CreateIamV2RoleBindingExecute(req)
}

func (c *Client) DeleteIamRoleBinding(id string) (mdsv2.IamV2RoleBinding, *http.Response, error) {
	req := c.MdsClient.RoleBindingsIamV2Api.DeleteIamV2RoleBinding(c.mdsApiContext(), id)
	return c.MdsClient.RoleBindingsIamV2Api.DeleteIamV2RoleBindingExecute(req)
}

func (c *Client) ListIamRoleBindings(crnPattern, principal, roleName string) ([]mdsv2.IamV2RoleBinding, error) {
	roleBindings := make([]mdsv2.IamV2RoleBinding, 0)

	collectedAllRoleBindings := false
	pageToken := ""
	for !collectedAllRoleBindings {
		roleBindingList, _, err := c.executeListRoleBindings(pageToken, crnPattern, principal, roleName)
		if err != nil {
			return nil, err
		}
		roleBindings = append(roleBindings, roleBindingList.GetData()...)

		// nextPageUrlStringNullable is nil for the last page
		nextPageUrlStringNullable := roleBindingList.GetMetadata().Next
		pageToken, collectedAllRoleBindings, err = extractMdsNextPagePageToken(nextPageUrlStringNullable)
		if err != nil {
			return nil, err
		}
	}
	return roleBindings, nil
}

func (c *Client) executeListRoleBindings(pageToken, crnPattern, principal, roleName string) (mdsv2.IamV2RoleBindingList, *http.Response, error) {
	var req mdsv2.ApiListIamV2RoleBindingsRequest
	if pageToken != "" {
		req = c.MdsClient.RoleBindingsIamV2Api.ListIamV2RoleBindings(c.mdsApiContext()).CrnPattern(crnPattern).Principal(principal).RoleName(roleName).PageSize(ccloudV2ListPageSize).PageToken(pageToken)
	} else {
		req = c.MdsClient.RoleBindingsIamV2Api.ListIamV2RoleBindings(c.mdsApiContext()).CrnPattern(crnPattern).Principal(principal).RoleName(roleName).PageSize(ccloudV2ListPageSize)
	}
	return c.MdsClient.RoleBindingsIamV2Api.ListIamV2RoleBindingsExecute(req)
}

func (c *Client) ListIamRoleBindingsNaive(crnPattern, principal, roleName string) (mdsv2.IamV2RoleBindingList, *http.Response, error) {
	req := c.MdsClient.RoleBindingsIamV2Api.ListIamV2RoleBindings(c.mdsApiContext()).CrnPattern(crnPattern)
	if principal != "" {
		req = req.Principal(principal)
	}
	if roleName != "" {
		req = req.RoleName(roleName)
	}
	return c.MdsClient.RoleBindingsIamV2Api.ListIamV2RoleBindingsExecute(req)
}
