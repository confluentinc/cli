package ccloudv2

import (
	quotasv2 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v2"
	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newQuotasClient(baseURL, userAgent string, isTest bool) *quotasv2.APIClient {
	quotasServer := getServerUrl(baseURL, isTest)
	cfg := quotasv2.NewConfiguration()
	cfg.Servers = quotasv2.ServerConfigurations{
		{URL: quotasServer, Description: "Confluent Cloud servicequota"},
	}
	cfg.UserAgent = userAgent
	cfg.Debug = plog.CliLogger.GetLevel() >= plog.DEBUG
	return quotasv2.NewAPIClient(cfg)
}
