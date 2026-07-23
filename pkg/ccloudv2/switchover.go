package ccloudv2

import (
	"context"
	"net/http"
	"os"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/switchover/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newSwitchoverClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *switchoverv1.APIClient {
	cfg := switchoverv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = switchoverv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return switchoverv1.NewAPIClient(cfg)
}

// switchoverApiContext normally authenticates with the logged-in session's
// bearer token. Local-test-only escape hatch: if CONFLUENT_CLOUD_API_KEY and
// CONFLUENT_CLOUD_API_SECRET (a Cloud API key, `api-key create --resource
// cloud`) are set, use Basic auth instead — the stag Switchover Early Access
// gate only applies to the bearer/login path, not Cloud API keys.
func (c *Client) switchoverApiContext() context.Context {
	if key, secret := os.Getenv("CONFLUENT_CLOUD_API_KEY"), os.Getenv("CONFLUENT_CLOUD_API_SECRET"); key != "" && secret != "" {
		return context.WithValue(context.Background(), switchoverv1.ContextBasicAuth, switchoverv1.BasicAuth{UserName: key, Password: secret})
	}
	return context.WithValue(context.Background(), switchoverv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) CreateSwitchoverPair(pair switchoverv1.SwitchoverV1SwitchoverPair) (switchoverv1.SwitchoverV1SwitchoverPair, error) {
	res, httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.CreateSwitchoverV1SwitchoverPair(c.switchoverApiContext()).SwitchoverV1SwitchoverPair(pair).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetSwitchoverPair(id, environment string) (switchoverv1.SwitchoverV1SwitchoverPair, error) {
	res, httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.GetSwitchoverV1SwitchoverPair(c.switchoverApiContext(), id).Environment(environment).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateSwitchoverPair(id, environment string, update switchoverv1.SwitchoverV1SwitchoverPairUpdateRequest) (switchoverv1.SwitchoverV1SwitchoverPair, error) {
	res, httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.UpdateSwitchoverV1SwitchoverPair(c.switchoverApiContext(), id).Environment(environment).SwitchoverV1SwitchoverPairUpdateRequest(update).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) TriggerSwitchoverPairFailover(id string, req switchoverv1.SwitchoverV1SwitchoverPairFailoverRequest) (switchoverv1.SwitchoverV1SwitchoverPair, error) {
	res, httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.FailoverSwitchoverV1SwitchoverPair(c.switchoverApiContext(), id).SwitchoverV1SwitchoverPairFailoverRequest(req).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListSwitchoverPairs(environment string) ([]switchoverv1.SwitchoverV1SwitchoverPair, error) {
	var list []switchoverv1.SwitchoverV1SwitchoverPair

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListSwitchoverPairs(environment, pageToken)
		if err != nil {
			return nil, errors.CatchCCloudV2Error(err, httpResp)
		}
		list = append(list, page.GetData()...)

		metadata := page.GetMetadata()
		pagination := metadata.GetPagination()
		pageToken = pagination.GetNextPageToken()
		done = pageToken == ""
	}

	return list, nil
}

func (c *Client) executeListSwitchoverPairs(environment, pageToken string) (switchoverv1.SwitchoverV1SwitchoverPairList, *http.Response, error) {
	req := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.ListSwitchoverV1SwitchoverPairs(c.switchoverApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) DeleteSwitchoverPair(id, environment string) error {
	httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.DeleteSwitchoverV1SwitchoverPair(c.switchoverApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateSwitchoverEndpoint(endpoint switchoverv1.SwitchoverV1SwitchoverEndpoint) (switchoverv1.SwitchoverV1SwitchoverEndpoint, error) {
	res, httpResp, err := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.CreateSwitchoverV1SwitchoverEndpoint(c.switchoverApiContext()).SwitchoverV1SwitchoverEndpoint(endpoint).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetSwitchoverEndpoint(id, environment string) (switchoverv1.SwitchoverV1SwitchoverEndpoint, error) {
	res, httpResp, err := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.GetSwitchoverV1SwitchoverEndpoint(c.switchoverApiContext(), id).Environment(environment).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateSwitchoverEndpoint(id, environment string, update switchoverv1.SwitchoverV1SwitchoverEndpointUpdateRequest) (switchoverv1.SwitchoverV1SwitchoverEndpoint, error) {
	res, httpResp, err := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.UpdateSwitchoverV1SwitchoverEndpoint(c.switchoverApiContext(), id).Environment(environment).SwitchoverV1SwitchoverEndpointUpdateRequest(update).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListSwitchoverEndpoints(environment, switchoverPair string) ([]switchoverv1.SwitchoverV1SwitchoverEndpoint, error) {
	var list []switchoverv1.SwitchoverV1SwitchoverEndpoint

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListSwitchoverEndpoints(environment, switchoverPair, pageToken)
		if err != nil {
			return nil, errors.CatchCCloudV2Error(err, httpResp)
		}
		list = append(list, page.GetData()...)

		metadata := page.GetMetadata()
		pagination := metadata.GetPagination()
		pageToken = pagination.GetNextPageToken()
		done = pageToken == ""
	}

	return list, nil
}

func (c *Client) executeListSwitchoverEndpoints(environment, switchoverPair, pageToken string) (switchoverv1.SwitchoverV1SwitchoverEndpointList, *http.Response, error) {
	req := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.ListSwitchoverV1SwitchoverEndpoints(c.switchoverApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if switchoverPair != "" {
		req = req.SwitchoverPair(switchoverPair)
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) DeleteSwitchoverEndpoint(id, environment string) error {
	httpResp, err := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.DeleteSwitchoverV1SwitchoverEndpoint(c.switchoverApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}
