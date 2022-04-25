package v1

type StreamGovernanceV1Cluster struct {
	APIVersion             string `json:"api_version" hcl:"api_version"`
	Kind                   string `json:"kind" hcl:"kind"`
	Id                     string `json:"id" hcl:"id"`
	ResourceName           string `json:"resource_name" hcl:"resource_name"`
	SchemaRegistryEndpoint string `json:"schema_registry_endpoint" hcl:"schema_registry_endpoint"`
	Environment            string `json:"environment" hcl:"environment"`
	Package                string `json:"package" hcl:"package"`
	Status                 string `json:"status" hcl:"status"`
}
