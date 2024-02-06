package ccloudv2

import (
	"context"
	"net/http"

	networkingdnsforwarderv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking-dnsforwarder/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newNetworkingDnsForwarderClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *networkingdnsforwarderv1.APIClient {
	cfg := networkingdnsforwarderv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = networkingdnsforwarderv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return networkingdnsforwarderv1.NewAPIClient(cfg)
}

func (c *Client) networkingDnsForwarderApiContext() context.Context {
	return context.WithValue(context.Background(), networkingdnsforwarderv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) ListDnsForwarders(environment string) ([]networkingdnsforwarderv1.NetworkingV1DnsForwarder, error) {
	var list []networkingdnsforwarderv1.NetworkingV1DnsForwarder

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListDnsForwarders(environment, pageToken)
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

func (c *Client) executeListDnsForwarders(environment, pageToken string) (networkingdnsforwarderv1.NetworkingV1DnsForwarderList, error) {
	req := c.NetworkingDnsForwarderClient.DnsForwardersNetworkingV1Api.ListNetworkingV1DnsForwarders(c.networkingDnsForwarderApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetDnsForwarder(environment, id string) (networkingdnsforwarderv1.NetworkingV1DnsForwarder, error) {
	resp, httpResp, err := c.NetworkingDnsForwarderClient.DnsForwardersNetworkingV1Api.GetNetworkingV1DnsForwarder(c.networkingDnsForwarderApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteDnsForwarder(environment, id string) error {
	httpResp, err := c.NetworkingDnsForwarderClient.DnsForwardersNetworkingV1Api.DeleteNetworkingV1DnsForwarder(c.networkingDnsForwarderApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateDnsForwarder(forwarder networkingdnsforwarderv1.NetworkingV1DnsForwarder) (networkingdnsforwarderv1.NetworkingV1DnsForwarder, error) {
	resp, httpResp, err := c.NetworkingDnsForwarderClient.DnsForwardersNetworkingV1Api.CreateNetworkingV1DnsForwarder(c.networkingDnsForwarderApiContext()).NetworkingV1DnsForwarder(forwarder).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateDnsForwarder(environment, id string, dnsForwarderUpdate networkingdnsforwarderv1.NetworkingV1DnsForwarder) (networkingdnsforwarderv1.NetworkingV1DnsForwarder, error) {
	resp, httpResp, err := c.NetworkingDnsForwarderClient.DnsForwardersNetworkingV1Api.UpdateNetworkingV1DnsForwarder(c.networkingDnsForwarderApiContext(), id).NetworkingV1DnsForwarder(dnsForwarderUpdate).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
