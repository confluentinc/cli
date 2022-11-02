package ccloudv2

import (
	"context"
	"net/http"

	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"

	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newMdsClient(baseURL, userAgent string, isTest bool) *mdsv2.APIClient {
	cfg := mdsv2.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = mdsv2.ServerConfigurations{{URL: getServerUrl(baseURL, isTest), Description: "Confluent Cloud MDS"}}
	cfg.UserAgent = userAgent

	return mdsv2.NewAPIClient(cfg)
}

func (c *Client) mdsApiContext() context.Context {
	return context.WithValue(context.Background(), mdsv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateIamRoleBinding(iamV2RoleBinding *mdsv2.IamV2RoleBinding) (mdsv2.IamV2RoleBinding, *http.Response, error) {
	req := c.MdsClient.RoleBindingsIamV2Api.CreateIamV2RoleBinding(c.mdsApiContext()).IamV2RoleBinding(*iamV2RoleBinding)
	return c.MdsClient.RoleBindingsIamV2Api.CreateIamV2RoleBindingExecute(req)
}

func (c *Client) DeleteIamRoleBinding(id string) (mdsv2.IamV2RoleBinding, *http.Response, error) {
	req := c.MdsClient.RoleBindingsIamV2Api.DeleteIamV2RoleBinding(c.mdsApiContext(), id)
	return c.MdsClient.RoleBindingsIamV2Api.DeleteIamV2RoleBindingExecute(req)
}

func (c *Client) ListIamRoleBindings(iamV2RoleBinding *mdsv2.IamV2RoleBinding) (mdsv2.IamV2RoleBindingList, *http.Response, error) {
	principal := *iamV2RoleBinding.Principal
	roleName := *iamV2RoleBinding.RoleName
	req := c.MdsClient.RoleBindingsIamV2Api.ListIamV2RoleBindings(c.mdsApiContext()).CrnPattern(*iamV2RoleBinding.CrnPattern)
	if principal != "" {
		req = req.Principal(principal)
	}
	if roleName != "" {
		req = req.RoleName(roleName)
	}
	return c.MdsClient.RoleBindingsIamV2Api.ListIamV2RoleBindingsExecute(req)
}
