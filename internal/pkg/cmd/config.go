package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/confluentinc/ccloud-sdk-go"
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

type ConfigHelper struct {
	Config *config.Config
	Client *ccloud.Client
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
		kc, err := c.Client.Kafka.Describe(context.Background(), req)
		if err != nil {
			if err != ccloud.ErrNotFound {
				return nil, err
			}
			return nil, &errors.UnspecifiedKafkaClusterError{KafkaClusterID: clusterID}
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

func (c *ConfigHelper) UseAPIKey(apiKey, clusterID string) error {
	cfg, err := c.Config.Context()
	if err != nil {
		return err
	}

	cluster, found := cfg.KafkaClusters[clusterID]
	if !found {
		return fmt.Errorf("unknown kafka cluster: %s", clusterID)
	}

	_, found = cluster.APIKeys[apiKey]
	if !found {
		// check if this is API key exists server-side
		key, err := c.Client.APIKey.Get(context.Background(), &authv1.ApiKey{AccountId: c.Config.Auth.Account.Id, Key: apiKey})
		if err != nil {
			return err
		}
		// check if the key is for the right cluster
		found := false
		for _, c := range key.LogicalClusters {
			if c.Id == clusterID {
				found = true
				break
			}
		}
		// this means the requested api-key belongs to a different cluster
		if !found {
			return fmt.Errorf("Invalid api-key %s for cluster %s", apiKey, clusterID)
		}
		// this means the requested api-key exists, but we just don't have the secret saved locally
		return &errors.UnconfiguredAPISecretError{APIKey: apiKey, ClusterID: clusterID}
	}

	cluster.APIKey = apiKey
	return c.Config.Save()
}
