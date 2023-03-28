package v1

type LocalPorts struct {
	KafkaRestPort  string `json:"kafka_rest_port"`
	BrokerPort     string `json:"broker_port"`
	ControllerPort string `json:"controller_port"`
	PlaintextPort  string `json:"plaintext_port"`
}
