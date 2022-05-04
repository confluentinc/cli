package ccloudv2

import (
	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2/cli/v1"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	quotasv2 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v2"
)

// Client represents a Confluent Cloud Client as defined by ccloud-sdk-v2
type Client struct {
	AuthToken    string
	CliClient    *cliv1.APIClient
	CmkClient    *cmkv2.APIClient
	IamClient    *iamv2.APIClient
	OrgClient    *orgv2.APIClient
	QuotasClient *quotasv2.APIClient
}

func NewClient(baseURL, userAgent string, isTest bool, authToken string) *Client {
	return &Client{
		AuthToken:    authToken,
		CliClient:    newCliClient(baseURL, userAgent, isTest),
		CmkClient:    newCmkClient(baseURL, userAgent, isTest),
		IamClient:    newIamClient(baseURL, userAgent, isTest),
		OrgClient:    newOrgClient(baseURL, userAgent, isTest),
		QuotasClient: newQuotasClient(baseURL, userAgent, isTest),
	}
}
