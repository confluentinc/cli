package ccloudv2

import (
	"context"
	"net/http"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
)

func NewOrgClient(baseURL string, isTest bool) *orgv2.APIClient {
	orgServer := getServerUrl(baseURL, isTest)
	cfg := orgv2.NewConfiguration()
	cfg.Servers = orgv2.ServerConfigurations{
		{URL: orgServer, Description: "Confluent Cloud ORG"},
	}
	return orgv2.NewAPIClient(cfg)
}

func (c *Client) orgApiContext() context.Context {
	auth := context.WithValue(context.Background(), orgv2.ContextAccessToken, c.AuthToken)
	return auth
}

func (c *Client) CreateOrgEnvironment(environment orgv2.OrgV2Environment) (orgv2.OrgV2Environment, *http.Response, error) {
	req := c.OrgClient.EnvironmentsOrgV2Api.CreateOrgV2Environment(c.orgApiContext()).OrgV2Environment(environment)
	return c.OrgClient.EnvironmentsOrgV2Api.CreateOrgV2EnvironmentExecute(req)
}

func (c *Client) GetOrgEnvironment(envId string) (orgv2.OrgV2Environment, *http.Response, error) {
	req := c.OrgClient.EnvironmentsOrgV2Api.GetOrgV2Environment(c.orgApiContext(), envId)
	return c.OrgClient.EnvironmentsOrgV2Api.GetOrgV2EnvironmentExecute(req)
}

func (c *Client) ListEnvironments() (orgv2.OrgV2EnvironmentList, *http.Response, error) {
	req := c.OrgClient.EnvironmentsOrgV2Api.ListOrgV2Environments(c.orgApiContext()).PageSize(100)
	return c.OrgClient.EnvironmentsOrgV2Api.ListOrgV2EnvironmentsExecute(req)
}

func (c *Client) UpdateOrgEnvironment(envId string, updateEnvironment orgv2.OrgV2Environment) (orgv2.OrgV2Environment, *http.Response, error) {
	req := c.OrgClient.EnvironmentsOrgV2Api.UpdateOrgV2Environment(c.orgApiContext(), envId).OrgV2Environment(updateEnvironment)
	return c.OrgClient.EnvironmentsOrgV2Api.UpdateOrgV2EnvironmentExecute(req)
}

func (c *Client) DeleteOrgEnvironment(envId string) (*http.Response, error) {
	req := c.OrgClient.EnvironmentsOrgV2Api.DeleteOrgV2Environment(c.orgApiContext(), envId)
	return c.OrgClient.EnvironmentsOrgV2Api.DeleteOrgV2EnvironmentExecute(req)
}
