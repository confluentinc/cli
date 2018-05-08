package shared

import (
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
)

type AuthConfig struct {
	User      *orgv1.User    `json:"user" hcl:"user"`
	Account   *orgv1.Account `json:"account" hcl:"account"`
}
