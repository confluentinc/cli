package config

// Platform represents a Confluent Platform deployment
type Platform struct {
	Server string `json:"server" hcl:"server"`
}

func (p *Platform) String() string {
	return p.Server
}

