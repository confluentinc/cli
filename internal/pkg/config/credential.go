package config

import (
	"fmt"
)

// Credential represent an authentication mechanism for a Platform
type Credential struct {
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

func (c *Credential) String() string {
	switch c.CredentialType {
	case Username:
		return fmt.Sprintf("%s-%s", &c.CredentialType, c.Username)
	case APIKey:
		return fmt.Sprintf("%s-%s", &c.CredentialType, c.APIKeyPair.Key)
	default:
		panic(fmt.Sprintf("Credential type %d unknown.", c.CredentialType))
	}
}
