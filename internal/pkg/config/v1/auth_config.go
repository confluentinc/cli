package v1

import ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

// AuthConfig represents an authenticated user.
type AuthConfig struct {
	User         *ccloudv1.User         `json:"user"`
	Organization *ccloudv1.Organization `json:"organization,omitempty"`
	Account      *ccloudv1.Account      `json:"account,omitempty"`
	Accounts     []*ccloudv1.Account    `json:"accounts,omitempty"`
}
