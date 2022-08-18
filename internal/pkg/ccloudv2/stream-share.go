package ccloudv2

import (
	"context"
	"fmt"
	"net/http"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newCdxClient(baseURL, userAgent string, isTest bool) *cdxv1.APIClient {
	cfg := cdxv1.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = cdxv1.ServerConfigurations{{URL: getServerUrl(baseURL, isTest)}}
	cfg.UserAgent = userAgent

	return cdxv1.NewAPIClient(cfg)
}

func (c *Client) cdxApiContext() context.Context {
	return context.WithValue(context.Background(), cdxv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) ResendInvite(shareId string) (*http.Response, error) {
	req := c.StreamShareClient.ProviderSharesCdxV1Api.ResendCdxV1ProviderShare(c.cdxApiContext(), shareId)
	return c.StreamShareClient.ProviderSharesCdxV1Api.ResendCdxV1ProviderShareExecute(req)
}

func (c *Client) DeleteProviderShare(shareId string) (*http.Response, error) {
	req := c.StreamShareClient.ProviderSharesCdxV1Api.DeleteCdxV1ProviderShare(c.cdxApiContext(), shareId)
	return c.StreamShareClient.ProviderSharesCdxV1Api.DeleteCdxV1ProviderShareExecute(req)
}

func (c *Client) DescribeProviderShare(shareId string) (cdxv1.CdxV1ProviderShare, *http.Response, error) {
	req := c.StreamShareClient.ProviderSharesCdxV1Api.GetCdxV1ProviderShare(c.cdxApiContext(), shareId)
	return c.StreamShareClient.ProviderSharesCdxV1Api.GetCdxV1ProviderShareExecute(req)
}

func (c *Client) DeleteConsumerShare(shareId string) (*http.Response, error) {
	req := c.StreamShareClient.ConsumerSharesCdxV1Api.DeleteCdxV1ConsumerShare(c.cdxApiContext(), shareId)
	return c.StreamShareClient.ConsumerSharesCdxV1Api.DeleteCdxV1ConsumerShareExecute(req)
}

func (c *Client) DescribeConsumerShare(shareId string) (cdxv1.CdxV1ConsumerShare, *http.Response, error) {
	req := c.StreamShareClient.ConsumerSharesCdxV1Api.GetCdxV1ConsumerShare(c.cdxApiContext(), shareId)
	return c.StreamShareClient.ConsumerSharesCdxV1Api.GetCdxV1ConsumerShareExecute(req)
}

func (c *Client) CreateInvite(environment, kafkaCluster, topic, email string) (cdxv1.CdxV1ProviderShare, *http.Response, error) {
	deliveryMethod := "Email"
	req := c.StreamShareClient.ProviderSharesCdxV1Api.CreateCdxV1ProviderShare(c.cdxApiContext()).
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
	return c.StreamShareClient.ProviderSharesCdxV1Api.CreateCdxV1ProviderShareExecute(req)
}

func (c *Client) ListProviderShares(sharedResource string) ([]cdxv1.CdxV1ProviderShare, error) {
	providerShares := make([]cdxv1.CdxV1ProviderShare, 0)

	collectedAllShares := false
	pageToken := ""
	for !collectedAllShares {
		sharesList, httpResp, err := c.executeListProviderShares(sharedResource, pageToken)
		if err != nil {
			return nil, errors.CatchV2ErrorDetailWithResponse(err, httpResp)
		}
		providerShares = append(providerShares, sharesList.GetData()...)

		// nextPageUrlStringNullable is nil for the last page
		nextPageUrlStringNullable := sharesList.GetMetadata().Next
		pageToken, collectedAllShares, err = extractCdxNextPagePageToken(nextPageUrlStringNullable)
		if err != nil {
			return nil, err
		}
	}

	return providerShares, nil
}

func (c *Client) ListConsumerShares(sharedResource string) ([]cdxv1.CdxV1ConsumerShare, error) {
	consumerShares := make([]cdxv1.CdxV1ConsumerShare, 0)

	collectedAllShares := false
	pageToken := ""
	for !collectedAllShares {
		sharesList, httpResp, err := c.executeListConsumerShares(sharedResource, pageToken)
		if err != nil {
			return nil, errors.CatchV2ErrorDetailWithResponse(err, httpResp)
		}
		consumerShares = append(consumerShares, sharesList.GetData()...)

		// nextPageUrlStringNullable is nil for the last page
		nextPageUrlStringNullable := sharesList.GetMetadata().Next
		pageToken, collectedAllShares, err = extractCdxNextPagePageToken(nextPageUrlStringNullable)
		if err != nil {
			return nil, err
		}
	}

	return consumerShares, nil
}

func (c *Client) executeListConsumerShares(sharedResource, pageToken string) (cdxv1.CdxV1ConsumerShareList, *http.Response, error) {
	req := c.StreamShareClient.ConsumerSharesCdxV1Api.ListCdxV1ConsumerShares(c.cdxApiContext()).
		SharedResource(sharedResource).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.StreamShareClient.ConsumerSharesCdxV1Api.ListCdxV1ConsumerSharesExecute(req)
}

func (c *Client) executeListProviderShares(sharedResource, pageToken string) (cdxv1.CdxV1ProviderShareList, *http.Response, error) {
	req := c.StreamShareClient.ProviderSharesCdxV1Api.ListCdxV1ProviderShares(c.cdxApiContext()).
		SharedResource(sharedResource).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.StreamShareClient.ProviderSharesCdxV1Api.ListCdxV1ProviderSharesExecute(req)
}

func (c *Client) RedeemSharedToken(token, awsAccount, azureSubscription string) (cdxv1.CdxV1RedeemTokenResponse, *http.Response, error) {
	redeemTokenRequest := cdxv1.CdxV1RedeemTokenRequest{
		Token:             &token,
		AwsAccount:        &awsAccount,
		AzureSubscription: &azureSubscription,
	}
	req := c.StreamShareClient.SharedTokensCdxV1Api.RedeemCdxV1SharedToken(c.cdxApiContext()).
		CdxV1RedeemTokenRequest(redeemTokenRequest)
	return c.StreamShareClient.SharedTokensCdxV1Api.RedeemCdxV1SharedTokenExecute(req)
}
