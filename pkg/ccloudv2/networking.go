package ccloudv2

import (
	"context"
	"net/http"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newNetworkingClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *networkingv1.APIClient {
	cfg := networkingv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = networkingv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return networkingv1.NewAPIClient(cfg)
}

func (c *Client) networkingApiContext() context.Context {
	return context.WithValue(context.Background(), networkingv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
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

func (c *Client) ListNetworks(environment string, name, cloud, region, cidr, phase, connectionType []string) ([]networkingv1.NetworkingV1Network, error) {
	var list []networkingv1.NetworkingV1Network

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListNetworks(environment, pageToken, name, cloud, region, cidr, phase, connectionType)
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

func (c *Client) executeListNetworks(environment, pageToken string, name, cloud, region, cidr, phase, connectionType []string) (networkingv1.NetworkingV1NetworkList, error) {
	req := c.NetworkingClient.NetworksNetworkingV1Api.ListNetworkingV1Networks(c.networkingApiContext()).Environment(environment).SpecDisplayName(name).SpecCloud(cloud).SpecRegion(region).SpecCidr(cidr).SpecConnectionTypes(connectionType).StatusPhase(phase).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListPeerings(environment string, name, network, phase []string) ([]networkingv1.NetworkingV1Peering, error) {
	var list []networkingv1.NetworkingV1Peering

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListPeerings(environment, pageToken, name, network, phase)
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

func (c *Client) executeListPeerings(environment, pageToken string, name, network, phase []string) (networkingv1.NetworkingV1PeeringList, error) {
	req := c.NetworkingClient.PeeringsNetworkingV1Api.ListNetworkingV1Peerings(c.networkingApiContext()).Environment(environment).SpecDisplayName(name).SpecNetwork(network).StatusPhase(phase).PageSize(ccloudV2ListPageSize)
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

func (c *Client) ListTransitGatewayAttachments(environment string, name, network, phase []string) ([]networkingv1.NetworkingV1TransitGatewayAttachment, error) {
	var list []networkingv1.NetworkingV1TransitGatewayAttachment

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListTransitGatewayAttachments(environment, pageToken, name, network, phase)
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

func (c *Client) executeListTransitGatewayAttachments(environment, pageToken string, name, network, phase []string) (networkingv1.NetworkingV1TransitGatewayAttachmentList, error) {
	req := c.NetworkingClient.TransitGatewayAttachmentsNetworkingV1Api.ListNetworkingV1TransitGatewayAttachments(c.networkingApiContext()).Environment(environment).SpecDisplayName(name).SpecNetwork(network).StatusPhase(phase).PageSize(ccloudV2ListPageSize)
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

func (c *Client) ListPrivateLinkAccesses(environment string, name, network, phase []string) ([]networkingv1.NetworkingV1PrivateLinkAccess, error) {
	var list []networkingv1.NetworkingV1PrivateLinkAccess

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListPrivateLinkAccesses(environment, pageToken, name, network, phase)
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

func (c *Client) executeListPrivateLinkAccesses(environment, pageToken string, name, network, phase []string) (networkingv1.NetworkingV1PrivateLinkAccessList, error) {
	req := c.NetworkingClient.PrivateLinkAccessesNetworkingV1Api.ListNetworkingV1PrivateLinkAccesses(c.networkingApiContext()).Environment(environment).SpecDisplayName(name).SpecNetwork(network).StatusPhase(phase).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetPrivateLinkAccess(environment, id string) (networkingv1.NetworkingV1PrivateLinkAccess, error) {
	resp, httpResp, err := c.NetworkingClient.PrivateLinkAccessesNetworkingV1Api.GetNetworkingV1PrivateLinkAccess(c.networkingApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdatePrivateLinkAccess(environment, id string, privateLinkAccessUpdate networkingv1.NetworkingV1PrivateLinkAccessUpdate) (networkingv1.NetworkingV1PrivateLinkAccess, error) {
	resp, httpResp, err := c.NetworkingClient.PrivateLinkAccessesNetworkingV1Api.UpdateNetworkingV1PrivateLinkAccess(c.networkingApiContext(), id).NetworkingV1PrivateLinkAccessUpdate(privateLinkAccessUpdate).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeletePrivateLinkAccess(environment, id string) error {
	httpResp, err := c.NetworkingClient.PrivateLinkAccessesNetworkingV1Api.DeleteNetworkingV1PrivateLinkAccess(c.networkingApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreatePrivateLinkAccess(access networkingv1.NetworkingV1PrivateLinkAccess) (networkingv1.NetworkingV1PrivateLinkAccess, error) {
	resp, httpResp, err := c.NetworkingClient.PrivateLinkAccessesNetworkingV1Api.CreateNetworkingV1PrivateLinkAccess(c.networkingApiContext()).NetworkingV1PrivateLinkAccess(access).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListNetworkLinkServices(environment string, name, network, phase []string) ([]networkingv1.NetworkingV1NetworkLinkService, error) {
	var list []networkingv1.NetworkingV1NetworkLinkService

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListNetworkLinkServices(environment, pageToken, name, network, phase)
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

func (c *Client) executeListNetworkLinkServices(environment, pageToken string, name, network, phase []string) (networkingv1.NetworkingV1NetworkLinkServiceList, error) {
	req := c.NetworkingClient.NetworkLinkServicesNetworkingV1Api.ListNetworkingV1NetworkLinkServices(c.networkingApiContext()).Environment(environment).SpecDisplayName(name).SpecNetwork(network).StatusPhase(phase).PageSize(ccloudV2ListPageSize)

	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetNetworkLinkService(environment, id string) (networkingv1.NetworkingV1NetworkLinkService, error) {
	resp, httpResp, err := c.NetworkingClient.NetworkLinkServicesNetworkingV1Api.GetNetworkingV1NetworkLinkService(c.networkingApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteNetworkLinkService(environment, id string) error {
	httpResp, err := c.NetworkingClient.NetworkLinkServicesNetworkingV1Api.DeleteNetworkingV1NetworkLinkService(c.networkingApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateNetworkLinkService(service networkingv1.NetworkingV1NetworkLinkService) (networkingv1.NetworkingV1NetworkLinkService, error) {
	resp, httpResp, err := c.NetworkingClient.NetworkLinkServicesNetworkingV1Api.CreateNetworkingV1NetworkLinkService(c.networkingApiContext()).NetworkingV1NetworkLinkService(service).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateNetworkLinkService(id string, networkLinkServiceUpdate networkingv1.NetworkingV1NetworkLinkServiceUpdate) (networkingv1.NetworkingV1NetworkLinkService, error) {
	resp, httpResp, err := c.NetworkingClient.NetworkLinkServicesNetworkingV1Api.UpdateNetworkingV1NetworkLinkService(c.networkingApiContext(), id).NetworkingV1NetworkLinkServiceUpdate(networkLinkServiceUpdate).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListNetworkLinkEndpoints(environment string, name, network, phase, service []string) ([]networkingv1.NetworkingV1NetworkLinkEndpoint, error) {
	var list []networkingv1.NetworkingV1NetworkLinkEndpoint

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListNetworkLinkEndpoints(environment, pageToken, name, network, phase, service)
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

func (c *Client) executeListNetworkLinkEndpoints(environment, pageToken string, name, network, phase, service []string) (networkingv1.NetworkingV1NetworkLinkEndpointList, error) {
	req := c.NetworkingClient.NetworkLinkEndpointsNetworkingV1Api.ListNetworkingV1NetworkLinkEndpoints(c.networkingApiContext()).Environment(environment).SpecDisplayName(name).SpecNetwork(network).SpecNetworkLinkService(service).StatusPhase(phase).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetNetworkLinkEndpoint(environment, id string) (networkingv1.NetworkingV1NetworkLinkEndpoint, error) {
	resp, httpResp, err := c.NetworkingClient.NetworkLinkEndpointsNetworkingV1Api.GetNetworkingV1NetworkLinkEndpoint(c.networkingApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteNetworkLinkEndpoint(environment, id string) error {
	httpResp, err := c.NetworkingClient.NetworkLinkEndpointsNetworkingV1Api.DeleteNetworkingV1NetworkLinkEndpoint(c.networkingApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateNetworkLinkEndpoint(endpoint networkingv1.NetworkingV1NetworkLinkEndpoint) (networkingv1.NetworkingV1NetworkLinkEndpoint, error) {
	resp, httpResp, err := c.NetworkingClient.NetworkLinkEndpointsNetworkingV1Api.CreateNetworkingV1NetworkLinkEndpoint(c.networkingApiContext()).NetworkingV1NetworkLinkEndpoint(endpoint).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateNetworkLinkEndpoint(id string, networkLinkEndpointUpdate networkingv1.NetworkingV1NetworkLinkEndpointUpdate) (networkingv1.NetworkingV1NetworkLinkEndpoint, error) {
	resp, httpResp, err := c.NetworkingClient.NetworkLinkEndpointsNetworkingV1Api.UpdateNetworkingV1NetworkLinkEndpoint(c.networkingApiContext(), id).NetworkingV1NetworkLinkEndpointUpdate(networkLinkEndpointUpdate).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
