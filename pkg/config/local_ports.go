package config

type LocalPorts struct {
	BrokerPorts     []string `human:"Broker Ports" json:"broker_ports"`
	ControllerPorts []string `human:"Controller Ports" json:"controller_ports"`
	KafkaRestPort   string   `human:"Kafka Rest Port" json:"kafka_rest_port"`
	PlaintextPorts  []string `human:"Plaintext Ports" json:"plaintext_ports"`
}
