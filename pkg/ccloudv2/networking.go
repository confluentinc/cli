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

func (c *Client) GetNetwork(environment, id string) (networkingv1.NetworkingV1Network, error) {
	resp, httpResp, err := c.NetworkingClient.NetworksNetworkingV1Api.GetNetworkingV1Network(c.networkingApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteNetwork(environment, id string) error {
	httpResp, err := c.NetworkingClient.NetworksNetworkingV1Api.DeleteNetworkingV1Network(c.networkingApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateNetwork(environment, id string, networkingV1NetworkUpdate networkingv1.NetworkingV1NetworkUpdate) (networkingv1.NetworkingV1Network, error) {
	resp, httpResp, err := c.NetworkingClient.NetworksNetworkingV1Api.UpdateNetworkingV1Network(c.networkingApiContext(), id).NetworkingV1NetworkUpdate(networkingV1NetworkUpdate).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateNetwork(network networkingv1.NetworkingV1Network) (networkingv1.NetworkingV1Network, error) {
	resp, httpResp, err := c.NetworkingClient.NetworksNetworkingV1Api.CreateNetworkingV1Network(c.networkingApiContext()).NetworkingV1Network(network).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListNetworks(environment string) ([]networkingv1.NetworkingV1Network, error) {
	var list []networkingv1.NetworkingV1Network

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListNetworks(environment, pageToken)
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

func (c *Client) executeListNetworks(environment, pageToken string) (networkingv1.NetworkingV1NetworkList, error) {
	req := c.NetworkingClient.NetworksNetworkingV1Api.ListNetworkingV1Networks(c.networkingApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListPeerings(environment string) ([]networkingv1.NetworkingV1Peering, error) {
	var list []networkingv1.NetworkingV1Peering

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListPeerings(environment, pageToken)
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

func (c *Client) executeListPeerings(environment, pageToken string) (networkingv1.NetworkingV1PeeringList, error) {
	req := c.NetworkingClient.PeeringsNetworkingV1Api.ListNetworkingV1Peerings(c.networkingApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetPeering(environment, id string) (networkingv1.NetworkingV1Peering, error) {
	resp, httpResp, err := c.NetworkingClient.PeeringsNetworkingV1Api.GetNetworkingV1Peering(c.networkingApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdatePeering(environment, id string, peeringUpdate networkingv1.NetworkingV1PeeringUpdate) (networkingv1.NetworkingV1Peering, error) {
	resp, httpResp, err := c.NetworkingClient.PeeringsNetworkingV1Api.UpdateNetworkingV1Peering(c.networkingApiContext(), id).NetworkingV1PeeringUpdate(peeringUpdate).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeletePeering(environment, id string) error {
	httpResp, err := c.NetworkingClient.PeeringsNetworkingV1Api.DeleteNetworkingV1Peering(c.networkingApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreatePeering(peering networkingv1.NetworkingV1Peering) (networkingv1.NetworkingV1Peering, error) {
	resp, httpResp, err := c.NetworkingClient.PeeringsNetworkingV1Api.CreateNetworkingV1Peering(c.networkingApiContext()).NetworkingV1Peering(peering).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListTransitGatewayAttachments(environment string) ([]networkingv1.NetworkingV1TransitGatewayAttachment, error) {
	var list []networkingv1.NetworkingV1TransitGatewayAttachment

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListTransitGatewayAttachments(environment, pageToken)
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

func (c *Client) executeListTransitGatewayAttachments(environment, pageToken string) (networkingv1.NetworkingV1TransitGatewayAttachmentList, error) {
	req := c.NetworkingClient.TransitGatewayAttachmentsNetworkingV1Api.ListNetworkingV1TransitGatewayAttachments(c.networkingApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetTransitGatewayAttachment(environment, id string) (networkingv1.NetworkingV1TransitGatewayAttachment, error) {
	resp, httpResp, err := c.NetworkingClient.TransitGatewayAttachmentsNetworkingV1Api.GetNetworkingV1TransitGatewayAttachment(c.networkingApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateTransitGatewayAttachment(environment, id string, transitGatewayAttachmentUpdate networkingv1.NetworkingV1TransitGatewayAttachmentUpdate) (networkingv1.NetworkingV1TransitGatewayAttachment, error) {
	resp, httpResp, err := c.NetworkingClient.TransitGatewayAttachmentsNetworkingV1Api.UpdateNetworkingV1TransitGatewayAttachment(c.networkingApiContext(), id).NetworkingV1TransitGatewayAttachmentUpdate(transitGatewayAttachmentUpdate).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteTransitGatewayAttachment(environment, id string) error {
	httpResp, err := c.NetworkingClient.TransitGatewayAttachmentsNetworkingV1Api.DeleteNetworkingV1TransitGatewayAttachment(c.networkingApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateTransitGatewayAttachment(attachment networkingv1.NetworkingV1TransitGatewayAttachment) (networkingv1.NetworkingV1TransitGatewayAttachment, error) {
	resp, httpResp, err := c.NetworkingClient.TransitGatewayAttachmentsNetworkingV1Api.CreateNetworkingV1TransitGatewayAttachment(c.networkingApiContext()).NetworkingV1TransitGatewayAttachment(attachment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
