package ccloudv2

import (
	"context"
	"net/http"

	servicequotav1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func newServiceQuotaClient(url, userAgent string, unsafeTrace bool) *servicequotav1.APIClient {
	cfg := servicequotav1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = NewRetryableHttpClient(unsafeTrace)
	cfg.Servers = servicequotav1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return servicequotav1.NewAPIClient(cfg)
}

func (c *Client) serviceQuotasApiContext() context.Context {
	return context.WithValue(context.Background(), servicequotav1.ContextAccessToken, c.AuthToken)
}

func (c *Client) ListServiceQuotas(quotaScope, kafkaCluster, environment, network, quotaCode string) ([]servicequotav1.ServiceQuotaV1AppliedQuota, error) {
	var list []servicequotav1.ServiceQuotaV1AppliedQuota

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListAppliedQuotas(pageToken, quotaScope, kafkaCluster, environment, network, quotaCode)
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

func (c *Client) executeListAppliedQuotas(pageToken, quotaScope, kafkaCluster, environment, network, quotaCode string) (servicequotav1.ServiceQuotaV1AppliedQuotaList, *http.Response, error) {
	req := c.ServiceQuotaClient.AppliedQuotasServiceQuotaV1Api.ListServiceQuotaV1AppliedQuotas(c.serviceQuotasApiContext()).Scope(quotaScope).KafkaCluster(kafkaCluster).Environment(environment).Network(network).Id(quotaCode)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
