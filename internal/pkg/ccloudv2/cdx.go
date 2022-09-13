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

func (c *Client) ResendInvite(shareId string) (*http.Response, error) {
	req := c.CdxClient.ProviderSharesCdxV1Api.ResendCdxV1ProviderShare(c.cdxApiContext(), shareId)
	return c.CdxClient.ProviderSharesCdxV1Api.ResendCdxV1ProviderShareExecute(req)
}

func (c *Client) DeleteProviderShare(shareId string) (*http.Response, error) {
	req := c.CdxClient.ProviderSharesCdxV1Api.DeleteCdxV1ProviderShare(c.cdxApiContext(), shareId)
	return c.CdxClient.ProviderSharesCdxV1Api.DeleteCdxV1ProviderShareExecute(req)
}

func (c *Client) DescribeProviderShare(shareId string) (cdxv1.CdxV1ProviderShare, *http.Response, error) {
	req := c.CdxClient.ProviderSharesCdxV1Api.GetCdxV1ProviderShare(c.cdxApiContext(), shareId)
	return c.CdxClient.ProviderSharesCdxV1Api.GetCdxV1ProviderShareExecute(req)
}

func (c *Client) DeleteConsumerShare(shareId string) (*http.Response, error) {
	req := c.CdxClient.ConsumerSharesCdxV1Api.DeleteCdxV1ConsumerShare(c.cdxApiContext(), shareId)
	return c.CdxClient.ConsumerSharesCdxV1Api.DeleteCdxV1ConsumerShareExecute(req)
}

func (c *Client) DescribeConsumerShare(shareId string) (cdxv1.CdxV1ConsumerShare, *http.Response, error) {
	req := c.CdxClient.ConsumerSharesCdxV1Api.GetCdxV1ConsumerShare(c.cdxApiContext(), shareId)
	return c.CdxClient.ConsumerSharesCdxV1Api.GetCdxV1ConsumerShareExecute(req)
}

func (c *Client) CreateInvite(environment, kafkaCluster, topic, email string) (cdxv1.CdxV1ProviderShare, *http.Response, error) {
	deliveryMethod := "Email"
	req := c.CdxClient.ProviderSharesCdxV1Api.CreateCdxV1ProviderShare(c.cdxApiContext()).
		CdxV1CreateProviderShareRequest(cdxv1.CdxV1CreateProviderShareRequest{
			ConsumerRestriction: &cdxv1.CdxV1CreateProviderShareRequestConsumerRestrictionOneOf{
				CdxV1EmailConsumerRestriction: &cdxv1.CdxV1EmailConsumerRestriction{
					Kind:  deliveryMethod,
					Email: email,
				},
			},
			DeliveryMethod: &deliveryMethod,
			Resources:      &[]string{fmt.Sprintf("crn://confluent.cloud/kafka=%s/topic=%s/environment=%s", kafkaCluster, topic, environment)},
		})
	return c.CdxClient.ProviderSharesCdxV1Api.CreateCdxV1ProviderShareExecute(req)
}

func (c *Client) ListProviderShares(sharedResource string) ([]cdxv1.CdxV1ProviderShare, error) {
	var list []cdxv1.CdxV1ProviderShare

	done := false
	pageToken := ""
	for !done {
		page, r, err := c.executeListProviderShares(sharedResource, pageToken)
		if err != nil {
			return nil, errors.CatchV2ErrorDetailWithResponse(err, r)
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
		page, r, err := c.executeListConsumerShares(sharedResource, pageToken)
		if err != nil {
			return nil, errors.CatchV2ErrorDetailWithResponse(err, r)
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

func (c *Client) RedeemSharedToken(token, awsAccountId, azureSubscriptionId string) (cdxv1.CdxV1RedeemTokenResponse, *http.Response, error) {
	redeemTokenRequest := cdxv1.CdxV1RedeemTokenRequest{
		Token:             &token,
		AwsAccount:        &awsAccountId,
		AzureSubscription: &azureSubscriptionId,
	}
	req := c.CdxClient.SharedTokensCdxV1Api.RedeemCdxV1SharedToken(c.cdxApiContext()).CdxV1RedeemTokenRequest(redeemTokenRequest)
	return c.CdxClient.SharedTokensCdxV1Api.RedeemCdxV1SharedTokenExecute(req)
}
