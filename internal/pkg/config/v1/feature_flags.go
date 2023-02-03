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
	Values         map[string]interface{} `json:"values" hcl:"values"`
	CcloudValues   map[string]interface{} `json:"ccloud_values" hcl:"ccloud_values"`
	LastUpdateTime int64                  `json:"last_update_time" hcl:"last_update_time"`
	User           lduser.User            `json:"user" hcl:"user"`
}

// ResolveToLaunchDarklyClient resolves to a LaunchDarkly client based on the string platform name that is passed in. It
// defaults to CliLaunchDarklyClient (CLI LaunchDarkly project).
func ResolveToLaunchDarklyClient(platformName string) LaunchDarklyClient {
	switch platformName {
	case "confluent.cloud":
		return CcloudProdLaunchDarklyClient
	case "stag.cpdev.cloud":
		return CcloudStagLaunchDarklyClient
	case "devel.cpdev.cloud":
		return CcloudDevelLaunchDarklyClient
	default:
		return CliLaunchDarklyClient
	}
}

func (c LaunchDarklyClient) IsCcloudLaunchDarklyClient() bool {
	return c == CcloudProdLaunchDarklyClient ||
		c == CcloudStagLaunchDarklyClient ||
		c == CcloudDevelLaunchDarklyClient
}
