package v1

import (
	"strings"
	"time"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
)

// KafkaClusterConfig represents a connection to a Kafka cluster.
type KafkaClusterConfig struct {
	ID           string                 `json:"id" hcl:"id"`
	Name         string                 `json:"name" hcl:"name"`
	Bootstrap    string                 `json:"bootstrap_servers" hcl:"bootstrap_servers"`
	APIEndpoint  string                 `json:"api_endpoint,omitempty" hcl:"api_endpoint"`
	RestEndpoint string                 `json:"rest_endpoint,omitempty" hcl:"rest_endpoint"`
	APIKeys      map[string]*APIKeyPair `json:"api_keys" hcl:"api_keys"`
	// APIKey is your active api key for this cluster and references a key in the APIKeys map
	APIKey     string    `json:"api_key,omitempty" hcl:"api_key"`
	LastUpdate time.Time `json:"last_update,omitempty" hcl:"last_update"`
}

func NewKafkaClusterConfig(cluster *schedv1.KafkaCluster) *KafkaClusterConfig {
	return &KafkaClusterConfig{
		ID:           cluster.Id,
		Name:         cluster.Name,
		Bootstrap:    strings.TrimPrefix(cluster.Endpoint, "SASL_SSL://"),
		APIEndpoint:  cluster.ApiEndpoint,
		RestEndpoint: cluster.RestEndpoint,
		APIKeys:      make(map[string]*APIKeyPair),
		LastUpdate:   time.Now(),
	}
}
