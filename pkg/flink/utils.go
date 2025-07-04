package flink

var (
	ConnectionTypes             = []string{"openai", "azureml", "azureopenai", "bedrock", "sagemaker", "googleai", "vertexai", "mongodb", "elastic", "pinecone", "couchbase", "confluent_jdbc", "mcp_server"}
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
		"mcp_server":     {"auth-type"},
	}

	ConnectionSecretTypeMapping = map[string][]string{
		"api-key":               {"openai", "azureml", "azureopenai", "googleai", "elastic", "pinecone", "mcp_server"},
		"aws-access-key":        {"bedrock", "sagemaker"},
		"aws-secret-key":        {"bedrock", "sagemaker"},
		"aws-session-token":     {"bedrock", "sagemaker"},
		"service-key":           {"vertexai"},
		"username":              {"mongodb", "couchbase", "confluent_jdbc"},
		"password":              {"mongodb", "couchbase", "confluent_jdbc"},
		"auth-type":             {"mcp_server"},
		"bearer-token":          {"mcp_server"},
		"oauth2-token-endpoint": {"mcp_server"},
		"oauth2-client-secret":  {"mcp_server"},
		"oauth2-client-id":      {"mcp_server"},
		"oauth2-scope":          {"mcp_server"},
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
		"mcp_server":     {"auth-type"},
	}
	ConnectionSecretBackendKeyMapping = map[string]string{
		"api-key":               "API_KEY",
		"aws-access-key":        "AWS_ACCESS_KEY_ID",
		"aws-secret-key":        "AWS_SECRET_ACCESS_KEY",
		"aws-session-token":     "AWS_SESSION_TOKEN",
		"service-key":           "SERVICE_KEY",
		"username":              "USERNAME",
		"password":              "PASSWORD",
		"auth-type":             "AUTH_TYPE",
		"bearer-token":          "BEARER_TOKEN",
		"oauth2-token-endpoint": "OAUTH2_TOKEN_ENDPOINT",
		"oauth2-client-secret":  "OAUTH2_CLIENT_SECRET",
		"oauth2-client-id":      "OAUTH2_CLIENT_ID",
		"oauth2-scope":          "OAUTH2_SCOPE",
	}

	ConnectionTypeDynamicKeyMapping = map[string]string{
		"mcp_server": "auth-type",
	}

	ConnectionDynamicRequiredSecretMapping = map[string]map[string][]string{
		"mcp_server": {
			"NO_AUTH": {},
			"API_KEY": {"api-key"},
			"BEARER":  {"bearer-token"},
			"OAUTH2":  {"oauth2-token-endpoint", "oauth2-client-id", "oauth2-client-secret", "oauth2-scope"},
		},
	}

	ConnectionDynamicSecretMapping = map[string]map[string][]string{
		"mcp_server": {
			"NO_AUTH": {},
			"API_KEY": {"api-key"},
			"BEARER":  {"bearer-token"},
			"OAUTH2":  {"oauth2-token-endpoint", "oauth2-client-id", "oauth2-client-secret", "oauth2-scope"},
		},
	}
)
