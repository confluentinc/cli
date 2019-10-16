package config

import (
	"fmt"
)

// APIKeyPair holds an API Key and Secret.
type APIKeyPair struct {
	Key    string `json:"api_key" hcl:"api_key"`
	Secret string `json:"api_secret" hcl:"api_secret"`
}

// KafkaClusterConfig represents a connection to a Kafka cluster.
type KafkaClusterConfig struct {
	ID          string                 `json:"id" hcl:"id"`
	Name        string                 `json:"name" hcl:"name"`
	Bootstrap   string                 `json:"bootstrap_servers" hcl:"bootstrap_servers"`
	APIEndpoint string                 `json:"api_endpoint,omitempty" hcl:"api_endpoint"`
	APIKeys     map[string]*APIKeyPair `json:"api_keys" hcl:"api_keys"`
	// APIKey is your active api key for this cluster and references a key in the APIKeys map
	APIKey string `json:"api_key,omitempty" hcl:"api_key"`
}

type SchemaRegistryCluster struct {
	SchemaRegistryEndpoint string      `json:"schema_registry_endpoint" hcl:"schema_registry_endpoint"`
	SrCredentials          *APIKeyPair `json:"schema_registry_credentials" hcl:"schema_registry_credentials"`
}

// Context represents a specific CLI context.
type Context struct {
	Name       string
	Platform   string `json:"platform" hcl:"platform"`
	Credential string `json:"credentials" hcl:"credentials"`
	// KafkaClusters store connection info for interacting directly with Kafka (e.g., consume/produce, etc)
	// N.B. These may later be exposed in the CLI to directly register kafkas (outside a Control Plane)
	KafkaClusters map[string]*KafkaClusterConfig `json:"kafka_clusters" hcl:"kafka_clusters"`
	// Kafka is your active Kafka cluster and references a key in the KafkaClusters map
	Kafka string `json:"kafka_cluster" hcl:"kafka_cluster"`
	// SR map keyed by environment-id
	SchemaRegistryClusters map[string]*SchemaRegistryCluster `json:"schema_registry_cluster" hcl:"schema_registry_cluster"`
}

func (c *Context) SetActiveCluster(clusterId string) error {
	if _, ok := c.KafkaClusters[clusterId]; !ok {
		return fmt.Errorf("cluster \"%s\" does not exist in context \"%s\"", clusterId, c.Name)
	}
	c.Kafka = clusterId
	return nil
}
