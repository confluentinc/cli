package v1

import "gopkg.in/launchdarkly/go-sdk-common.v2/lduser"

type FeatureFlags struct {
	Values         map[string]interface{} `json:"values" hcl:"values"`
	LastUpdateTime int64                  `json:"last_update_time" hcl:"last_update_time"`
	User           lduser.User            `json:"user" hcl:"user"`
}
