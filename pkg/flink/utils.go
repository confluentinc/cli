package flink

var (
	ConnectionTypes             = []string{"openai", "azureml", "azureopenai", "bedrock", "sagemaker", "googleai", "vertexai", "mongodb", "elastic", "pinecone", "couchbase", "confluent_jdbc", "rest"}
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
		"confluent_jdbc": {"username", "password"},
		"rest":           {"username", "password", "bearer-token", "oauth-token-endpoint", "oauth-client-id", "oauth-client-secret", "oauth-scope"},
	}

	ConnectionSecretTypeMapping = map[string][]string{
		"api-key":              {"openai", "azureml", "azureopenai", "googleai", "elastic", "pinecone"},
		"aws-access-key":       {"bedrock", "sagemaker"},
		"aws-secret-key":       {"bedrock", "sagemaker"},
		"aws-session-token":    {"bedrock", "sagemaker"},
		"service-key":          {"vertexai"},
		"username":             {"mongodb", "couchbase", "confluent_jdbc", "rest"},
		"password":             {"mongodb", "couchbase", "confluent_jdbc", "rest"},
		"bearer-token":         {"rest"},
		"oauth-token-endpoint": {"rest"},
		"oauth-client-id":      {"rest"},
		"oauth-client-secret":  {"rest"},
		"oauth-scope":          {"rest"},
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
		"confluent_jdbc": {"username", "password"},
		"rest":           {},
	}
	ConnectionSecretBackendKeyMapping = map[string]string{
		"api-key":              "API_KEY",
		"aws-access-key":       "AWS_ACCESS_KEY_ID",
		"aws-secret-key":       "AWS_SECRET_ACCESS_KEY",
		"aws-session-token":    "AWS_SESSION_TOKEN",
		"service-key":          "SERVICE_KEY",
		"username":             "USERNAME",
		"password":             "PASSWORD",
		"bearer-token":         "BEARER.TOKEN",
		"oauth-token-endpoint": "OAUTH2.TOKEN_ENDPOINT",
		"oauth-client-id":      "OAUTH2.CLIENT_ID",
		"oauth-client-secret":  "OAUTH2.CLIENT_SECRET",
		"oauth-scope":          "OAUTH2.SCOPE",
	}
)
