package cmd

import (
	"context"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
)

type KafkaREST struct {
	Client  *kafkarestv3.APIClient
	Context context.Context
}

func NewKafkaREST(client *kafkarestv3.APIClient, context context.Context) *KafkaREST {
	return &KafkaREST{
		Client:  client,
		Context: context,
	}
}
