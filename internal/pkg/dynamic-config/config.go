package dynamicconfig

import (
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
)

// KafkaCluster creates a schedv1 struct from the Kafka cluster of the current context.
func KafkaCluster(ctx *DynamicContext) (*schedv1.KafkaCluster, error) {
	environmentId, err := ctx.AuthenticatedEnvId()
	if err != nil {
		return nil, err
	}

	config, err := ctx.GetKafkaClusterForCommand()
	if err != nil {
		return nil, err
	}

	cluster := &schedv1.KafkaCluster{
		AccountId:   environmentId,
		Id:          config.ID,
		ApiEndpoint: config.APIEndpoint,
	}

	return cluster, nil
}
