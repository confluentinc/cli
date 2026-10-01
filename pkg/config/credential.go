package config

// Credential represents an authentication mechanism for a Platform
type Credential struct {
	Name           string         `json:"name"`
	Username       string         `json:"username"`
	APIKeyPair     *APIKeyPair    `json:"api_key_pair"`
	CredentialType CredentialType `json:"credential_type"`
}
