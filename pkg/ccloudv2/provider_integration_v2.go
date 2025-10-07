package ccloudv2

import (
	"context"
	"net/http"

	piv2 "github.com/confluentinc/ccloud-sdk-go-v2/provider-integration/v2"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newProviderIntegrationV2Client(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *piv2.APIClient {
	configuration := piv2.NewConfiguration()
	configuration.HTTPClient = httpClient
	configuration.UserAgent = userAgent
	configuration.Servers = []piv2.ServerConfiguration{
		{
			URL: url,
		},
	}
	configuration.Debug = unsafeTrace
	return piv2.NewAPIClient(configuration)
}

func (c *Client) V2ApiContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, piv2.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

// CreatePimV2Integration creates a new provider integration
func (c *Client) CreatePimV2Integration(ctx context.Context, request piv2.PimV2Integration) (piv2.PimV2Integration, error) {
	integration, httpResp, err := c.ProviderIntegrationV2Client.IntegrationsPimV2Api.CreatePimV2Integration(c.V2ApiContext(ctx)).PimV2Integration(request).Execute()
	if err != nil {
		return piv2.PimV2Integration{}, errors.CatchCCloudV2Error(err, httpResp)
	}
	return integration, nil
}

// UpdatePimV2Integration updates a provider integration
func (c *Client) UpdatePimV2Integration(ctx context.Context, id string, request piv2.PimV2IntegrationUpdate) (piv2.PimV2Integration, error) {
	integration, httpResp, err := c.ProviderIntegrationV2Client.IntegrationsPimV2Api.UpdatePimV2Integration(c.V2ApiContext(ctx), id).PimV2IntegrationUpdate(request).Execute()
	if err != nil {
		return piv2.PimV2Integration{}, errors.CatchCCloudV2Error(err, httpResp)
	}
	return integration, nil
}

// ValidatePimV2Integration validates a provider integration
func (c *Client) ValidatePimV2Integration(ctx context.Context, request piv2.PimV2IntegrationValidateRequest) error {
	httpResp, err := c.ProviderIntegrationV2Client.IntegrationsPimV2Api.ValidatePimV2Integration(c.V2ApiContext(ctx)).PimV2IntegrationValidateRequest(request).Execute()
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}
	return nil
}

// GetPimV2Integration gets a provider integration
func (c *Client) GetPimV2Integration(ctx context.Context, id, environmentId string) (piv2.PimV2Integration, error) {
	integration, httpResp, err := c.ProviderIntegrationV2Client.IntegrationsPimV2Api.GetPimV2Integration(c.V2ApiContext(ctx), id).Environment(environmentId).Execute()
	if err != nil {
		return piv2.PimV2Integration{}, errors.CatchCCloudV2Error(err, httpResp)
	}
	return integration, nil
}

// DeletePimV2Integration deletes a provider integration
func (c *Client) DeletePimV2Integration(ctx context.Context, id, environmentId string) error {
	httpResp, err := c.ProviderIntegrationV2Client.IntegrationsPimV2Api.DeletePimV2Integration(c.V2ApiContext(ctx), id).Environment(environmentId).Execute()
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}
	return nil
}

// ListPimV2Integrations lists provider integrations with pagination support
func (c *Client) ListPimV2Integrations(ctx context.Context, environmentId string) ([]piv2.PimV2Integration, error) {
	var list []piv2.PimV2Integration
	done := false
	pageToken := ""

	for !done {
		page, httpResp, err := c.executeListPimV2Integrations(ctx, environmentId, pageToken)
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

func (c *Client) executeListPimV2Integrations(ctx context.Context, environmentId, pageToken string) (piv2.PimV2IntegrationList, *http.Response, error) {
	req := c.ProviderIntegrationV2Client.IntegrationsPimV2Api.ListPimV2Integrations(c.V2ApiContext(ctx)).Environment(environmentId).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
