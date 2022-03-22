package ccloudv2

import (
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
)

// Client represents a Confluent Cloud Client as defined by ccloud-sdk-v2
type Client struct {
	IamClient *iamv2.APIClient
	AuthToken string
}

func NewClient(iamClient *iamv2.APIClient, authToken string) *Client {
	return &Client{IamClient: iamClient, AuthToken: authToken}
}
