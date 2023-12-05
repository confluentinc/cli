package ccloudv2

import (
	"context"
	"net/http"

	networkingipv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-ip/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newNetworkingIpClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *networkingipv1.APIClient {
	cfg := networkingipv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = networkingipv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return networkingipv1.NewAPIClient(cfg)
}

func (c *Client) networkingIpApiContext() context.Context {
	return context.WithValue(context.Background(), networkingipv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) ListIpAddresses() ([]networkingipv1.NetworkingV1IpAddress, error) {
	var list []networkingipv1.NetworkingV1IpAddress

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListIpAddresses(pageToken)
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

func (c *Client) executeListIpAddresses(pageToken string) (networkingipv1.NetworkingV1IpAddressList, error) {
	req := c.NetworkingIpClient.IPAddressesNetworkingV1Api.ListNetworkingV1IpAddresses(c.networkingIpApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
