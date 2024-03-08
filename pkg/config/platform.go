package config

// Platform represents a Confluent Platform deployment
type Platform struct {
	Name       string `json:"name"`
	Server     string `json:"server"`
	CaCertPath string `json:"ca_cert_path,omitempty"`
}

func (p *Platform) GetName() string {
	if p != nil {
		return p.Name
	}
	return ""
}

func (p *Platform) GetServer() string {
	if p != nil {
		return p.Server
	}
	return ""
}

func (p *Platform) GetCaCertPath() string {
	if p != nil {
		return p.CaCertPath
	}
	return ""
}
