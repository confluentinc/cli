package ccloudv2

import (
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
)

// Client represents a Confluent Cloud Client as defined by ccloud-sdk-v2
type Client struct {
	IamClient *iamv2.APIClient
	KafkaRESTProvider *CloudKafkaRESTProvider
	AuthToken string
}

func NewClient(baseUrl string, isTest bool, authToken string) *Client {
	client := &Client{
		IamClient:         newIamClient(baseUrl, isTest),
		AuthToken:         authToken,
	}
	return client
}
