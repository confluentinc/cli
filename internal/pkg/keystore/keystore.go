//go:generate mocker --prefix "" --dst ../mock/keystore.go --pkg mock --selfpkg github.com/confluentinc/cli keystore.go KeyStore
package keystore

import (
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type KeyStore interface {
	HasAPIKey(key string, clusterID, environment string) (bool, error)
	StoreAPIKey(key *authv1.ApiKey, clusterID, environment string) error
	DeleteAPIKey(key string) error
}

type ConfigKeyStore struct {
	Config *config.Config
	Helper *cmd.ConfigHelper
}

func (c *ConfigKeyStore) HasAPIKey(key string, clusterID, environment string) (bool, error) {
	kcc, err := c.Helper.KafkaClusterConfig(clusterID, environment)
	if err != nil {
		return false, err
	}

	_, found := kcc.APIKeys[key]
	return found, nil
}

// StoreAPIKey creates a new API key pair in the local key store for later usage
func (c *ConfigKeyStore) StoreAPIKey(key *authv1.ApiKey, clusterID, environment string) error {
	kcc, err := c.Helper.KafkaClusterConfig(clusterID, environment)
	if err != nil {
		return err
	}
	kcc.APIKeys[key.Key] = &config.APIKeyPair{
		Key:    key.Key,
		Secret: key.Secret,
	}
	return c.Config.Save()
}

func (c *ConfigKeyStore) DeleteAPIKey(key string) error {
	cfg, err := c.Config.Context()
	if err != nil {
		return err
	}
	for _, cluster := range cfg.KafkaClusters {
		for apiKey := range cluster.APIKeys {
			if apiKey == key {
				delete(cluster.APIKeys, apiKey)
			}
		}
	}
	return c.Config.Save()
}
