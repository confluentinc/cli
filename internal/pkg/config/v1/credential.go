package v1

// Credential represent an authentication mechanism for a Platform
type Credential struct {
	Name           string         `json:"name"`
	Username       string         `json:"username"`
	Password       string         `json:"password"`
	APIKeyPair     *APIKeyPair    `json:"api_key_pair"`
	CredentialType CredentialType `json:"credential_type"`
}
