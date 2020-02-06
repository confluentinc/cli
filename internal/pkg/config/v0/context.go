package v0

// Context represents a specific CLI context.
type Context struct {
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
