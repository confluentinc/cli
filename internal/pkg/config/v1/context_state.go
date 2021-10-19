package v1

type ContextState struct {
	Auth      *AuthConfig `json:"auth" hcl:"auth"`
	AuthToken string      `json:"auth_token" hcl:"auth_token"`
}
