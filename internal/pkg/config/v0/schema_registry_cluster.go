package v0

type SchemaRegistryCluster struct {
	SchemaRegistryEndpoint string      `json:"schema_registry_endpoint" hcl:"schema_registry_endpoint"`
	SrCredentials          *APIKeyPair `json:"schema_registry_credentials" hcl:"schema_registry_credentials"`
}
