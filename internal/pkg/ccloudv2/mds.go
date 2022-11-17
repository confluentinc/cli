package ccloudv2

import (
	"context"

	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newMdsClient(url, userAgent string, unsafeTrace bool) *mdsv2.APIClient {
	cfg := mdsv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = mdsv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return mdsv2.NewAPIClient(cfg)
}

func (c *Client) mdsApiContext() context.Context {
	return context.WithValue(context.Background(), mdsv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateIamRoleBinding(iamV2RoleBinding *mdsv2.IamV2RoleBinding) (mdsv2.IamV2RoleBinding, error) {
	req := c.MdsClient.RoleBindingsIamV2Api.CreateIamV2RoleBinding(c.mdsApiContext()).IamV2RoleBinding(*iamV2RoleBinding)
	resp, httpResp, err := c.MdsClient.RoleBindingsIamV2Api.CreateIamV2RoleBindingExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteIamRoleBinding(id string) (mdsv2.IamV2RoleBinding, error) {
	req := c.MdsClient.RoleBindingsIamV2Api.DeleteIamV2RoleBinding(c.mdsApiContext(), id)
	resp, httpResp, err := c.MdsClient.RoleBindingsIamV2Api.DeleteIamV2RoleBindingExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListIamRoleBindings(principal, roleName, crnPattern string) (mdsv2.IamV2RoleBindingList, error) { // add pagination
	req := c.MdsClient.RoleBindingsIamV2Api.ListIamV2RoleBindings(c.mdsApiContext()).CrnPattern(crnPattern)
	if principal != "" {
		req = req.Principal(principal)
	}
	if roleName != "" {
		req = req.RoleName(roleName)
	}
	resp, httpResp, err := c.MdsClient.RoleBindingsIamV2Api.ListIamV2RoleBindingsExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
