package ccloudv2

import (
	"context"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newNetowrkingClient(url, userAgent string, unsafeTrace bool) *networkingv1.APIClient {
	cfg := networkingv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClient(unsafeTrace)
	cfg.Servers = networkingv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return networkingv1.NewAPIClient(cfg)
}

func (c *Client) networkingApiContext() context.Context {
	return context.WithValue(context.Background(), networkingv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) GetNetwork(envId, id string) (networkingv1.NetworkingV1Network, error) {
	resp, httpResp, err := c.NetworkingClient.NetworksNetworkingV1Api.GetNetworkingV1Network(c.networkingApiContext(), id).Environment(envId).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteNetwork(envId, id string) error {
	httpResp, err := c.NetworkingClient.NetworksNetworkingV1Api.DeleteNetworkingV1Network(c.networkingApiContext(), id).Environment(envId).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateNetwork(envId, id string, updateReq networkingv1.NetworkingV1NetworkUpdate) (networkingv1.NetworkingV1Network, error) {
	resp, httpResp, err := c.NetworkingClient.NetworksNetworkingV1Api.UpdateNetworkingV1Network(c.networkingApiContext(), id).NetworkingV1NetworkUpdate(updateReq).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListNetworks(envId string, names, clouds, regions, connectionTypes, cidrs, phases []string) ([]networkingv1.NetworkingV1Network, error) {
	var list []networkingv1.NetworkingV1Network

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListNetworks(envId, pageToken, names, clouds, regions, connectionTypes, cidrs, phases)
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

func (c *Client) executeListNetworks(envId, pageToken string, names, clouds, regions, connectionTypes, cidrs, phases []string) (networkingv1.NetworkingV1NetworkList, error) {
	req := c.NetworkingClient.NetworksNetworkingV1Api.ListNetworkingV1Networks(c.networkingApiContext()).Environment(envId).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	if len(names) > 0 {
		req = req.SpecDisplayName(networkingv1.MultipleSearchFilter{Items: names})
	}

	if len(clouds) > 0 {
		req = req.SpecCloud(networkingv1.MultipleSearchFilter{Items: clouds})
	}

	if len(regions) > 0 {
		req = req.SpecRegion(networkingv1.MultipleSearchFilter{Items: regions})
	}

	if len(connectionTypes) > 0 {
		req = req.ConnectionType(networkingv1.MultipleSearchFilter{Items: connectionTypes})
	}

	if len(cidrs) > 0 {
		req = req.SpecCidr(networkingv1.MultipleSearchFilter{Items: cidrs})
	}

	if len(phases) > 0 {
		req = req.StatusPhase(networkingv1.MultipleSearchFilter{Items: phases})
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
