package types

type ApplicationOptions struct {
	MOCK_STATEMENTS_OUTPUT_DEMO bool
	HTTP_CLIENT_UNSAFE_TRACE    bool
	FLINK_GATEWAY_URL           string
	DEFAULT_PROPERTIES          map[string]string
	USER_AGENT                  string
}
