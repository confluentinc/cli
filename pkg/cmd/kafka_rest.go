package cmd

import (
	"context"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
)

type KafkaREST struct {
	Context     context.Context
	CloudClient *ccloudv2.KafkaRestClient
	Client      *kafkarestv3.APIClient
}

func (k *KafkaREST) GetClusterId() string {
	if k == nil || k.CloudClient == nil {
		return ""
	}
	return k.CloudClient.ClusterId
}
