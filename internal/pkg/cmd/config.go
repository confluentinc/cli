package cmd

import (
	"github.com/confluentinc/ccloud-sdk-go"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

type ConfigHelper struct {
	Config *config.Config
	Kafka  ccloud.Kafka
}

// KafkaCluster returns the current kafka cluster context
func (c *ConfigHelper) KafkaCluster(clusterID string) (*kafkav1.KafkaCluster, error) {
	kafka, err := c.KafkaClusterConfig(clusterID)
	if err != nil {
		return nil, err
	}
	return &kafkav1.KafkaCluster{AccountId: c.Config.Auth.Account.Id, Id: kafka.ID, ApiEndpoint: kafka.APIEndpoint}, nil
}

// KafkaClusterConfig returns the overridden or current KafkaClusterConfig
func (c *ConfigHelper) KafkaClusterConfig(clusterID string) (*config.KafkaClusterConfig, error) {
	// TODO BUG: this will result in a NoContext error even when clusterID is passed
	context, err := c.Config.Context()
	if err != nil {
		return nil, err
	}

	if clusterID == "" {
		if context.Kafka == "" {
			return nil, errors.ErrNoKafkaContext
		}
		clusterID = context.Kafka
	}

	cluster, found := context.KafkaClusters[context.Kafka]
	if !found {
		return nil, errors.NewUnknownKafkaContextError(clusterID)
	}
	return cluster, nil
}
