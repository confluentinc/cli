package config

type EnvironmentContext struct {
	CurrentFlinkCatalog           string                 `json:"current_flink_catalog,omitempty"`
	CurrentFlinkCloudProvider     string                 `json:"current_flink_cloud_provider,omitempty"`
	CurrentFlinkComputePool       string                 `json:"current_flink_compute_pool,omitempty"`
	CurrentFlinkDatabase          string                 `json:"current_flink_database,omitempty"`
	CurrentFlinkRegion            string                 `json:"current_flink_region,omitempty"`
	CurrentServiceAccount         string                 `json:"current_service_account,omitempty"`
	CurrentFlinkAccessType        string                 `json:"current_flink_access_type,omitempty"`
	CurrentFlinkEndpoint          string                 `json:"current_flink_endpoint,omitempty"`
	CurrentSchemaRegistryEndpoint string                 `json:"current_schema_registry_endpoint,omitempty"`
	ConnectLogsQueryState         *ConnectLogsQueryState `json:"connect_logs_state,omitempty"`
}
