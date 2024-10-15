package flink

var (
	ConnectionTypes             = []string{"openai", "azureml", "azureopenai", "bedrock", "sagemaker", "googleai", "vertexai", "mongodb", "elastic", "pinecone"}
	ConnectionTypeSecretMapping = map[string][]string{
		"openai":      {"api-key"},
		"azureml":     {"api-key"},
		"azureopenai": {"api-key"},
		"bedrock":     {"aws-access-key", "aws-secret-key"},
		"sagemaker":   {"aws-access-key", "aws-secret-key"},
		"googleai":    {"api-key"},
		"vertexai":    {"service-key"},
		"mongodb":     {"username", "password"},
		"elastic":     {"api-key"},
		"pinecone":    {"api-key"},
	}

	ConnectionSecretTypeMapping = map[string][]string{
		"api-key":        {"openai", "azureml", "azureopenai", "googleai", "elastic", "pinecone"},
		"aws-access-key": {"bedrock", "sagemaker"},
		"aws-secret-key": {"bedrock", "sagemaker"},
		"service-key":    {"vertexai"},
		"username":       {"mongodb"},
		"password":       {"mongodb"},
	}

	ConnectionRequiredSecretMapping = map[string][]string{
		"openai":      {"api-key"},
		"azureml":     {"api-key"},
		"azureopenai": {"api-key"},
		"bedrock":     {"aws-access-key", "aws-secret-key"},
		"sagemaker":   {"aws-access-key", "aws-secret-key"},
		"googleai":    {"api-key"},
		"vertexai":    {"service-key"},
		"mongodb":     {"username", "password"},
		"elastic":     {"api-key"},
		"pinecone":    {"api-key"},
	}
	ConnectionSecretBackendKeyMapping = map[string]string{
		"api-key":        "API_KEY",
		"aws-access-key": "AWS_ACCESS_KEY_ID",
		"aws-secret-key": "AWS_SECRET_ACCESS_KEY",
		"service-key":    "SERVICE_KEY",
		"username":       "USERNAME",
		"password":       "PASSWORD",
	}
)
