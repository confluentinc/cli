package v1

import "gopkg.in/launchdarkly/go-sdk-common.v2/lduser"

type LaunchDarklyClient int

const (
	CliLaunchDarklyClient LaunchDarklyClient = iota
	CcloudProdLaunchDarklyClient
	CcloudStagLaunchDarklyClient
	CcloudDevelLaunchDarklyClient
)

type FeatureFlags struct {
	Values         map[string]any `json:"values" hcl:"values"`
	CcloudValues   map[string]any `json:"ccloud_values" hcl:"ccloud_values"`
	LastUpdateTime int64          `json:"last_update_time" hcl:"last_update_time"`
	User           lduser.User    `json:"user" hcl:"user"`
}
