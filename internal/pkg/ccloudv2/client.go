package ccloudv2

import (
	cliv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/cli/v1"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
)

// Client represents a Confluent Cloud Client as defined by ccloud-sdk-v2
type Client struct {
	CliClient *cliv1.APIClient
	CmkClient *cmkv2.APIClient
	IamClient *iamv2.APIClient
	OrgClient *orgv2.APIClient
	AuthToken string
}

func NewClient(cliClient *cliv1.APIClient, cmkClient *cmkv2.APIClient, iamClient *iamv2.APIClient, orgClient *orgv2.APIClient, authToken string) *Client {
	return &Client{
		CliClient: cliClient,
		CmkClient: cmkClient,
		IamClient: iamClient,
		OrgClient: orgClient,
		AuthToken: authToken,
	}
}

func NewClientWithConfigs(baseURL, userAgent string, isTest bool, authToken string) *Client {
	cliClient := newCliClient(baseURL, userAgent, isTest)
	cmkClient := newCmkClient(baseURL, userAgent, isTest)
	iamClient := newIamClient(baseURL, userAgent, isTest)
	orgClient := newOrgClient(baseURL, userAgent, isTest)
	return NewClient(cliClient, cmkClient, iamClient, orgClient, authToken)
}
