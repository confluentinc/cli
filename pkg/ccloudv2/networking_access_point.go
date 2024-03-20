package ccloudv2

import (
	"context"
	"net/http"

	networkingaccesspointv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

type DnsRecordListParameters struct {
	Domains     []string
	Gateways    []string
	Names       []string
	ResourceIds []string
}

func newNetworkingAccessPointClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *networkingaccesspointv1.APIClient {
	cfg := networkingaccesspointv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = networkingaccesspointv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return networkingaccesspointv1.NewAPIClient(cfg)
}

func (c *Client) networkingAccessPointApiContext() context.Context {
	return context.WithValue(context.Background(), networkingaccesspointv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) ListAccessPoints(environment string, names []string) ([]networkingaccesspointv1.NetworkingV1AccessPoint, error) {
	var list []networkingaccesspointv1.NetworkingV1AccessPoint

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListAccessPoints(environment, pageToken, names)
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

func (c *Client) executeListAccessPoints(environment, pageToken string, names []string) (networkingaccesspointv1.NetworkingV1AccessPointList, error) {
	req := c.NetworkingAccessPointClient.AccessPointsNetworkingV1Api.ListNetworkingV1AccessPoints(c.networkingAccessPointApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	if names != nil {
		req = req.SpecDisplayName(names)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetAccessPoint(environment, id string) (networkingaccesspointv1.NetworkingV1AccessPoint, error) {
	resp, httpResp, err := c.NetworkingAccessPointClient.AccessPointsNetworkingV1Api.GetNetworkingV1AccessPoint(c.networkingAccessPointApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteAccessPoint(environment, id string) error {
	httpResp, err := c.NetworkingAccessPointClient.AccessPointsNetworkingV1Api.DeleteNetworkingV1AccessPoint(c.networkingAccessPointApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateAccessPoint(id string, accessPointUpdate networkingaccesspointv1.NetworkingV1AccessPointUpdate) (networkingaccesspointv1.NetworkingV1AccessPoint, error) {
	resp, httpResp, err := c.NetworkingAccessPointClient.AccessPointsNetworkingV1Api.UpdateNetworkingV1AccessPoint(c.networkingAccessPointApiContext(), id).NetworkingV1AccessPointUpdate(accessPointUpdate).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateAccessPoint(accessPoint networkingaccesspointv1.NetworkingV1AccessPoint) (networkingaccesspointv1.NetworkingV1AccessPoint, error) {
	resp, httpResp, err := c.NetworkingAccessPointClient.AccessPointsNetworkingV1Api.CreateNetworkingV1AccessPoint(c.networkingAccessPointApiContext()).NetworkingV1AccessPoint(accessPoint).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListDnsRecords(environment string, listParameters DnsRecordListParameters) ([]networkingaccesspointv1.NetworkingV1DnsRecord, error) {
	var list []networkingaccesspointv1.NetworkingV1DnsRecord

	done := false
	pageToken := ""
	for !done {
		page, err := c.executeListDnsRecords(environment, pageToken, listParameters)
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

func (c *Client) executeListDnsRecords(environment, pageToken string, listParameters DnsRecordListParameters) (networkingaccesspointv1.NetworkingV1DnsRecordList, error) {
	req := c.NetworkingAccessPointClient.DNSRecordsNetworkingV1Api.ListNetworkingV1DnsRecords(c.networkingAccessPointApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	if listParameters.Gateways != nil {
		req = req.SpecGateway(listParameters.Gateways)
	}

	if listParameters.Domains != nil {
		req = req.SpecDomain(listParameters.Domains)
	}

	if listParameters.Names != nil {
		req = req.SpecDisplayName(listParameters.Names)
	}

	if listParameters.ResourceIds != nil {
		req = req.ResourceId(listParameters.ResourceIds)
	}

	resp, httpResp, err := req.Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetDnsRecord(environment, id string) (networkingaccesspointv1.NetworkingV1DnsRecord, error) {
	resp, httpResp, err := c.NetworkingAccessPointClient.DNSRecordsNetworkingV1Api.GetNetworkingV1DnsRecord(c.networkingAccessPointApiContext(), id).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteDnsRecord(environment, id string) error {
	httpResp, err := c.NetworkingAccessPointClient.DNSRecordsNetworkingV1Api.DeleteNetworkingV1DnsRecord(c.networkingAccessPointApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateDnsRecord(id string, dnsRecordUpdate networkingaccesspointv1.NetworkingV1DnsRecordUpdate) (networkingaccesspointv1.NetworkingV1DnsRecord, error) {
	resp, httpResp, err := c.NetworkingAccessPointClient.DNSRecordsNetworkingV1Api.UpdateNetworkingV1DnsRecord(c.networkingAccessPointApiContext(), id).NetworkingV1DnsRecordUpdate(dnsRecordUpdate).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateDnsRecord(dnsRecord networkingaccesspointv1.NetworkingV1DnsRecord) (networkingaccesspointv1.NetworkingV1DnsRecord, error) {
	resp, httpResp, err := c.NetworkingAccessPointClient.DNSRecordsNetworkingV1Api.CreateNetworkingV1DnsRecord(c.networkingAccessPointApiContext()).NetworkingV1DnsRecord(dnsRecord).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
