package v1

import "gopkg.in/launchdarkly/go-sdk-common.v2/lduser"

type LaunchDarkly struct {
	FlagValues     map[string]interface{} `json:"flag_values hcl:flag_values"`
	FlagUpdateTime int64                  `json:"flag_update_time hcl:flag_update_time"`
	User           lduser.User            `json:"flag_user hcl: flag_user"`
}
