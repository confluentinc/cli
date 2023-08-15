package config

type LocalPorts struct {
	BrokerPorts     []string `human:"Broker Ports" json:"broker_ports"`
	ControllerPorts []string `human:"Controller Ports" json:"controller_ports"`
	KafkaRestPorts  []string `human:"Kafka Rest Ports" json:"kafka_rest_ports"`
	PlaintextPorts  []string `human:"Plaintext Ports" json:"plaintext_ports"`
}
