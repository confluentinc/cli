package v2

import v1 "github.com/confluentinc/cli/internal/pkg/config/v1"

type ContextState struct {
	Auth      *v1.AuthConfig `json:"auth" hcl:"auth"`
	AuthToken string         `json:"auth_token" hcl:"auth_token"`
}
