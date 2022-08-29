package ccloudv2

import (
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"
	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	servicequotav1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"

	testserver "github.com/confluentinc/cli/test/test-server"
)

// Client represents a Confluent Cloud Client as defined by ccloud-sdk-go-v2
type Client struct {
	AuthToken string
	JwtToken  string

	ApiKeysClient          *apikeysv2.APIClient
	CdxClient              *cdxv1.APIClient
	CliClient              *cliv1.APIClient
	CmkClient              *cmkv2.APIClient
	ConnectClient          *connectv1.APIClient
	IamClient              *iamv2.APIClient
	IdentityProviderClient *identityproviderv2.APIClient
	MetricsClient          *metricsv2.APIClient
	OrgClient              *orgv2.APIClient
	ServiceQuotaClient     *servicequotav1.APIClient
}

func NewClient(baseUrl string, isTest bool, authToken, userAgent string, unsafeTrace bool) *Client {
	url := getServerUrl(baseUrl)
	if isTest {
		url = testserver.TestV2CloudURL.String()
	}

	return &Client{
		AuthToken: authToken,

		ApiKeysClient:          newApiKeysClient(url, userAgent, unsafeTrace),
		CdxClient:              newCdxClient(url, userAgent, unsafeTrace),
		CliClient:              newCliClient(url, userAgent, unsafeTrace),
		CmkClient:              newCmkClient(url, userAgent, unsafeTrace),
		ConnectClient:          newConnectClient(url, userAgent, unsafeTrace),
		IamClient:              newIamClient(url, userAgent, unsafeTrace),
		IdentityProviderClient: newIdentityProviderClient(url, userAgent, unsafeTrace),
		MetricsClient:          newMetricsClient(baseUrl, userAgent, unsafeTrace, isTest),
		OrgClient:              newOrgClient(url, userAgent, unsafeTrace),
		ServiceQuotaClient:     newServiceQuotaClient(url, userAgent, unsafeTrace),
	}
}
