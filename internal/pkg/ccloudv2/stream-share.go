package ccloudv2

import (
	streamsharev1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/cdx/v1"
	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newStreamShareClient(baseURL, userAgent string, isTest bool) *streamsharev1.APIClient {
	cfg := streamsharev1.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = streamsharev1.ServerConfigurations{{URL: getServerUrl(baseURL, isTest), Description: "Confluent Cloud Stream Sharing"}}
	cfg.UserAgent = userAgent

	return streamsharev1.NewAPIClient(cfg)
}
