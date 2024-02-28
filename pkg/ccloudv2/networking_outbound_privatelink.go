package ccloudv2

import (
	"context"
	"net/http"

	networkingoutboundprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-outbound-privatelink/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newNetworkingOutboundPrivateLinkClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *networkingoutboundprivatelinkv1.APIClient {
	cfg := networkingoutboundprivatelinkv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = networkingoutboundprivatelinkv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return networkingoutboundprivatelinkv1.NewAPIClient(cfg)
}

func (c *Client) networkingOutboundPrivateLinkApiContext() context.Context {
	return context.WithValue(context.Background(), networkingoutboundprivatelinkv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) ListAccessPoints(environment string) ([]networkingoutboundprivatelinkv1.NetworkingV1AccessPoint, error) {
	var list []networkingoutboundprivatelinkv1.NetworkingV1AccessPoint

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListAccessPoints(environment, pageToken)
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

func (c *Client) executeListAccessPoints(environment, pageToken string) (networkingoutboundprivatelinkv1.NetworkingV1AccessPointList, error) {
	req := c.NetworkingOutboundPrivateLinkClient.AccessPointsNetworkingV1Api.ListNetworkingV1AccessPoints(c.networkingOutboundPrivateLinkApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetAccessPoint(environment, id string) (networkingoutboundprivatelinkv1.NetworkingV1AccessPoint, error) {
	resp, httpResp, err := c.NetworkingOutboundPrivateLinkClient.AccessPointsNetworkingV1Api.GetNetworkingV1AccessPoint(c.networkingOutboundPrivateLinkApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteAccessPoint(environment, id string) error {
	httpResp, err := c.NetworkingOutboundPrivateLinkClient.AccessPointsNetworkingV1Api.DeleteNetworkingV1AccessPoint(c.networkingOutboundPrivateLinkApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateAccessPoint(id string, accessPointUpdate networkingoutboundprivatelinkv1.NetworkingV1AccessPointUpdate) (networkingoutboundprivatelinkv1.NetworkingV1AccessPoint, error) {
	resp, httpResp, err := c.NetworkingOutboundPrivateLinkClient.AccessPointsNetworkingV1Api.UpdateNetworkingV1AccessPoint(c.networkingOutboundPrivateLinkApiContext(), id).NetworkingV1AccessPointUpdate(accessPointUpdate).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateAccessPoint(accessPoint networkingoutboundprivatelinkv1.NetworkingV1AccessPoint) (networkingoutboundprivatelinkv1.NetworkingV1AccessPoint, error) {
	resp, httpResp, err := c.NetworkingOutboundPrivateLinkClient.AccessPointsNetworkingV1Api.CreateNetworkingV1AccessPoint(c.networkingOutboundPrivateLinkApiContext()).NetworkingV1AccessPoint(accessPoint).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListDnsRecords(environment string) ([]networkingoutboundprivatelinkv1.NetworkingV1DnsRecord, error) {
	var list []networkingoutboundprivatelinkv1.NetworkingV1DnsRecord

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListDnsRecords(environment, pageToken)
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

func (c *Client) executeListDnsRecords(environment, pageToken string) (networkingoutboundprivatelinkv1.NetworkingV1DnsRecordList, error) {
	req := c.NetworkingOutboundPrivateLinkClient.DnsRecordsNetworkingV1Api.ListNetworkingV1DnsRecords(c.networkingOutboundPrivateLinkApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetDnsRecord(environment, id string) (networkingoutboundprivatelinkv1.NetworkingV1DnsRecord, error) {
	resp, httpResp, err := c.NetworkingOutboundPrivateLinkClient.DnsRecordsNetworkingV1Api.GetNetworkingV1DnsRecord(c.networkingOutboundPrivateLinkApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteDnsRecord(environment, id string) error {
	httpResp, err := c.NetworkingOutboundPrivateLinkClient.DnsRecordsNetworkingV1Api.DeleteNetworkingV1DnsRecord(c.networkingOutboundPrivateLinkApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateDnsRecord(id string, dnsRecordUpdate networkingoutboundprivatelinkv1.NetworkingV1DnsRecordUpdate) (networkingoutboundprivatelinkv1.NetworkingV1DnsRecord, error) {
	resp, httpResp, err := c.NetworkingOutboundPrivateLinkClient.DnsRecordsNetworkingV1Api.UpdateNetworkingV1DnsRecord(c.networkingOutboundPrivateLinkApiContext(), id).NetworkingV1DnsRecordUpdate(dnsRecordUpdate).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateDnsRecord(dnsRecord networkingoutboundprivatelinkv1.NetworkingV1DnsRecord) (networkingoutboundprivatelinkv1.NetworkingV1DnsRecord, error) {
	resp, httpResp, err := c.NetworkingOutboundPrivateLinkClient.DnsRecordsNetworkingV1Api.CreateNetworkingV1DnsRecord(c.networkingOutboundPrivateLinkApiContext()).NetworkingV1DnsRecord(dnsRecord).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
