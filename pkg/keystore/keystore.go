//go:generate mocker --prefix "" --dst ../mock/keystore.go --pkg mock keystore.go KeyStore
package keystore

import (
	"github.com/confluentinc/cli/v3/pkg/config"
	dynamicconfig "github.com/confluentinc/cli/v3/pkg/dynamic-config"
	"github.com/confluentinc/cli/v3/pkg/errors"
)

type KeyStore interface {
	HasAPIKey(key, clusterId string) (bool, error)
	StoreAPIKey(key *config.APIKeyPair, clusterId string) error
	DeleteAPIKey(key string) error
}

type ConfigKeyStore struct {
	Config *dynamicconfig.DynamicConfig
}

func (c *ConfigKeyStore) HasAPIKey(key, clusterId string) (bool, error) {
	ctx := c.Config.Context()
	if ctx == nil {
		return false, new(errors.NotLoggedInError)
	}
	kcc, err := ctx.FindKafkaCluster(clusterId)
	if err != nil {
		return false, err
	}
	_, found := kcc.APIKeys[key]
	return found, nil
}

// StoreAPIKey creates a new API key pair in the local key store for later usage
func (c *ConfigKeyStore) StoreAPIKey(key *config.APIKeyPair, clusterId string) error {
	ctx := c.Config.Context()
	if ctx == nil {
		return new(errors.NotLoggedInError)
	}
	kcc, err := ctx.FindKafkaCluster(clusterId)
	if err != nil {
		return err
	}
	kcc.APIKeys[key.Key] = key
	err = kcc.EncryptAPIKeys()
	if err != nil {
		return err
	}
	return c.Config.Save()
}

func (c *ConfigKeyStore) DeleteAPIKey(key string) error {
	ctx := c.Config.Context()
	if ctx == nil {
		return new(errors.NotLoggedInError)
	}
	ctx.KafkaClusterContext.DeleteApiKey(key)
	return c.Config.Save()
}
