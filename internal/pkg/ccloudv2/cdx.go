package ccloudv2

import (
	"context"
	"net/http"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newCdxClient(url, userAgent string, unsafeTrace bool) *cdxv1.APIClient {
	cfg := cdxv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClient(unsafeTrace)
	cfg.Servers = cdxv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return cdxv1.NewAPIClient(cfg)
}

func (c *Client) cdxApiContext() context.Context {
	return context.WithValue(context.Background(), cdxv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) ResendInvite(shareId string) error {
	httpResp, err := c.CdxClient.ProviderSharesCdxV1Api.ResendCdxV1ProviderShare(c.cdxApiContext(), shareId).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteProviderShare(shareId string) error {
	httpResp, err := c.CdxClient.ProviderSharesCdxV1Api.DeleteCdxV1ProviderShare(c.cdxApiContext(), shareId).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DescribeProviderShare(shareId string) (cdxv1.CdxV1ProviderShare, error) {
	resp, httpResp, err := c.CdxClient.ProviderSharesCdxV1Api.GetCdxV1ProviderShare(c.cdxApiContext(), shareId).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteConsumerShare(shareId string) error {
	httpResp, err := c.CdxClient.ConsumerSharesCdxV1Api.DeleteCdxV1ConsumerShare(c.cdxApiContext(), shareId).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DescribeConsumerShare(shareId string) (cdxv1.CdxV1ConsumerShare, error) {
	resp, httpResp, err := c.CdxClient.ConsumerSharesCdxV1Api.GetCdxV1ConsumerShare(c.cdxApiContext(), shareId).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateProviderInvite(shareReq cdxv1.CdxV1CreateProviderShareRequest) (cdxv1.CdxV1ProviderShare, error) {
	resp, httpResp, err := c.CdxClient.ProviderSharesCdxV1Api.CreateCdxV1ProviderShare(c.cdxApiContext()).CdxV1CreateProviderShareRequest(shareReq).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListProviderShares(sharedResource string) ([]cdxv1.CdxV1ProviderShare, error) {
	var list []cdxv1.CdxV1ProviderShare

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListProviderShares(sharedResource, pageToken)
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

func (c *Client) ListConsumerShares(sharedResource string) ([]cdxv1.CdxV1ConsumerShare, error) {
	var list []cdxv1.CdxV1ConsumerShare

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListConsumerShares(sharedResource, pageToken)
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

func (c *Client) ListConsumerSharedResources(streamShareId string) ([]cdxv1.CdxV1ConsumerSharedResource, error) {
	var list []cdxv1.CdxV1ConsumerSharedResource

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListConsumerSharedResources(streamShareId, pageToken)
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

func (c *Client) executeListConsumerSharedResources(streamShareId, pageToken string) (cdxv1.CdxV1ConsumerSharedResourceList, *http.Response, error) {
	req := c.CdxClient.ConsumerSharedResourcesCdxV1Api.ListCdxV1ConsumerSharedResources(c.cdxApiContext()).StreamShare(streamShareId).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) executeListConsumerShares(sharedResource, pageToken string) (cdxv1.CdxV1ConsumerShareList, *http.Response, error) {
	req := c.CdxClient.ConsumerSharesCdxV1Api.ListCdxV1ConsumerShares(c.cdxApiContext()).SharedResource(sharedResource).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) executeListProviderShares(sharedResource, pageToken string) (cdxv1.CdxV1ProviderShareList, *http.Response, error) {
	req := c.CdxClient.ProviderSharesCdxV1Api.ListCdxV1ProviderShares(c.cdxApiContext()).SharedResource(sharedResource).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) RedeemSharedToken(token, awsAccountId, azureSubscriptionId, gcpProjectId string) (cdxv1.CdxV1RedeemTokenResponse, error) {
	redeemTokenRequest := cdxv1.CdxV1RedeemTokenRequest{
		Token:             &token,
		AwsAccount:        &awsAccountId,
		AzureSubscription: &azureSubscriptionId,
		GcpProject:        &gcpProjectId,
	}
	resp, httpResp, err := c.CdxClient.SharedTokensCdxV1Api.RedeemCdxV1SharedToken(c.cdxApiContext()).CdxV1RedeemTokenRequest(redeemTokenRequest).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetPrivateLinkNetworkConfig(sharedResourceId string) (cdxv1.CdxV1Network, error) {
	resp, httpResp, err := c.CdxClient.ConsumerSharedResourcesCdxV1Api.NetworkCdxV1ConsumerSharedResource(c.cdxApiContext(), sharedResourceId).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) StreamShareOptInOrOut(status bool) (cdxv1.CdxV1OptIn, error) {
	resp, httpResp, err := c.CdxClient.OptInsCdxV1Api.UpdateCdxV1OptIn(c.cdxApiContext()).CdxV1OptIn(cdxv1.CdxV1OptIn{StreamShareEnabled: &status}).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
