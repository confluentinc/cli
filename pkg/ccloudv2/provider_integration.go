package ccloudv2

import (
	"context"
	"net/http"

	pi "github.com/confluentinc/ccloud-sdk-go-v2/provider-integration/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newProviderIntegrationClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *pi.APIClient {
	cfg := pi.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = pi.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return pi.NewAPIClient(cfg)
}

func (c *Client) providerIntegrationApiContext() context.Context {
	return context.WithValue(context.Background(), pi.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) CreateProviderIntegration(providerIntegration pi.PimV1Integration) (pi.PimV1Integration, error) {
	res, httpResp, err := c.ProviderIntegrationClient.IntegrationsPimV1Api.CreatePimV1Integration(c.providerIntegrationApiContext()).PimV1Integration(providerIntegration).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DescribeProviderIntegration(id, environment string) (pi.PimV1Integration, error) {
	res, httpResp, err := c.ProviderIntegrationClient.IntegrationsPimV1Api.GetPimV1Integration(c.providerIntegrationApiContext(), id).Environment(environment).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListProviderIntegrations(cloud, environment string) ([]pi.PimV1Integration, error) {
	var list []pi.PimV1Integration
	done := false
	pageToken := ""

	for !done {
		page, httpResp, err := c.executeListProviderIntegration(cloud, pageToken, environment)
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

func (c *Client) DeleteProviderIntegration(id, environment string) error {
	httpResp, err := c.ProviderIntegrationClient.IntegrationsPimV1Api.DeletePimV1Integration(c.providerIntegrationApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) executeListProviderIntegration(provider, pageToken, environment string) (pi.PimV1IntegrationList, *http.Response, error) {
	req := c.ProviderIntegrationClient.IntegrationsPimV1Api.ListPimV1Integrations(c.providerIntegrationApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if provider != "" {
		req = req.Provider(provider)
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
