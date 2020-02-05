package v1

import (
	"fmt"

	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
)

// Credential represent an authentication mechanism for a Platform
type Credential struct {
	Username       string
	Password       string
	APIKeyPair     *v0.APIKeyPair
	CredentialType CredentialType
}

func (c *Credential) String() string {
	switch c.CredentialType {
	case Username:
		return fmt.Sprintf("%d-%s", c.CredentialType, c.Username)
	case APIKey:
		return fmt.Sprintf("%d-%s", c.CredentialType, c.APIKeyPair.Key)
	default:
		panic(fmt.Sprintf("Credential type %d unknown.", c.CredentialType))
	}
}
