package cmd

import (
	"context"
	"strings"

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
func (c *ConfigHelper) KafkaCluster(clusterID, environment string) (*kafkav1.KafkaCluster, error) {
	kafka, err := c.KafkaClusterConfig(clusterID, environment)
	if err != nil {
		return nil, err
	}
	return &kafkav1.KafkaCluster{AccountId: c.Config.Auth.Account.Id, Id: kafka.ID, ApiEndpoint: kafka.APIEndpoint}, nil
}

// KafkaClusterConfig returns the overridden or current KafkaClusterConfig
func (c *ConfigHelper) KafkaClusterConfig(clusterID, environment string) (*config.KafkaClusterConfig, error) {
	ctx, err := c.Config.Context()
	if err != nil {
		return nil, err
	}

	if clusterID == "" {
		if ctx.Kafka == "" {
			return nil, errors.ErrNoKafkaContext
		}
		clusterID = ctx.Kafka
	}

	if ctx.KafkaClusters == nil {
		ctx.KafkaClusters = map[string]*config.KafkaClusterConfig{}
	}
	cluster, found := ctx.KafkaClusters[clusterID]
	if !found {
		// Let's fetch the cluster details
		req := &kafkav1.KafkaCluster{AccountId: environment, Id: clusterID}
		kc, err := c.Kafka.Describe(context.Background(), req)
		if err != nil {
			if err != ccloud.ErrNotFound {
				return nil, err
			}
			return nil, errors.NewUnknownKafkaContextError(clusterID)
		}
		cluster = &config.KafkaClusterConfig{
			ID:          clusterID,
			Bootstrap:   strings.TrimPrefix(kc.Endpoint, "SASL_SSL://"),
			APIEndpoint: kc.ApiEndpoint,
			APIKeys:     make(map[string]*config.APIKeyPair),
		}

		// Then save it locally for reuse
		ctx.KafkaClusters[clusterID] = cluster
		err = c.Config.Save()
		if err != nil {
			return nil, err
		}
	}
	return cluster, nil
}
