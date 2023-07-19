package v1

import (
	"time"
)

// KafkaClusterConfig represents a connection to a Kafka cluster.
type KafkaClusterConfig struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Bootstrap    string                 `json:"bootstrap_servers"`
	RestEndpoint string                 `json:"rest_endpoint,omitempty"`
	APIKeys      map[string]*APIKeyPair `json:"api_keys"`
	APIKey       string                 `json:"api_key,omitempty"`
	LastUpdate   time.Time              `json:"last_update,omitempty"`
}

func (k *KafkaClusterConfig) GetName() string {
	if k == nil {
		return ""
	}
	return k.Name
}

func (k *KafkaClusterConfig) DecryptAPIKeys() error {
	for _, key := range k.APIKeys {
		err := key.DecryptSecret()
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *KafkaClusterConfig) EncryptAPIKeys() error {
	for _, key := range k.APIKeys {
		err := key.EncryptSecret()
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *KafkaClusterConfig) GetApiSecret() string {
	if k.APIKeys != nil {
		if apiKey, ok := k.APIKeys[k.APIKey]; ok {
			return apiKey.Secret
		}
	}
	return ""
}
