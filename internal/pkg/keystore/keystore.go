//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst ../mock/keystore.go --pkg mock keystore.go KeyStore
package keystore

import (
	authv1 "github.com/confluentinc/ccloudapis/auth/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

type KeyStore interface {
	HasAPIKey(key string, clusterId string, cmd *cobra.Command) (bool, error)
	StoreAPIKey(key *authv1.ApiKey, clusterId string, cmd *cobra.Command) error
	DeleteAPIKey(key string, cmd *cobra.Command) error
}

type ConfigKeyStore struct {
	Config *cmd.DynamicConfig
}

func (c *ConfigKeyStore) HasAPIKey(key string, clusterId string, cmd *cobra.Command) (bool, error) {
	ctx, err := c.Config.Context(cmd)
	if err != nil {
		return false, err
	}
	if ctx == nil {
		return false, errors.ErrNoContext
	}
	kcc, err := ctx.FindKafkaCluster(cmd, clusterId)
	if err != nil {
		return false, err
	}
	_, found := kcc.APIKeys[key]
	return found, nil
}

// StoreAPIKey creates a new API key pair in the local key store for later usage
func (c *ConfigKeyStore) StoreAPIKey(key *authv1.ApiKey, clusterId string, cmd *cobra.Command) error {
	ctx, err := c.Config.Context(cmd)
	if err != nil {
		return err
	}
	if ctx == nil {
		return errors.ErrNoContext
	}
	kcc, err := ctx.FindKafkaCluster(cmd, clusterId)
	if err != nil {
		return err
	}
	kcc.APIKeys[key.Key] = &v0.APIKeyPair{
		Key:    key.Key,
		Secret: key.Secret,
	}
	return c.Config.Save()
}

func (c *ConfigKeyStore) DeleteAPIKey(key string, cmd *cobra.Command) error {
	context, err := c.Config.Context(cmd)
	if err != nil {
		return err
	}
	if context == nil {
		return errors.ErrNoContext
	}
	for _, cluster := range context.KafkaClusters {
		for apiKey := range cluster.APIKeys {
			if apiKey == key {
				delete(cluster.APIKeys, apiKey)
			}
		}
	}
	return c.Config.Save()
}
