package v1

import "gopkg.in/launchdarkly/go-sdk-common.v2/lduser"

type FeatureFlags struct {
	CliValues      map[string]any `json:"values"`
	CcloudValues   map[string]any `json:"ccloud_values"`
	LastUpdateTime int64          `json:"last_update_time"`
	User           lduser.User    `json:"user"`
}

type LaunchDarklyClient int

const (
	CliLaunchDarklyClient LaunchDarklyClient = iota
	CcloudProdLaunchDarklyClient
	CcloudStagLaunchDarklyClient
	CcloudDevelLaunchDarklyClient
)
