package v0

// Platform represents a Confluent Platform deployment
type Platform struct {
	Server string `json:"server" hcl:"server"`
}
