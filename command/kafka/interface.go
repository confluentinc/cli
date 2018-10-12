package kafka

import (
	"context"

	plugin "github.com/hashicorp/go-plugin"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
)

type Kafka interface {
	List(ctx context.Context, cluster *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, error)
	Describe(ctx context.Context, cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, error)
	Create(ctx context.Context, config *schedv1.KafkaClusterConfig) (*schedv1.KafkaCluster, error)
	Delete(ctx context.Context, cluster *schedv1.KafkaCluster) error
}

type Plugin struct {
	plugin.NetRPCUnsupportedPlugin

	Impl Kafka
}
