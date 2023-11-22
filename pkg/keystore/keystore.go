package keystore

import (
	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/config"
	dynamicconfig "github.com/confluentinc/cli/v3/pkg/dynamic-config"
)

type ConfigKeyStore struct {
	Config *config.Config
}

func (c *ConfigKeyStore) HasAPIKey(client *ccloudv2.Client, key, clusterId string) (bool, error) {
	kcc, err := dynamicconfig.FindKafkaCluster(client, c.Config.Context(), clusterId)
	if err != nil {
		return false, err
	}

	_, found := kcc.APIKeys[key]
	return found, nil
}

// StoreAPIKey creates a new API key pair in the local key store for later usage
func (c *ConfigKeyStore) StoreAPIKey(client *ccloudv2.Client, key *config.APIKeyPair, clusterId string) error {
	kcc, err := dynamicconfig.FindKafkaCluster(client, c.Config.Context(), clusterId)
	if err != nil {
		return err
	}

	kcc.APIKeys[key.Key] = key
	if err := kcc.EncryptAPIKeys(); err != nil {
		return err
	}

	return c.Config.Save()
}

func (c *ConfigKeyStore) DeleteAPIKey(key string) error {
	c.Config.Context().KafkaClusterContext.DeleteApiKey(key)
	return c.Config.Save()
}
