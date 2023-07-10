package v1

type EnvironmentContext struct {
	CurrentFlinkComputePool string `json:"current_flink_compute_pool,omitempty"`
	CurrentIdentityPool     string `json:"current_identity_pool,omitempty"`
	CurrentFlinkCatalog     string `json:"current_flink_catalog,omitempty"`
	CurrentFlinkDatabase    string `json:"current_flink_database,omitempty"`
}

func NewEnvironmentContext() *EnvironmentContext {
	return &EnvironmentContext{}
}
