package ccloudv2

import (
	quotasv1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"

	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newServiceQuotaClient(baseURL, userAgent string, isTest bool) *quotasv1.APIClient {
	cfg := quotasv1.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = quotasv1.ServerConfigurations{{URL: getServerUrl(baseURL, isTest), Description: "Confluent Cloud Service Quota"}}
	cfg.UserAgent = userAgent

	return quotasv1.NewAPIClient(cfg)
}
