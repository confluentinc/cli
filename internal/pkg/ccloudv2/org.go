package ccloudv2

import (
	"context"
	"net/http"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"

	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newOrgClient(baseURL, userAgent string, isTest bool) *orgv2.APIClient {
	cfg := orgv2.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = orgv2.ServerConfigurations{{URL: getServerUrl(baseURL, isTest), Description: "Confluent Cloud Org"}}
	cfg.UserAgent = userAgent

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

func (c *Client) UpdateOrgEnvironment(envId string, updateEnvironment orgv2.OrgV2Environment) (orgv2.OrgV2Environment, *http.Response, error) {
	req := c.OrgClient.EnvironmentsOrgV2Api.UpdateOrgV2Environment(c.orgApiContext(), envId).OrgV2Environment(updateEnvironment)
	return c.OrgClient.EnvironmentsOrgV2Api.UpdateOrgV2EnvironmentExecute(req)
}

func (c *Client) DeleteOrgEnvironment(envId string) (*http.Response, error) {
	req := c.OrgClient.EnvironmentsOrgV2Api.DeleteOrgV2Environment(c.orgApiContext(), envId)
	return c.OrgClient.EnvironmentsOrgV2Api.DeleteOrgV2EnvironmentExecute(req)
}

func (c *Client) ListOrgEnvironments() ([]orgv2.OrgV2Environment, error) {
	environments := make([]orgv2.OrgV2Environment, 0)

	collectedAllEnvironments := false
	pageToken := ""
	for !collectedAllEnvironments {
		environmentList, _, err := c.executeListEnvironments(pageToken)
		if err != nil {
			return nil, err
		}
		environments = append(environments, environmentList.GetData()...)

		// nextPageUrlStringNullable is nil for the last page
		nextPageUrlStringNullable := environmentList.GetMetadata().Next
		pageToken, collectedAllEnvironments, err = extractOrgNextPagePageToken(nextPageUrlStringNullable)
		if err != nil {
			return nil, err
		}
	}
	return environments, nil
}

func (c *Client) executeListEnvironments(pageToken string) (orgv2.OrgV2EnvironmentList, *http.Response, error) {
	var req orgv2.ApiListOrgV2EnvironmentsRequest
	if pageToken != "" {
		req = c.OrgClient.EnvironmentsOrgV2Api.ListOrgV2Environments(c.orgApiContext()).PageSize(ccloudV2ListPageSize).PageToken(pageToken)
	} else {
		req = c.OrgClient.EnvironmentsOrgV2Api.ListOrgV2Environments(c.orgApiContext()).PageSize(ccloudV2ListPageSize)
	}
	return c.OrgClient.EnvironmentsOrgV2Api.ListOrgV2EnvironmentsExecute(req)
}
