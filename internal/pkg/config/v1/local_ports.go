package v1

type LocalPorts struct {
	RestPort       string `json:"rest_port"`
	BrokerPort     string `json:"broker_port"`
	ControllerPort string `json:"controller_port"`
	PlaintextPort  string `json:"plaintext_port"`
}
