package v1

// APIKeyPair holds an API Key and Secret.
type APIKeyPair struct {
	Key    string `json:"api_key"`
	Secret string `json:"api_secret"`
}
