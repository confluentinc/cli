package cmd

import (
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
)

// KafkaCluster creates an KafkaV1 struct from the Kafka cluster of the current context.
func KafkaCluster(cmd *AuthenticatedCLICommand) (*kafkav1.KafkaCluster, error) {
	kcc, err := cmd.Context.ActiveKafkaCluster(cmd.Command)
	if err != nil {
		return nil, err
	}
	return &kafkav1.KafkaCluster{AccountId: cmd.EnvironmentId(), Id: kcc.ID, ApiEndpoint: kcc.APIEndpoint}, nil
}
