package ccloudv2

import (
	"context"
	"fmt"
	"net/http"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newCdxClient(url, userAgent string, unsafeTrace bool) *cdxv1.APIClient {
	cfg := cdxv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = cdxv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return cdxv1.NewAPIClient(cfg)
}

func (c *Client) cdxApiContext() context.Context {
	return context.WithValue(context.Background(), cdxv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) ResendInvite(shareId string) error {
	req := c.CdxClient.ProviderSharesCdxV1Api.ResendCdxV1ProviderShare(c.cdxApiContext(), shareId)
	httpResp, err := c.CdxClient.ProviderSharesCdxV1Api.ResendCdxV1ProviderShareExecute(req)
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteProviderShare(shareId string) error {
	req := c.CdxClient.ProviderSharesCdxV1Api.DeleteCdxV1ProviderShare(c.cdxApiContext(), shareId)
	httpResp, err := c.CdxClient.ProviderSharesCdxV1Api.DeleteCdxV1ProviderShareExecute(req)
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DescribeProviderShare(shareId string) (cdxv1.CdxV1ProviderShare, error) {
	req := c.CdxClient.ProviderSharesCdxV1Api.GetCdxV1ProviderShare(c.cdxApiContext(), shareId)
	resp, httpResp, err := c.CdxClient.ProviderSharesCdxV1Api.GetCdxV1ProviderShareExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteConsumerShare(shareId string) error {
	req := c.CdxClient.ConsumerSharesCdxV1Api.DeleteCdxV1ConsumerShare(c.cdxApiContext(), shareId)
	httpResp, err := c.CdxClient.ConsumerSharesCdxV1Api.DeleteCdxV1ConsumerShareExecute(req)
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DescribeConsumerShare(shareId string) (cdxv1.CdxV1ConsumerShare, error) {
	req := c.CdxClient.ConsumerSharesCdxV1Api.GetCdxV1ConsumerShare(c.cdxApiContext(), shareId)
	resp, httpResp, err := c.CdxClient.ConsumerSharesCdxV1Api.GetCdxV1ConsumerShareExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateInvite(environment, kafkaCluster, topic, email string) (cdxv1.CdxV1ProviderShare, error) {
	deliveryMethod := "Email"
	req := c.CdxClient.ProviderSharesCdxV1Api.CreateCdxV1ProviderShare(c.cdxApiContext()).
		CdxV1CreateShareRequest(cdxv1.CdxV1CreateShareRequest{
			Environment:  &environment,
			KafkaCluster: &kafkaCluster,
			ConsumerRestriction: &cdxv1.CdxV1CreateShareRequestConsumerRestrictionOneOf{
				CdxV1EmailConsumerRestriction: &cdxv1.CdxV1EmailConsumerRestriction{
					Kind:  deliveryMethod,
					Email: email,
				},
			},
			DeliveryMethod: &deliveryMethod,
			Resources:      &[]string{fmt.Sprintf("crn://confluent.cloud/kafka=%s/topic=%s", kafkaCluster, topic)},
		})
	resp, httpResp, err := c.CdxClient.ProviderSharesCdxV1Api.CreateCdxV1ProviderShareExecute(req)
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

		pageToken, done, err = extractCdxNextPageToken(page.GetMetadata().Next)
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

		pageToken, done, err = extractCdxNextPageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}

	return list, nil
}

func (c *Client) executeListConsumerShares(sharedResource, pageToken string) (cdxv1.CdxV1ConsumerShareList, *http.Response, error) {
	req := c.CdxClient.ConsumerSharesCdxV1Api.ListCdxV1ConsumerShares(c.cdxApiContext()).
		SharedResource(sharedResource).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.CdxClient.ConsumerSharesCdxV1Api.ListCdxV1ConsumerSharesExecute(req)
}

func (c *Client) executeListProviderShares(sharedResource, pageToken string) (cdxv1.CdxV1ProviderShareList, *http.Response, error) {
	req := c.CdxClient.ProviderSharesCdxV1Api.ListCdxV1ProviderShares(c.cdxApiContext()).SharedResource(sharedResource).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.CdxClient.ProviderSharesCdxV1Api.ListCdxV1ProviderSharesExecute(req)
}

func extractCdxNextPageToken(nextPageUrlStringNullable cdxv1.NullableString) (string, bool, error) {
	if !nextPageUrlStringNullable.IsSet() {
		return "", true, nil
	}
	nextPageUrlString := *nextPageUrlStringNullable.Get()
	pageToken, err := extractPageToken(nextPageUrlString)
	return pageToken, false, err
}

func (c *Client) RedeemSharedToken(token, awsAccountId, azureSubscriptionId string) (cdxv1.CdxV1RedeemTokenResponse, error) {
	redeemTokenRequest := cdxv1.CdxV1RedeemTokenRequest{
		Token:             &token,
		AwsAccount:        &awsAccountId,
		AzureSubscription: &azureSubscriptionId,
	}
	req := c.CdxClient.SharedTokensCdxV1Api.RedeemCdxV1SharedToken(c.cdxApiContext()).CdxV1RedeemTokenRequest(redeemTokenRequest)
	resp, httpResp, err := c.CdxClient.SharedTokensCdxV1Api.RedeemCdxV1SharedTokenExecute(req)
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
