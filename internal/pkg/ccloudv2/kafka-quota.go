package ccloudv2

import (
	kafkaquotas "github.com/confluentinc/ccloud-sdk-go-v2-internal/kafka-quotas/v1"
	plog "github.com/confluentinc/cli/internal/pkg/log"
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
