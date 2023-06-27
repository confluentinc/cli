package v1

type EnvironmentContext struct {
	CurrentFlinkComputePool string `json:"current_flink_compute_pool,omitempty"`
	CurrentIdentityPool     string `json:"current_identity_pool,omitempty"`
}

func NewEnvironmentContext() *EnvironmentContext {
	return &EnvironmentContext{}
}
