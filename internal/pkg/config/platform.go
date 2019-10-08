package config

// Platform represents a Confluent Platform deployment
type Platform struct {
	Name   string `json:"name" hcl:"name"`
	Server string `json:"server" hcl:"server"`
}
