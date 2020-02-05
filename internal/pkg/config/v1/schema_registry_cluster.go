package v1

import v0 "github.com/confluentinc/cli/internal/pkg/config/v0"

type SchemaRegistryCluster struct {
	SchemaRegistryEndpoint string         `json:"schema_registry_endpoint" hcl:"schema_registry_endpoint"`
	SrCredentials          *v0.APIKeyPair `json:"schema_registry_credentials" hcl:"schema_registry_credentials"`
}
