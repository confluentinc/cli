package config

// Platform represents a Confluent Platform deployment
type Platform struct {
	Server 		string	`json:"server" hcl:"server"`
	CaCertPath  string	`json:"ca_cert_path,omitempty" hcl:"ca_cert_path" hcle:"omitempty"`
}

func (p *Platform) String() string {
	return p.Server
}
