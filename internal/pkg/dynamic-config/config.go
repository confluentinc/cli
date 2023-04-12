package dynamicconfig

import (
	"github.com/confluentinc/cli/internal/pkg/ccstructs"
)

// KafkaCluster creates a schedv1 struct from the Kafka cluster of the current context.
func KafkaCluster(ctx *DynamicContext) (*ccstructs.KafkaCluster, error) {
	environmentId, err := ctx.EnvironmentId()
	if err != nil {
		return nil, err
	}

	config, err := ctx.GetKafkaClusterForCommand()
	if err != nil {
		return nil, err
	}

	cluster := &ccstructs.KafkaCluster{
		AccountId: environmentId,
		Id:        config.ID,
	}

	return cluster, nil
}
