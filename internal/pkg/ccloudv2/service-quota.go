package ccloudv2

import (
	servicequotav2 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v2"

	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newServiceQuotaClient(baseURL, userAgent string, isTest bool) *servicequotav2.APIClient {
	cfg := servicequotav2.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = servicequotav2.ServerConfigurations{{URL: getServerUrl(baseURL, isTest), Description: "Confluent Cloud Service Quota"}}
	cfg.UserAgent = userAgent

	return servicequotav2.NewAPIClient(cfg)
}
