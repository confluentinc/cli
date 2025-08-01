package flink

var (
	ConnectionTypes             = []string{"openai", "azureml", "azureopenai", "bedrock", "sagemaker", "googleai", "vertexai", "mongodb", "elastic", "pinecone", "couchbase", "confluent_jdbc", "rest", "mcp_server"}
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
		"rest":           {"username", "password", "token", "token-endpoint", "client-id", "client-secret", "scope"},
		"mcp_server":     {"api-key", "token", "token-endpoint", "client-id", "client-secret", "scope", "sse-endpoint"},
	}

	ConnectionSecretTypeMapping = map[string][]string{
		"api-key":           {"openai", "azureml", "azureopenai", "googleai", "elastic", "pinecone", "mcp_server"},
		"aws-access-key":    {"bedrock", "sagemaker"},
		"aws-secret-key":    {"bedrock", "sagemaker"},
		"aws-session-token": {"bedrock", "sagemaker"},
		"service-key":       {"vertexai"},
		"username":          {"mongodb", "couchbase", "confluent_jdbc", "rest"},
		"password":          {"mongodb", "couchbase", "confluent_jdbc", "rest"},
		"token":             {"rest", "mcp_server"},
		"token-endpoint":    {"rest", "mcp_server"},
		"client-id":         {"rest", "mcp_server"},
		"client-secret":     {"rest", "mcp_server"},
		"scope":             {"rest", "mcp_server"},
		"sse-endpoint":      {"mcp_server"},
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
		"mcp_server":     {},
	}

	ConnectionOneOfRequiredSecretsMapping = map[string][][]string{
		"rest":       {{"username", "password"}, {"token"}, {"token-endpoint", "client-id", "client-secret", "scope"}},
		"mcp_server": {{"api-key"}, {"token"}, {"token-endpoint", "client-id", "client-secret", "scope"}},
	}

	ConnectionSecretBackendKeyMapping = map[string]string{
		"api-key":           "API_KEY",
		"aws-access-key":    "AWS_ACCESS_KEY_ID",
		"aws-secret-key":    "AWS_SECRET_ACCESS_KEY",
		"aws-session-token": "AWS_SESSION_TOKEN",
		"service-key":       "SERVICE_KEY",
		"username":          "USERNAME",
		"password":          "PASSWORD",
		"token":             "BEARER_TOKEN",
		"token-endpoint":    "OAUTH2_TOKEN_ENDPOINT",
		"client-id":         "OAUTH2_CLIENT_ID",
		"client-secret":     "OAUTH2_CLIENT_SECRET",
		"scope":             "OAUTH2_SCOPE",
		"sse-endpoint":      "SSE_ENDPOINT",
	}
)
