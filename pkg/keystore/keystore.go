package keystore

import (
	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/kafka"
)

type ConfigKeyStore struct {
	Config *config.Config
}

func (c *ConfigKeyStore) HasAPIKey(client *ccloudv2.Client, key, clusterId string) (bool, error) {
	kcc, err := kafka.FindCluster(client, c.Config.Context(), clusterId)
	if err != nil {
		return false, err
	}

	_, found := kcc.APIKeys[key]
	return found, nil
}

// StoreAPIKey creates a new API key pair in the local key store for later usage
func (c *ConfigKeyStore) StoreAPIKey(client *ccloudv2.Client, key *config.APIKeyPair, clusterId string) error {
	kcc, err := kafka.FindCluster(client, c.Config.Context(), clusterId)
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
	ctx := c.Config.Context()
	ctx.KafkaClusterContext.DeleteApiKey(key)
	ctx.DeleteGlobalAPIKey(key)
	return c.Config.Save()
}

// HasGlobalAPIKey reports whether a Global API key with the given id is stored locally.
func (c *ConfigKeyStore) HasGlobalAPIKey(key string) bool {
	return c.Config.Context().HasGlobalAPIKey(key)
}

// StoreGlobalAPIKey persists a Global API key pair on the active context. The secret is encrypted
// at rest, matching the behavior of StoreAPIKey for cluster-scoped keys.
func (c *ConfigKeyStore) StoreGlobalAPIKey(pair *config.APIKeyPair) error {
	if err := c.Config.Context().StoreGlobalAPIKey(pair); err != nil {
		return err
	}
	return c.Config.Save()
}
