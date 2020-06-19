package v1

import orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

// AuthConfig represents an authenticated user.
type AuthConfig struct {
	User          *orgv1.User          `json:"user"`
	Account       *orgv1.Account       `json:"account"`
	Accounts      []*orgv1.Account     `json:"accounts"`
	Organization  *orgv1.Organization  `json:"organization"`
}
