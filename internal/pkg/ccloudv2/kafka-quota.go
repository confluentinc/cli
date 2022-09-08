package ccloudv2

import (
	"context"
	"net/http"

	kafkaquotasv1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"
)

func newKafkaQuotasClient(url, userAgent string, unsafeTrace bool) *kafkaquotasv1.APIClient {
	cfg := kafkaquotasv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = kafkaquotasv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return kafkaquotasv1.NewAPIClient(cfg)
}

func (c *Client) quotaContext() context.Context {
	return context.WithValue(context.Background(), kafkaquotasv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) ListKafkaQuotas(clusterId, envId string) ([]kafkaquotasv1.KafkaQuotasV1ClientQuota, *http.Response, error) {
	var list []kafkaquotasv1.KafkaQuotasV1ClientQuota

	done := false
	pageToken := ""
	for !done {
		page, resp, err := c.listQuotas(clusterId, envId, pageToken)
		if err != nil {
			return nil, resp, err
		}
		list = append(list, page.GetData()...)

		// nextPageUrlStringNullable is nil for the last page
		pageToken, done, err = extractKafkaQuotasNextPagePageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, nil, err
		}
	}
	return list, nil, nil
}

func (c *Client) listQuotas(clusterId, envId, pageToken string) (kafkaquotasv1.KafkaQuotasV1ClientQuotaList, *http.Response, error) {
	req := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.ListKafkaQuotasV1ClientQuotas(c.quotaContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req.PageToken(pageToken)
	}
	req = req.Cluster(clusterId).Environment(envId)
	return c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.ListKafkaQuotasV1ClientQuotasExecute(req)
}

func (c *Client) CreateKafkaQuota(quota kafkaquotasv1.KafkaQuotasV1ClientQuota) (kafkaquotasv1.KafkaQuotasV1ClientQuota, *http.Response, error) {
	req := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.CreateKafkaQuotasV1ClientQuota(c.quotaContext()).KafkaQuotasV1ClientQuota(quota)
	quota, resp, err := req.Execute()
	return quota, resp, err
}

func (c *Client) UpdateKafkaQuota(quota kafkaquotasv1.KafkaQuotasV1ClientQuotaUpdate) (kafkaquotasv1.KafkaQuotasV1ClientQuota, *http.Response, error) {
	req := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.UpdateKafkaQuotasV1ClientQuota(c.quotaContext(), *quota.Id)
	req = req.KafkaQuotasV1ClientQuotaUpdate(quota)
	updatedQuota, resp, err := req.Execute()
	return updatedQuota, resp, err
}

func (c *Client) DescribeKafkaQuota(quotaId string) (kafkaquotasv1.KafkaQuotasV1ClientQuota, *http.Response, error) {
	req := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.GetKafkaQuotasV1ClientQuota(c.quotaContext(), quotaId)
	quota, resp, err := req.Execute()
	return quota, resp, err
}

func (c *Client) DeleteKafkaQuota(quotaId string) (*http.Response, error) {
	req := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.DeleteKafkaQuotasV1ClientQuota(c.quotaContext(), quotaId)
	resp, err := req.Execute()
	return resp, err
}

func extractKafkaQuotasNextPagePageToken(nextPageUrlStringNullable kafkaquotasv1.NullableString) (string, bool, error) {
	if nextPageUrlStringNullable.IsSet() {
		nextPageUrlString := *nextPageUrlStringNullable.Get()
		pageToken, err := extractPageToken(nextPageUrlString)
		if err != nil {
			return "", true, nil
		}
		return pageToken, false, err
	} else {
		return "", true, nil
	}
}
