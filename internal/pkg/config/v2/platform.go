package v2

// Platform represents a Confluent Platform deployment
type Platform struct {
	Name       string `json:"name" hcl:"name"`
	Server     string `json:"server" hcl:"server"`
	CaCertPath string `json:"ca_cert_path,omitempty" hcl:"ca_cert_path" hcle:"omitempty"`
}
