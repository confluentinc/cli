package kafka

type ListACLsContextKey string

const Requester ListACLsContextKey = "requester"

var (
	Clouds         = []string{"aws", "azure", "gcp"}
	Availabilities = []string{"single-zone", "multi-zone"}
	Types          = []string{"basic", "standard", "dedicated"}
)
