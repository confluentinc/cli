package ccloudv2

import (
	"net/http"

	piv2 "github.com/confluentinc/ccloud-sdk-go-v2/provider-integration/v2"
)

func newProviderIntegrationV2Client(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *piv2.APIClient {
	configuration := piv2.NewConfiguration()
	configuration.HTTPClient = httpClient
	configuration.UserAgent = userAgent
	configuration.Servers = []piv2.ServerConfiguration{
		{
			URL: url,
		},
	}
	configuration.Debug = unsafeTrace
	return piv2.NewAPIClient(configuration)
}