package ccloudv2

import (
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	quotasv2 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v2"
)

// Client represents a Confluent Cloud Client as defined by ccloud-sdk-v2
type Client struct {
	CmkClient    *cmkv2.APIClient
	IamClient    *iamv2.APIClient
	OrgClient    *orgv2.APIClient
	QuotasClient *quotasv2.APIClient
	AuthToken    string
}

func NewClient(cmkClient *cmkv2.APIClient, iamClient *iamv2.APIClient, orgClient *orgv2.APIClient, quotasClient *quotasv2.APIClient, authToken string) *Client {
	return &Client{CmkClient: cmkClient, IamClient: iamClient, OrgClient: orgClient, QuotasClient: quotasClient, AuthToken: authToken}
}

func NewClientWithConfigs(baseURL, userAgent string, isTest bool, authToken string) *Client {
	cmkClient := newCmkClient(baseURL, userAgent, isTest)
	iamClient := newIamClient(baseURL, userAgent, isTest)
	orgClient := newOrgClient(baseURL, userAgent, isTest)
	quotasClient := newQuotasClient(baseURL, userAgent, isTest)
	return NewClient(cmkClient, iamClient, orgClient, quotasClient, authToken)
}
