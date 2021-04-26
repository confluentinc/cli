package kafka

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
)

func ListKafkaClusters(client *ccloud.Client, environmentId string) ([]*schedv1.KafkaCluster, error) {
	req := &schedv1.KafkaCluster{AccountId: environmentId}
	return client.Kafka.List(context.Background(), req)
}
