package ccloudv2

import (
	servicequotav1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"
)

func newServiceQuotaClient(url, userAgent string, unsafeTrace bool) *servicequotav1.APIClient {
	cfg := servicequotav1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = newRetryableHttpClient(unsafeTrace)
	cfg.Servers = servicequotav1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return servicequotav1.NewAPIClient(cfg)
}
