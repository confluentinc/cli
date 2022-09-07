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
	cfg.Servers = kafkaquotasv1.ServerConfigurations{
		{URL: url},
	}
	cfg.UserAgent = userAgent

	return kafkaquotasv1.NewAPIClient(cfg)
}

func (c *Client) quotaContext() context.Context {
	return context.WithValue(context.Background(), kafkaquotasv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) ListKafkaQuotas(clusterId, envId string) ([]kafkaquotasv1.KafkaQuotasV1ClientQuota, error) {
	quotas := make([]kafkaquotasv1.KafkaQuotasV1ClientQuota, 0)

	collectedAllQuotas := false
	pageToken := ""
	for !collectedAllQuotas {
		quotaList, _, err := c.listQuotas(clusterId, envId, pageToken)
		if err != nil {
			return nil, err
		}
		quotas = append(quotas, quotaList.GetData()...)

		// nextPageUrlStringNullable is nil for the last page
		nextPageUrlStringNullable := quotaList.GetMetadata().Next
		pageToken, collectedAllQuotas, err = extractKafkaQuotasNextPagePageToken(nextPageUrlStringNullable)
		if err != nil {
			return nil, err
		}
	}
	return quotas, nil
}

func (c *Client) listQuotas(clusterId, envId, pageToken string) (kafkaquotasv1.KafkaQuotasV1ClientQuotaList, *http.Response, error) {
	var req kafkaquotasv1.ApiListKafkaQuotasV1ClientQuotasRequest
	if pageToken != "" {
		req = c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.ListKafkaQuotasV1ClientQuotas(c.quotaContext()).PageSize(ccloudV2ListPageSize).PageToken(pageToken)
	} else {
		req = c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.ListKafkaQuotasV1ClientQuotas(c.quotaContext()).PageSize(ccloudV2ListPageSize)
	}
	req = req.Cluster(clusterId).Environment(envId)
	return c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.ListKafkaQuotasV1ClientQuotasExecute(req)
}

func (c *Client) CreateKafkaQuota(displayName string, description string, throughput *kafkaquotasv1.KafkaQuotasV1Throughput,
	cluster *kafkaquotasv1.ObjectReference, principals *[]kafkaquotasv1.ObjectReference,
	environment *kafkaquotasv1.ObjectReference) (kafkaquotasv1.KafkaQuotasV1ClientQuota, error) {
	req := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.CreateKafkaQuotasV1ClientQuota(c.quotaContext())
	req = req.KafkaQuotasV1ClientQuota(kafkaquotasv1.KafkaQuotasV1ClientQuota{
		DisplayName: &displayName,
		Description: &description,
		Throughput:  throughput,
		Cluster:     cluster,
		Principals:  principals,
		Environment: environment,
	})
	quota, _, err := req.Execute()
	return quota, err
}

func (c *Client) UpdateKafkaQuota(quota kafkaquotasv1.KafkaQuotasV1ClientQuotaUpdate) (kafkaquotasv1.KafkaQuotasV1ClientQuota, error) {
	req := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.UpdateKafkaQuotasV1ClientQuota(c.quotaContext(), *quota.Id)
	req = req.KafkaQuotasV1ClientQuotaUpdate(quota)
	updatedQuota, _, err := req.Execute()
	return updatedQuota, err
}

func (c *Client) DescribeKafkaQuota(quotaId string) (kafkaquotasv1.KafkaQuotasV1ClientQuota, error) {
	req := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.GetKafkaQuotasV1ClientQuota(c.quotaContext(), quotaId)
	quota, _, err := req.Execute()
	return quota, err
}

func (c *Client) DeleteKafkaQuota(quotaId string) error {
	req := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.DeleteKafkaQuotasV1ClientQuota(c.quotaContext(), quotaId)
	_, err := req.Execute()
	return err
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
