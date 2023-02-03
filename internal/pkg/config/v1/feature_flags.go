package v1

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

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

// GetCcloudLaunchDarklyClient resolves to a LaunchDarkly client based on the string platform name that is passed in.
func GetCcloudLaunchDarklyClient(platformName string) (LaunchDarklyClient, error) {
	switch platformName {
	case "confluent.cloud":
		return CcloudProdLaunchDarklyClient, nil
	case "stag.cpdev.cloud":
		return CcloudStagLaunchDarklyClient, nil
	case "devel.cpdev.cloud":
		return CcloudDevelLaunchDarklyClient, nil
	default:
		return -1, errors.New(errors.NonCcloudPlatformNameErrorMsg)
	}
}
