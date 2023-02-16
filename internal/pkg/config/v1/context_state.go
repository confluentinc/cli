package v1

type ContextState struct {
	Auth             *AuthConfig `json:"auth" hcl:"auth"`
	AuthToken        string      `json:"auth_token" hcl:"auth_token"`
	AuthRefreshToken string      `json:"auth_refresh_token" hcl:"auth_refresh_token"`
	Salt             []byte      `json:"salt,omitempty"`
	Nonce            []byte      `json:"nonce,omitempty"`
}
