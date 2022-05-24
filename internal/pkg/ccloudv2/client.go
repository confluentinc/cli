package ccloudv2

import (
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	servicequotav2 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v2"
)

// Client represents a Confluent Cloud Client as defined by ccloud-sdk-v2
type Client struct {
	AuthToken          string
	CmkClient          *cmkv2.APIClient
	IamClient          *iamv2.APIClient
	OrgClient          *orgv2.APIClient
	ServiceQuotaClient *servicequotav2.APIClient
}

func NewClient(baseURL, userAgent string, isTest bool, authToken string) *Client {
	return &Client{
		AuthToken:          authToken,
		CmkClient:          newCmkClient(baseURL, userAgent, isTest),
		IamClient:          newIamClient(baseURL, userAgent, isTest),
		OrgClient:          newOrgClient(baseURL, userAgent, isTest),
		ServiceQuotaClient: newServiceQuotaClient(baseURL, userAgent, isTest),
	}
}
