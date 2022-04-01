package ccloudv2

import (
	"context"
	"log"
	"net/http"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newOrgClient(baseURL, userAgent string, isTest bool) *orgv2.APIClient {
	orgServer := getServerUrl(baseURL, isTest)
	cfg := orgv2.NewConfiguration()
	cfg.Servers = orgv2.ServerConfigurations{
		{URL: orgServer, Description: "Confluent Cloud ORG"},
	}
	cfg.UserAgent = userAgent
	cfg.Debug = plog.CliLogger.GetLevel() >= plog.DEBUG
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
		environmentList, resp, err := c.executeListEnvironments(pageToken)
		if err != nil {
			log.Printf("[ERROR] Environments get failed %v, %s", resp, err)
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
