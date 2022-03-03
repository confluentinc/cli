package ccloudv2

import (
	"strings"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
)

// ccloud sdk v2 API client
type Client struct {
	CmkClient *cmkv2.APIClient
	OrgClient *orgv2.APIClient
	AuthToken string
}

func NewCcloudV2Client(cmkClient *cmkv2.APIClient, orgClient *orgv2.APIClient) *Client {
	return &Client{CmkClient: cmkClient, OrgClient: orgClient}
}

func (c *Client) SetAuthToken(authToken string) {
	c.AuthToken = authToken
}

func getV2ServerUrl(baseURL string, isTest bool) string {
	if isTest {
		return "http://127.0.0.1:2048"
	}
	if strings.Contains(baseURL, "devel") {
		return "https://api.devel.cpdev.cloud"
	} else if strings.Contains(baseURL, "stag") {
		return "https://api.stag.cpdev.cloud"
	}
	return "https://api.confluent.cloud"
}
