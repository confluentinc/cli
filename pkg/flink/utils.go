package flink

var (
	ConnectionTypes             = []string{"openai", "azureml", "azureopenai", "bedrock", "sagemaker", "googleai", "vertexai", "mongodb", "elastic", "pinecone", "couchbase", "confluent-jdbc"}
	ConnectionTypeSecretMapping = map[string][]string{
		"openai":         {"api-key"},
		"azureml":        {"api-key"},
		"azureopenai":    {"api-key"},
		"bedrock":        {"aws-access-key", "aws-secret-key", "aws-session-token"},
		"sagemaker":      {"aws-access-key", "aws-secret-key", "aws-session-token"},
		"googleai":       {"api-key"},
		"vertexai":       {"service-key"},
		"mongodb":        {"username", "password"},
		"elastic":        {"api-key"},
		"pinecone":       {"api-key"},
		"couchbase":      {"username", "password"},
		"confluent-jdbc": {"username", "password"},
	}

	ConnectionSecretTypeMapping = map[string][]string{
		"api-key":           {"openai", "azureml", "azureopenai", "googleai", "elastic", "pinecone"},
		"aws-access-key":    {"bedrock", "sagemaker"},
		"aws-secret-key":    {"bedrock", "sagemaker"},
		"aws-session-token": {"bedrock", "sagemaker"},
		"service-key":       {"vertexai"},
		"username":          {"mongodb", "couchbase", "confluent-jdbc"},
		"password":          {"mongodb", "couchbase", "confluent-jdbc"},
	}

	ConnectionRequiredSecretMapping = map[string][]string{
		"openai":         {"api-key"},
		"azureml":        {"api-key"},
		"azureopenai":    {"api-key"},
		"bedrock":        {"aws-access-key", "aws-secret-key"},
		"sagemaker":      {"aws-access-key", "aws-secret-key"},
		"googleai":       {"api-key"},
		"vertexai":       {"service-key"},
		"mongodb":        {"username", "password"},
		"elastic":        {"api-key"},
		"pinecone":       {"api-key"},
		"couchbase":      {"username", "password"},
		"confluent-jdbc": {"username", "password"},
	}
	ConnectionSecretBackendKeyMapping = map[string]string{
		"api-key":           "API_KEY",
		"aws-access-key":    "AWS_ACCESS_KEY_ID",
		"aws-secret-key":    "AWS_SECRET_ACCESS_KEY",
		"aws-session-token": "AWS_SESSION_TOKEN",
		"service-key":       "SERVICE_KEY",
		"username":          "USERNAME",
		"password":          "PASSWORD",
	}
)
