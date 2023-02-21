package v1

import (
	"time"
)

// KafkaClusterConfig represents a connection to a Kafka cluster.
type KafkaClusterConfig struct {
	ID           string                 `json:"id" hcl:"id"`
	Name         string                 `json:"name" hcl:"name"`
	Bootstrap    string                 `json:"bootstrap_servers" hcl:"bootstrap_servers"`
	RestEndpoint string                 `json:"rest_endpoint,omitempty" hcl:"rest_endpoint"`
	APIKeys      map[string]*APIKeyPair `json:"api_keys" hcl:"api_keys"`
	// APIKey is your active API key for this cluster and references a key in the APIKeys map
	APIKey     string    `json:"api_key,omitempty" hcl:"api_key"`
	LastUpdate time.Time `json:"last_update,omitempty" hcl:"last_update"`
}

func (k *KafkaClusterConfig) GetName() string {
	if k == nil {
		return ""
	}
	return k.Name
}
