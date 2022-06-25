package ccloudv2

import (
	servicequotav1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"

	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newServiceQuotaClient(baseURL, userAgent string, isTest bool) *servicequotav1.APIClient {
	cfg := servicequotav1.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = servicequotav1.ServerConfigurations{{URL: getServerUrl(baseURL, isTest), Description: "Confluent Cloud Service Quota"}}
	cfg.UserAgent = userAgent

	return servicequotav1.NewAPIClient(cfg)
}
