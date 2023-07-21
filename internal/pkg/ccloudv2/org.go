package ccloudv2

import (
	"context"
	"fmt"
	"net/http"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newOrgClient(url, userAgent string, unsafeTrace bool) *orgv2.APIClient {
	cfg := orgv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = orgv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return orgv2.NewAPIClient(cfg)
}

func (c *Client) orgApiContext() context.Context {
	return context.WithValue(context.Background(), orgv2.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateOrgEnvironment(environment orgv2.OrgV2Environment) (orgv2.OrgV2Environment, error) {
	res, httpResp, err := c.OrgClient.EnvironmentsOrgV2Api.CreateOrgV2Environment(c.orgApiContext()).OrgV2Environment(environment).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetOrgEnvironment(envId string) (orgv2.OrgV2Environment, error) {
	env, httpResp, err := c.OrgClient.EnvironmentsOrgV2Api.GetOrgV2Environment(c.orgApiContext(), envId).Execute()
	if err != nil {
		if envId, err = EnvironmentNameToId(envId, c); err != nil {
			if err.Error() == fmt.Sprintf(ResourceNameNotFoundErrorMsg, envId) {
				err = errors.NewErrorWithSuggestions(err.Error(), errors.NotValidEnvironmentIdSuggestions)
			}
			return env, err
		}
		env, httpResp, err = c.OrgClient.EnvironmentsOrgV2Api.GetOrgV2Environment(c.orgApiContext(), envId).Execute()
	}
	return env, errors.CatchCCloudV2ResourceNotFoundError(err, envId, httpResp)
}

func (c *Client) UpdateOrgEnvironment(envId string, updateEnvironment orgv2.OrgV2Environment) (orgv2.OrgV2Environment, error) {
	env, httpResp, err := c.OrgClient.EnvironmentsOrgV2Api.UpdateOrgV2Environment(c.orgApiContext(), envId).OrgV2Environment(updateEnvironment).Execute()
	if err != nil {
		if envId, err = EnvironmentNameToId(envId, c); err != nil {
			if err.Error() == fmt.Sprintf(ResourceNameNotFoundErrorMsg, envId) {
				err = errors.NewErrorWithSuggestions(err.Error(), errors.NotValidEnvironmentIdSuggestions)
			}
			return env, err
		}
		env, httpResp, err = c.OrgClient.EnvironmentsOrgV2Api.UpdateOrgV2Environment(c.orgApiContext(), envId).OrgV2Environment(updateEnvironment).Execute()
	}
	return env, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteOrgEnvironment(envId string) error {
	httpResp, err := c.OrgClient.EnvironmentsOrgV2Api.DeleteOrgV2Environment(c.orgApiContext(), envId).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListOrgEnvironments() ([]orgv2.OrgV2Environment, error) {
	var list []orgv2.OrgV2Environment

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListEnvironments(pageToken)
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

func (c *Client) executeListEnvironments(pageToken string) (orgv2.OrgV2EnvironmentList, *http.Response, error) {
	req := c.OrgClient.EnvironmentsOrgV2Api.ListOrgV2Environments(c.orgApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) GetOrgOrganization(orgId string) (orgv2.OrgV2Organization, *http.Response, error) {
	return c.OrgClient.OrganizationsOrgV2Api.GetOrgV2Organization(c.orgApiContext(), orgId).Execute()
}

func (c *Client) UpdateOrgOrganization(orgId string, updateOrganization orgv2.OrgV2Organization) (orgv2.OrgV2Organization, *http.Response, error) {
	return c.OrgClient.OrganizationsOrgV2Api.UpdateOrgV2Organization(c.orgApiContext(), orgId).OrgV2Organization(updateOrganization).Execute()
}

func (c *Client) ListOrgOrganizations() ([]orgv2.OrgV2Organization, error) {
	var list []orgv2.OrgV2Organization

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListOrganizations(pageToken)
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

func (c *Client) executeListOrganizations(pageToken string) (orgv2.OrgV2OrganizationList, *http.Response, error) {
	req := c.OrgClient.OrganizationsOrgV2Api.ListOrgV2Organizations(c.orgApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
