package ccloudv2

import (
	"context"
	"net/http"

	networkinggatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newNetworkingGatewayClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *networkinggatewayv1.APIClient {
	cfg := networkinggatewayv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = networkinggatewayv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return networkinggatewayv1.NewAPIClient(cfg)
}

func (c *Client) networkingGatewayApiContext() context.Context {
	return context.WithValue(context.Background(), networkinggatewayv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) GetNetworkGateway(environment, id string) (networkinggatewayv1.NetworkingV1Gateway, error) {
	resp, httpResp, err := c.NetworkingGatewayClient.GatewaysNetworkingV1Api.GetNetworkingV1Gateway(c.networkingGatewayApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteNetworkGateway(environment, id string) error {
	httpResp, err := c.NetworkingGatewayClient.GatewaysNetworkingV1Api.DeleteNetworkingV1Gateway(c.networkingGatewayApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateNetworkGateway(gateway networkinggatewayv1.NetworkingV1Gateway) (networkinggatewayv1.NetworkingV1Gateway, error) {
	resp, httpResp, err := c.NetworkingGatewayClient.GatewaysNetworkingV1Api.CreateNetworkingV1Gateway(c.networkingGatewayApiContext()).NetworkingV1Gateway(gateway).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateNetworkGateway(id string, gatewayUpdate networkinggatewayv1.NetworkingV1GatewayUpdate) (networkinggatewayv1.NetworkingV1Gateway, error) {
	resp, httpResp, err := c.NetworkingGatewayClient.GatewaysNetworkingV1Api.UpdateNetworkingV1Gateway(c.networkingGatewayApiContext(), id).NetworkingV1GatewayUpdate(gatewayUpdate).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListNetworkGateways(environment string, types, ids, regions, displayNames, phases []string) ([]networkinggatewayv1.NetworkingV1Gateway, error) {
	var list []networkinggatewayv1.NetworkingV1Gateway

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListNetworkGateways(environment, pageToken, types, ids, regions, displayNames, phases)
		if err != nil {
			return nil, err
		}
		list = append(list, page.GetData()...)

		pageToken, done, err = extractNextPageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *Client) executeListNetworkGateways(environment, pageToken string, types, ids, regions, displayNames, phases []string) (networkinggatewayv1.NetworkingV1GatewayList, error) {
	req := c.NetworkingGatewayClient.GatewaysNetworkingV1Api.ListNetworkingV1Gateways(c.networkingGatewayApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)

	if len(types) > 0 {
		req = req.GatewayType(networkinggatewayv1.MultipleSearchFilter{Items: types})
	}
	if len(ids) > 0 {
		req = req.Id(networkinggatewayv1.MultipleSearchFilter{Items: ids})
	}
	if len(regions) > 0 {
		req = req.SpecConfigRegion(networkinggatewayv1.MultipleSearchFilter{Items: regions})
	}
	if len(displayNames) > 0 {
		req = req.SpecDisplayName(networkinggatewayv1.MultipleSearchFilter{Items: displayNames})
	}
	if len(phases) > 0 {
		req = req.StatusPhase(networkinggatewayv1.MultipleSearchFilter{Items: phases})
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
