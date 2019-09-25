package config

// Platform represents a Confluent Platform deployment
type Platform struct {
	Name   string
	Server string `json:"server" hcl:"server"`
}
