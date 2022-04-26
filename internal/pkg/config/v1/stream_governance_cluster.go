package v1

type StreamGovernanceV1Cluster struct {
	Name                   string `json:"display_name" hcl:"display_name"`
	SchemaRegistryEndpoint string `json:"schema_registry_endpoint" hcl:"schema_registry_endpoint"`
	Environment            string `json:"environment" hcl:"environment"`
	Package                string `json:"package" hcl:"package"`
	Cloud                  string `json:"cloud" hcl:"cloud"`
	Region                 string `json:"region" hcl:"region"`
	Status                 string `json:"status" hcl:"status"`
}
