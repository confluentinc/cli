package ccloudv2

import (
	"strings"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
)

// ccloud sdk v2 API client
type Client struct {
	IamClient *iamv2.APIClient
	AuthToken string
}

func NewCcloudV2Client(iamClient *iamv2.APIClient, authToken string) *Client {
	return &Client{IamClient: iamClient, AuthToken: authToken}
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
