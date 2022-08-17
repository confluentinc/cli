package ccloudv2

import (
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	metricsv2 "github.com/confluentinc/ccloud-sdk-go-v2/metrics/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	servicequotav1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"
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
	KafkaRestClient        *kafkarestv3.APIClient
	MetricsClient          *metricsv2.APIClient
	OrgClient              *orgv2.APIClient
	ServiceQuotaClient     *servicequotav1.APIClient
}

func NewClient(authToken, baseURL, userAgent string, isTest bool) *Client {
	return &Client{
		AuthToken: authToken,

		ApiKeysClient:          newApiKeysClient(baseURL, userAgent, isTest),
		CdxClient:              newCdxClient(baseURL, userAgent, isTest),
		CliClient:              newCliClient(baseURL, userAgent, isTest),
		CmkClient:              newCmkClient(baseURL, userAgent, isTest),
		ConnectClient:          newConnectClient(baseURL, userAgent, isTest),
		IamClient:              newIamClient(baseURL, userAgent, isTest),
		IdentityProviderClient: newIdentityProviderClient(baseURL, userAgent, isTest),
		KafkaRestClient:        newKafkaRestClient(baseURL, userAgent, isTest),
		MetricsClient:          newMetricsClient(baseURL, userAgent, isTest),
		OrgClient:              newOrgClient(baseURL, userAgent, isTest),
		ServiceQuotaClient:     newServiceQuotaClient(baseURL, userAgent, isTest),
	}
}
