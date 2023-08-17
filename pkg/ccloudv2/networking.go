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

func (c *Client) UpdateNetwork(envId, id string, update networkingv1.NetworkingV1NetworkUpdate) (networkingv1.NetworkingV1Network, error) {
	update.Spec.SetEnvironment(networkingv1.ObjectReference{Id: envId})

	resp, httpResp, err := c.NetworkingClient.NetworksNetworkingV1Api.UpdateNetworkingV1Network(c.networkingApiContext(), id).NetworkingV1NetworkUpdate(update).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
