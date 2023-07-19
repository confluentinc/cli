package v1

// Credential represents an authentication mechanism for a Platform
type Credential struct {
	// Deprecated
	Password string `json:"password,omitempty"`

	Name           string         `json:"name"`
	Username       string         `json:"username"`
	APIKeyPair     *APIKeyPair    `json:"api_key_pair"`
	CredentialType CredentialType `json:"credential_type"`
}
