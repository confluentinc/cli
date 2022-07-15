package ccloudv2

import (
	"context"
	kafkaquotas "github.com/confluentinc/ccloud-sdk-go-v2-internal/kafka-quotas/v1"
	plog "github.com/confluentinc/cli/internal/pkg/log"
	"net/http"
)

func newKafkaQuotasClient(baseURL, userAgent string, isTest bool) *kafkaquotas.APIClient {
	quotasServer := getServerUrl(baseURL, isTest)
	cfg := kafkaquotas.NewConfiguration()
	cfg.Servers = kafkaquotas.ServerConfigurations{
		{URL: quotasServer, Description: "Confluent Cloud servicequota"},
	}
	cfg.UserAgent = userAgent
	cfg.Debug = plog.CliLogger.GetLevel() >= plog.DEBUG
	return kafkaquotas.NewAPIClient(cfg)
}

func (c *Client) quotaContext() context.Context {
	return context.WithValue(context.Background(), kafkaquotas.ContextAccessToken, c.AuthToken)
}

func (c *Client) ListKafkaQuotas(clusterId, envId string) ([]kafkaquotas.KafkaQuotasV1ClientQuota, error) {
	quotas := make([]kafkaquotas.KafkaQuotasV1ClientQuota, 0)

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

func (c *Client) listQuotas(clusterId, envId, pageToken string) (kafkaquotas.KafkaQuotasV1ClientQuotaList, *http.Response, error) {
	var req kafkaquotas.ApiListKafkaQuotasV1ClientQuotasRequest
	if pageToken != "" {
		req = c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.ListKafkaQuotasV1ClientQuotas(c.quotaContext()).PageSize(ccloudV2ListPageSize).PageToken(pageToken)
	} else {
		req = c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.ListKafkaQuotasV1ClientQuotas(c.quotaContext()).PageSize(ccloudV2ListPageSize)
	}
	req = req.Cluster(clusterId).Environment(envId)
	return c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.ListKafkaQuotasV1ClientQuotasExecute(req)
}

func (c *Client) CreateKafkaQuota(displayName string, description string, throughput *kafkaquotas.KafkaQuotasV1Throughput,
	cluster *kafkaquotas.ObjectReference, principals *[]kafkaquotas.ObjectReference,
	environment *kafkaquotas.ObjectReference) (kafkaquotas.KafkaQuotasV1ClientQuota, error) {
	req := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.CreateKafkaQuotasV1ClientQuota(c.quotaContext())
	req = req.KafkaQuotasV1ClientQuota(kafkaquotas.KafkaQuotasV1ClientQuota{
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

func (c *Client) UpdateKafkaQuota(quotaId string, displayName string, description string, throughput *kafkaquotas.KafkaQuotasV1Throughput,
	cluster *kafkaquotas.ObjectReference, principals *[]kafkaquotas.ObjectReference,
	environment *kafkaquotas.ObjectReference) (kafkaquotas.KafkaQuotasV1ClientQuota, error) {
	req := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.UpdateKafkaQuotasV1ClientQuota(c.quotaContext(), quotaId)
	req = req.KafkaQuotasV1ClientQuota(kafkaquotas.KafkaQuotasV1ClientQuota{
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

func (c *Client) DescribeKafkaQuota(quotaId string) (kafkaquotas.KafkaQuotasV1ClientQuota, error) {
	req := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.GetKafkaQuotasV1ClientQuota(c.quotaContext(), quotaId)
	quota, _, err := req.Execute()
	return quota, err
}

func (c *Client) DeleteKafkaQuota(quotaId string) error {
	req := c.KafkaQuotasClient.ClientQuotasKafkaQuotasV1Api.DeleteKafkaQuotasV1ClientQuota(c.quotaContext(), quotaId)
	_, err := req.Execute()
	return err
}
