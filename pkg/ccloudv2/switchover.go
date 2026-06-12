package ccloudv2

import (
	"context"
	"net/http"

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

func (c *Client) switchoverApiContext() context.Context {
	return context.WithValue(context.Background(), switchoverv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

// ---------------------------------------------------------------------------
// SwitchoverPair
// ---------------------------------------------------------------------------

func (c *Client) CreateSwitchoverPair(pair switchoverv1.SwitchoverV1SwitchoverPair) (switchoverv1.SwitchoverV1SwitchoverPair, error) {
	resp, httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.CreateSwitchoverV1SwitchoverPair(c.switchoverApiContext()).SwitchoverV1SwitchoverPair(pair).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DescribeSwitchoverPair(id, environment string) (switchoverv1.SwitchoverV1SwitchoverPair, error) {
	resp, httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.GetSwitchoverV1SwitchoverPair(c.switchoverApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateSwitchoverPair(id string, update switchoverv1.SwitchoverV1SwitchoverPair) (switchoverv1.SwitchoverV1SwitchoverPair, error) {
	resp, httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.UpdateSwitchoverV1SwitchoverPair(c.switchoverApiContext(), id).SwitchoverV1SwitchoverPair(update).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) FailoverSwitchoverPair(id string, request switchoverv1.SwitchoverV1SwitchoverPairFailoverRequest) (switchoverv1.SwitchoverV1SwitchoverPair, error) {
	resp, httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.FailoverSwitchoverV1SwitchoverPair(c.switchoverApiContext(), id).SwitchoverV1SwitchoverPairFailoverRequest(request).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteSwitchoverPair(id, environment string) error {
	httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.DeleteSwitchoverV1SwitchoverPair(c.switchoverApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
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

		pageToken, done, err = extractNextPageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
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

// ---------------------------------------------------------------------------
// SwitchoverEndpoint
// ---------------------------------------------------------------------------

func (c *Client) CreateSwitchoverEndpoint(endpoint switchoverv1.SwitchoverV1SwitchoverEndpoint) (switchoverv1.SwitchoverV1SwitchoverEndpoint, error) {
	resp, httpResp, err := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.CreateSwitchoverV1SwitchoverEndpoint(c.switchoverApiContext()).SwitchoverV1SwitchoverEndpoint(endpoint).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DescribeSwitchoverEndpoint(id, environment string) (switchoverv1.SwitchoverV1SwitchoverEndpoint, error) {
	resp, httpResp, err := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.GetSwitchoverV1SwitchoverEndpoint(c.switchoverApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateSwitchoverEndpoint(id string, update switchoverv1.SwitchoverV1SwitchoverEndpoint) (switchoverv1.SwitchoverV1SwitchoverEndpoint, error) {
	resp, httpResp, err := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.UpdateSwitchoverV1SwitchoverEndpoint(c.switchoverApiContext(), id).SwitchoverV1SwitchoverEndpoint(update).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ActivateSwitchoverEndpoint(id string) (switchoverv1.SwitchoverV1SwitchoverEndpoint, error) {
	resp, httpResp, err := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.ActivateSwitchoverV1SwitchoverEndpoint(c.switchoverApiContext(), id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteSwitchoverEndpoint(id, environment string) error {
	httpResp, err := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.DeleteSwitchoverV1SwitchoverEndpoint(c.switchoverApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListSwitchoverEndpoints(environment string) ([]switchoverv1.SwitchoverV1SwitchoverEndpoint, error) {
	var list []switchoverv1.SwitchoverV1SwitchoverEndpoint

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListSwitchoverEndpoints(environment, pageToken)
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

func (c *Client) executeListSwitchoverEndpoints(environment, pageToken string) (switchoverv1.SwitchoverV1SwitchoverEndpointList, *http.Response, error) {
	req := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.ListSwitchoverV1SwitchoverEndpoints(c.switchoverApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
