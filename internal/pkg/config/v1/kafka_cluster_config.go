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
