package v2

type CredentialType int

const (
	Username CredentialType = iota
	APIKey
)

func (c CredentialType) String() string {
	credTypes := [...]string{"username", "api-key"}
	return credTypes[c]
}
