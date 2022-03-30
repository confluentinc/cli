package cmd

import (
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
)

// KafkaCluster creates an schedv1 struct from the Kafka cluster of the current context.
func KafkaCluster(ctx *DynamicContext) (*schedv1.KafkaCluster, error) {
	kcc, err := ctx.GetKafkaClusterForCommand()
	if err != nil {
		return nil, err
	}
	envId, err := ctx.AuthenticatedEnvId()
	if err != nil {
		return nil, err
	}
	return &schedv1.KafkaCluster{AccountId: envId, Id: kcc.ID, ApiEndpoint: kcc.APIEndpoint}, nil
}
