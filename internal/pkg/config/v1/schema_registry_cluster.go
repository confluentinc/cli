package v1

type SchemaRegistryCluster struct {
	Id                     string      `json:"id"`
	SchemaRegistryEndpoint string      `json:"schema_registry_endpoint"`
	SrCredentials          *APIKeyPair `json:"schema_registry_credentials"`
}
