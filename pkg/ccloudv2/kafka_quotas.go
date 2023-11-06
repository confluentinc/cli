package ccloudv2

import (
	"context"
	"net/http"

	kafkaquotasv1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

func newKafkaQuotasClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *kafkaquotasv1.APIClient {
	cfg := kafkaquotasv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = kafkaquotasv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return kafkaquotasv1.NewAPIClient(cfg)
}

func (c *Client) kafkaQuotasContext() context.Context {
	return context.WithValue(context.Background(), kafkaquotasv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) ListKafkaQuotas(clusterId, envId string) ([]kafkaquotasv1.KafkaQuotasV1ClientQuota, error) {
	var list []kafkaquotasv1.KafkaQuotasV1ClientQuota

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.listQuotas(clusterId, envId, pageToken)
		if err != nil {
			return nil, errors.CatchCCloudV2Error(err, httpResp)
		}
		list = append(list, page.GetData()...)

		// nextPageUrlStringNullable is nil for the last page
		pageToken, done, err = extractNextPageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *Client) listQuotas(clusterId, envId, pageToken string) (kafkaquotasv1.KafkaQuotasV1ClientQuotaList, *http.Response, error) {
	req := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.ListKafkaQuotasV1ClientQuotas(c.kafkaQuotasContext()).PageSize(ccloudV2ListPageSize).SpecCluster(clusterId).Environment(envId)
	if pageToken != "" {
		req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) CreateKafkaQuota(quota kafkaquotasv1.KafkaQuotasV1ClientQuota) (kafkaquotasv1.KafkaQuotasV1ClientQuota, error) {
	resp, httpResp, err := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.CreateKafkaQuotasV1ClientQuota(c.kafkaQuotasContext()).KafkaQuotasV1ClientQuota(quota).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateKafkaQuota(quota kafkaquotasv1.KafkaQuotasV1ClientQuotaUpdate) (kafkaquotasv1.KafkaQuotasV1ClientQuota, error) {
	resp, httpResp, err := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.UpdateKafkaQuotasV1ClientQuota(c.kafkaQuotasContext(), *quota.Id).KafkaQuotasV1ClientQuotaUpdate(quota).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DescribeKafkaQuota(quotaId string) (kafkaquotasv1.KafkaQuotasV1ClientQuota, error) {
	resp, httpResp, err := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.GetKafkaQuotasV1ClientQuota(c.kafkaQuotasContext(), quotaId).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteKafkaQuota(quotaId string) error {
	httpResp, err := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.DeleteKafkaQuotasV1ClientQuota(c.kafkaQuotasContext(), quotaId).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}
