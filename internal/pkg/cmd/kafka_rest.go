package cmd

import (
	"context"
	kafkarest_cc "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	kafkarest_cp "github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
)

type CloudKafkaREST struct {
	Client  *kafkarest_cc.APIClient
	Context context.Context
}

type CPKafkaREST struct {
	Client *kafkarest_cp.APIClient
	Context context.Context
}

func NewCloudKafkaREST(client *kafkarest_cc.APIClient, context context.Context) *CloudKafkaREST {
	return &CloudKafkaREST{
		Client:  client,
		Context: context,
	}
}

func NewCPKafkaRest(client *kafkarest_cp.APIClient, context context.Context) *CPKafkaREST {
	return &CPKafkaREST{
		Client:  client,
		Context: context,
	}
}

func GetCloudKafkaRestBaseUrl(client *kafkarest_cc.APIClient) string {
	return client.GetConfig().Servers[0].URL
}
