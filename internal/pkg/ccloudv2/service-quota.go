package ccloudv2

import (
	"context"
	"net/http"

	servicequotav1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"

	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newServiceQuotaClient(baseURL, userAgent string, isTest bool) *servicequotav1.APIClient {
	cfg := servicequotav1.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = servicequotav1.ServerConfigurations{{URL: getServerUrl(baseURL, isTest)}}
	cfg.UserAgent = userAgent

	return servicequotav1.NewAPIClient(cfg)
}

func (c *Client) serviceQuotasApiContext() context.Context {
	return context.WithValue(context.Background(), servicequotav1.ContextAccessToken, c.AuthToken)
}

func (c *Client) ListServiceQuotas(quotaScope, kafkaCluster, environment, network string) ([]servicequotav1.ServiceQuotaV1AppliedQuota, error) {
	var list []servicequotav1.ServiceQuotaV1AppliedQuota

	done := false
	pageToken := ""
	for !done {
		page, _, err := c.executeListAppliedQuotas(pageToken, quotaScope, kafkaCluster, environment, network)
		if err != nil {
			return nil, err
		}
		list = append(list, page.GetData()...)

		pageToken, done, err = extractServiceQuotaNextPagePageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *Client) executeListAppliedQuotas(pageToken, quotaScope, kafkaCluster, environment, network string) (servicequotav1.ServiceQuotaV1AppliedQuotaList, *http.Response, error) {
	req := c.ServiceQuotaClient.AppliedQuotasServiceQuotaV1Api.ListServiceQuotaV1AppliedQuotas(c.serviceQuotasApiContext()).Scope(quotaScope).KafkaCluster(kafkaCluster).Environment(environment).Network(network)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return c.ServiceQuotaClient.AppliedQuotasServiceQuotaV1Api.ListServiceQuotaV1AppliedQuotasExecute(req)
}

func extractServiceQuotaNextPagePageToken(nextPageUrlStringNullable servicequotav1.NullableString) (string, bool, error) {
	if !nextPageUrlStringNullable.IsSet() {
		return "", true, nil
	}
	nextPageUrlString := *nextPageUrlStringNullable.Get()
	pageToken, err := extractPageToken(nextPageUrlString)
	return pageToken, false, err
}
