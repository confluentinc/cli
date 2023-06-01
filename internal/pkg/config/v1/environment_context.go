package v1

type EnvironmentContext struct {
	CurrentFlinkComputePool string `json:"current_flink_compute_pool,omitempty"`
}

func NewEnvironmentContext() *EnvironmentContext {
	return &EnvironmentContext{}
}
