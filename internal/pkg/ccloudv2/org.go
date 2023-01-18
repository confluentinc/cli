package ccloudv2

import (
	"context"
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

func (c *Client) CreateOrgEnvironment(environment orgv2.OrgV2Environment) (orgv2.OrgV2Environment, *http.Response, error) {
	req := c.OrgClient.EnvironmentsOrgV2Api.CreateOrgV2Environment(c.orgApiContext()).OrgV2Environment(environment)
	return c.OrgClient.EnvironmentsOrgV2Api.CreateOrgV2EnvironmentExecute(req)
}

func (c *Client) GetOrgEnvironment(envId string) (orgv2.OrgV2Environment, *http.Response, error) {
	req := c.OrgClient.EnvironmentsOrgV2Api.GetOrgV2Environment(c.orgApiContext(), envId)
	return c.OrgClient.EnvironmentsOrgV2Api.GetOrgV2EnvironmentExecute(req)
}

func (c *Client) UpdateOrgEnvironment(envId string, updateEnvironment orgv2.OrgV2Environment) (orgv2.OrgV2Environment, *http.Response, error) {
	req := c.OrgClient.EnvironmentsOrgV2Api.UpdateOrgV2Environment(c.orgApiContext(), envId).OrgV2Environment(updateEnvironment)
	return c.OrgClient.EnvironmentsOrgV2Api.UpdateOrgV2EnvironmentExecute(req)
}

func (c *Client) DeleteOrgEnvironment(envId string) (*http.Response, error) {
	req := c.OrgClient.EnvironmentsOrgV2Api.DeleteOrgV2Environment(c.orgApiContext(), envId)
	return c.OrgClient.EnvironmentsOrgV2Api.DeleteOrgV2EnvironmentExecute(req)
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

		pageToken, done, err = extractOrgNextPageToken(page.GetMetadata().Next)
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
	return c.OrgClient.EnvironmentsOrgV2Api.ListOrgV2EnvironmentsExecute(req)
}

func (c *Client) GetOrgOrganization(orgId string) (orgv2.OrgV2Organization, *http.Response, error) {
	req := c.OrgClient.OrganizationsOrgV2Api.GetOrgV2Organization(c.orgApiContext(), orgId)
	return c.OrgClient.OrganizationsOrgV2Api.GetOrgV2OrganizationExecute(req)
}

func (c *Client) UpdateOrgOrganization(orgId string, updateOrganization orgv2.OrgV2Organization) (orgv2.OrgV2Organization, *http.Response, error) {
	req := c.OrgClient.OrganizationsOrgV2Api.UpdateOrgV2Organization(c.orgApiContext(), orgId).OrgV2Organization(updateOrganization)
	return c.OrgClient.OrganizationsOrgV2Api.UpdateOrgV2OrganizationExecute(req)
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

		pageToken, done, err = extractOrgNextPageToken(page.GetMetadata().Next)
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
	return c.OrgClient.OrganizationsOrgV2Api.ListOrgV2OrganizationsExecute(req)
}

func extractOrgNextPageToken(nextPageUrlStringNullable orgv2.NullableString) (string, bool, error) {
	if !nextPageUrlStringNullable.IsSet() {
		return "", true, nil
	}
	nextPageUrlString := *nextPageUrlStringNullable.Get()
	pageToken, err := extractPageToken(nextPageUrlString)
	return pageToken, false, err
}
