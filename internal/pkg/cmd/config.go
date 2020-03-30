package cmd

import (
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/spf13/cobra"
)

// KafkaCluster creates an KafkaV1 struct from the Kafka cluster of the current context.
func KafkaCluster(cmd *cobra.Command, ctx *DynamicContext) (*kafkav1.KafkaCluster, error) {
	kcc, err := ctx.ActiveKafkaCluster(cmd)
	if err != nil {
		return nil, err
	}
	envId, err := ctx.AuthenticatedEnvId(cmd)
	if err != nil {
		return nil, err
	}
	return &kafkav1.KafkaCluster{AccountId: envId, Id: kcc.ID, ApiEndpoint: kcc.APIEndpoint}, nil
}
