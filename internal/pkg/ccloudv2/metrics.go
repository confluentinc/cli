package ccloudv2

import (
	"context"
	"net/http"

	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"

	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newMetricsClient(userAgent string, isTest bool) *metricsv2.APIClient {
	cfg := metricsv2.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = metricsv2.ServerConfigurations{{URL: getMetricsServerUrl(isTest), Description: "Confluent Cloud Metrics"}}
	cfg.UserAgent = userAgent

	return metricsv2.NewAPIClient(cfg)
}

func (c *Client) metricsApiContext() context.Context {
	return context.WithValue(context.Background(), metricsv2.ContextAccessToken, c.JwtToken)
}

func (c *Client) MetricsDatasetQuery(dataset string, query metricsv2.QueryRequest) (*metricsv2.QueryResponse, *http.Response, error) {
	return c.MetricsClient.Version2Api.V2MetricsDatasetQueryPost(c.metricsApiContext(), dataset).QueryRequest(query).Execute()
}
