package cmd

import (
	"context"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
)

type KafkaREST struct {
	Context     context.Context
	CloudClient *ccloudv2.Client
	Client      *kafkarestv3.APIClient // TODO: Rename to PlatformClient
}

func NewKafkaREST(ctx context.Context, cloudClient *ccloudv2.Client, client *kafkarestv3.APIClient) *KafkaREST {
	return &KafkaREST{
		Context:     ctx,
		CloudClient: cloudClient,
		Client:      client,
	}
}
