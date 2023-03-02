package v1

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
)

type LaunchDarklyClient int

const (
	CliLaunchDarklyClient LaunchDarklyClient = iota
	CcloudProdLaunchDarklyClient
	CcloudStagLaunchDarklyClient
	CcloudDevelLaunchDarklyClient
)

type FeatureFlags struct {
	Values         map[string]any `json:"values"`
	CcloudValues   map[string]any `json:"ccloud_values"`
	LastUpdateTime int64          `json:"last_update_time"`
	User           lduser.User    `json:"user"`
}

// GetCcloudLaunchDarklyClient resolves to a LaunchDarkly client based on the string platform name that is passed in.
func GetCcloudLaunchDarklyClient(platformName string) LaunchDarklyClient {
	switch platformName {
	case "stag.cpdev.cloud":
		return CcloudStagLaunchDarklyClient
	case "devel.cpdev.cloud":
		return CcloudDevelLaunchDarklyClient
	default:
		return CcloudProdLaunchDarklyClient
	}
}
