package v1

type EnvironmentContext struct {
	CurrentFlinkCatalog       string `json:"current_flink_catalog,omitempty"`
	CurrentFlinkCloudProvider string `json:"current_flink_cloud_provider,omitempty"`
	CurrentFlinkComputePool   string `json:"current_flink_compute_pool,omitempty"`
	CurrentFlinkDatabase      string `json:"current_flink_database,omitempty"`
	CurrentFlinkRegion        string `json:"current_flink_region,omitempty"`
	CurrentIdentityPool       string `json:"current_identity_pool,omitempty"`
}

func NewEnvironmentContext() *EnvironmentContext {
	return &EnvironmentContext{}
}
