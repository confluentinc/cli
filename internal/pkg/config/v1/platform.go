package v1

// Platform represents a Confluent Platform deployment
type Platform struct {
	Name       string `json:"name"`
	Server     string `json:"server"`
	CaCertPath string `json:"ca_cert_path,omitempty"`
}
