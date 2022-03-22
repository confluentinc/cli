package ccloudv2

import (
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
)

// ccloud sdk v2 API client
type Client struct {
	CmkClient *cmkv2.APIClient
	OrgClient *orgv2.APIClient
	AuthToken string
}

func NewClient(cmkClient *cmkv2.APIClient, orgClient *orgv2.APIClient, authToken string) *Client {
	return &Client{CmkClient: cmkClient, OrgClient: orgClient, AuthToken: authToken}
}
