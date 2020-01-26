package config

// Credential represent an authentication mechanism for a Platform
type Credential struct {
	Name           string         `json:"name"`
	Username       string         `json:"username"`
	Password       string         `json:"password"`
	APIKeyPair     *APIKeyPair    `json:"api_key_pair"`
	CredentialType CredentialType `json:"credential_type"`
}
type CredentialType int

const (
	Username CredentialType = iota
	APIKey
)

func (c CredentialType) String() string {
	credTypes := [...]string{"username", "api-key"}
	return credTypes[c]
}
