package v1

type LocalPorts struct {
	BrokerPort     string `human:"Broker Port" json:"broker_port"`
	ControllerPort string `human:"Controller Port" json:"controller_port"`
	KafkaRestPort  string `human:"Kafka Rest Port" json:"kafka_rest_port"`
	PlaintextPort  string `human:"Plaintext Port" json:"plaintext_port"`
}
