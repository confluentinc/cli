package v2

import (
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
)

// Credential represent an authentication mechanism for a Platform
type Credential struct {
	Name           string         `json:"name"`
	Username       string         `json:"username"`
	Password       string         `json:"password"`
	APIKeyPair     *v0.APIKeyPair `json:"api_key_pair"`
	CredentialType CredentialType `json:"credential_type"`
}
