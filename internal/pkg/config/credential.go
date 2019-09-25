package config

// Credential represent an authentication mechanism for a Platform
type Credential struct {
	Name           string
	Username       string
	Password       string
	APIKeyPair     *APIKeyPair
	CredentialType CredentialType
}
type CredentialType int

const (
	Username CredentialType = iota
	APIKey
)

func (c *CredentialType) String() string {
	credTypes := [...]string{"username", "api-key"}
	return credTypes[*c]
}
