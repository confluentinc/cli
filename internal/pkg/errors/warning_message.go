package errors

const (
	// api commands
	APIKeyNotRetrievableMsg = "Save the API key and secret. The secret is not retrievable later."
	APIKeyTime              = "It may take a couple of minutes for the API key to be ready."

	// kafka commands
	KafkaClusterTime = "It may take up to 5 minutes for the Kafka cluster to be ready."

	// secret commands
	SaveTheMasterKeyMsg = "Save the master key. It cannot be retrieved later."
)
