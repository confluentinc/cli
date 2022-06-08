package ccloudv2

import (
	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/identity-provider/v2"
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	servicequotav2 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v2"
)

// Client represents a Confluent Cloud Client as defined by ccloud-sdk-v2
type Client struct {
	AuthToken              string
	ApiKeysClient          *apikeysv2.APIClient
	CmkClient              *cmkv2.APIClient
	IamClient              *iamv2.APIClient
	IdentityProviderClient *identityproviderv2.APIClient
	IdentityPoolClient     *identityproviderv2.APIClient
	OrgClient              *orgv2.APIClient
	ServiceQuotaClient     *servicequotav2.APIClient
}

func NewClient(baseURL, userAgent string, isTest bool, authToken string) *Client {
	return &Client{
		AuthToken:              authToken,
		ApiKeysClient:          newApiKeysClient(baseURL, userAgent, isTest),
		CmkClient:              newCmkClient(baseURL, userAgent, isTest),
		IamClient:              newIamClient(baseURL, userAgent, isTest),
		IdentityProviderClient: newIdentityProviderClient(baseURL, userAgent, isTest),
		IdentityPoolClient:     newIdentityPoolClient(baseURL, userAgent, isTest),
		OrgClient:              newOrgClient(baseURL, userAgent, isTest),
		ServiceQuotaClient:     newServiceQuotaClient(baseURL, userAgent, isTest),
	}
}
