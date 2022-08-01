package ccloudv2

import (
	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/cdx/v1"
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"
	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	servicequotav1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"
)

// Client represents a Confluent Cloud Client as defined by ccloud-sdk-v2
type Client struct {
	AuthToken              string
	JwtToken               string
	ApiKeysClient          *apikeysv2.APIClient
	CliClient              *cliv1.APIClient
	CmkClient              *cmkv2.APIClient
	ConnectClient          *connectv1.APIClient
	IamClient              *iamv2.APIClient
	IdentityProviderClient *identityproviderv2.APIClient
	MetricsClient          *metricsv2.APIClient
	OrgClient              *orgv2.APIClient
	ServiceQuotaClient     *servicequotav1.APIClient
	StreamShareClient      *cdxv1.APIClient
}

func NewClient(baseURL, userAgent string, isTest bool, authToken string) *Client {
	return &Client{
		AuthToken:              authToken,
		ApiKeysClient:          newApiKeysClient(baseURL, userAgent, isTest),
		CliClient:              newCliClient(baseURL, userAgent, isTest),
		CmkClient:              newCmkClient(baseURL, userAgent, isTest),
		ConnectClient:          newConnectClient(baseURL, userAgent, isTest),
		IamClient:              newIamClient(baseURL, userAgent, isTest),
		IdentityProviderClient: newIdentityProviderClient(baseURL, userAgent, isTest),
		MetricsClient:          newMetricsClient(userAgent, isTest),
		OrgClient:              newOrgClient(baseURL, userAgent, isTest),
		ServiceQuotaClient:     newServiceQuotaClient(baseURL, userAgent, isTest),
		StreamShareClient:      newCdxClient(baseURL, userAgent, isTest),
	}
}
