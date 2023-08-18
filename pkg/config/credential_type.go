package config

type CredentialType int

const (
	Username CredentialType = iota
	APIKey
	None
)

func (c CredentialType) String() string {
	credTypes := [...]string{"username", "api-key", "none"}
	return credTypes[c]
}
