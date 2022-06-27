package ccloudv2

import (
	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	servicequotav1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"
)

// Client represents a Confluent Cloud Client as defined by ccloud-sdk-v2
type Client struct {
	AuthToken          string
	ApiKeysClient      *apikeysv2.APIClient
	ConnectClient      *connectv1.APIClient
	CmkClient          *cmkv2.APIClient
	IamClient          *iamv2.APIClient
	MdsClient          *mdsv2.APIClient
	OrgClient          *orgv2.APIClient
	ServiceQuotaClient *servicequotav1.APIClient
}

func NewClient(baseURL, userAgent string, isTest bool, authToken string) *Client {
	return &Client{
		AuthToken:          authToken,
		ApiKeysClient:      newApiKeysClient(baseURL, userAgent, isTest),
		ConnectClient:      newConnectClient(baseURL, userAgent, isTest),
		CmkClient:          newCmkClient(baseURL, userAgent, isTest),
		IamClient:          newIamClient(baseURL, userAgent, isTest),
		MdsClient:          newMdsClient(baseURL, userAgent, isTest),
		OrgClient:          newOrgClient(baseURL, userAgent, isTest),
		ServiceQuotaClient: newServiceQuotaClient(baseURL, userAgent, isTest),
	}
}
