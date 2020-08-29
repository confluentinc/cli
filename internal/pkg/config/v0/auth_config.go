package v0

import v1 "github.com/confluentinc/cc-structs/kafka/org/v1"

// AuthConfig represents an authenticated user.
type AuthConfig struct {
	User     *v1.User      `json:"user" hcl:"user"`
	Account  *v1.Account   `json:"account" hcl:"account"`
	Accounts []*v1.Account `json:"accounts" hcl:"accounts"`
}
