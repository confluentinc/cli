package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/config"
)

const RunRequirement = "run-requirement"

const (
	RequireCloudLogin                       = "cloud-login"
	RequireCloudLoginAllowFreeTrialEnded    = "cloud-login-allow-free-trial-ended"
	RequireCloudLoginOrOnPremLogin          = "cloud-login-or-on-prem-login"
	RequireNonAPIKeyCloudLogin              = "non-api-key-cloud-login"
	RequireNonAPIKeyCloudLoginOrOnPremLogin = "non-api-key-cloud-login-or-on-prem-login"
	RequireCloudLogout                      = "cloud-logout"
	RequireOnPremLogin                      = "on-prem-login"
)

var wrongLoginCommandsMap = map[string]string{
	"confluent cluster": "confluent kafka cluster",
}

// ErrIfMissingRunRequirement returns an error when a command or its parent doesn't meet a requirement;
// for example, an on-prem command shouldn't be used by a cloud user.
func ErrIfMissingRunRequirement(cmd *cobra.Command, cfg *config.Config) error {
	if cmd == nil {
		return nil
	}

	if requirement, ok := cmd.Annotations[RunRequirement]; ok {
		var f func() error

		switch requirement {
		case RequireCloudLogin:
			f = cfg.CheckIsCloudLogin
		case RequireCloudLoginAllowFreeTrialEnded:
			f = cfg.CheckIsCloudLoginAllowFreeTrialEnded
		case RequireCloudLoginOrOnPremLogin:
			f = cfg.CheckIsCloudLoginOrOnPremLogin
		case RequireNonAPIKeyCloudLogin:
			f = cfg.CheckIsNonAPIKeyCloudLogin
		case RequireNonAPIKeyCloudLoginOrOnPremLogin:
			f = cfg.CheckIsNonAPIKeyCloudLoginOrOnPremLogin
		case RequireCloudLogout:
			f = cfg.CheckIsCloudLogout
		case RequireOnPremLogin:
			f = cfg.CheckIsOnPremLogin
		}

		if err := f(); err != nil {
			if err == config.RunningOnPremCommandInCloudErr {
				for topLevelCmd, suggestedCmd := range wrongLoginCommandsMap {
					if strings.HasPrefix(cmd.CommandPath(), topLevelCmd) {
						suggestCmdPath := strings.Replace(cmd.CommandPath(), topLevelCmd, suggestedCmd, 1)
						return config.RunningSimilarOnPremCommandInCloudErr(cmd.CommandPath(), suggestCmdPath)
					}
				}
			}
			return err
		}
	}

	return ErrIfMissingRunRequirement(cmd.Parent(), cfg)
}

func CommandRequiresCloudAuth(cmd *cobra.Command, cfg *config.Config) bool {
	if requirement, ok := cmd.Annotations[RunRequirement]; ok {
		switch requirement {
		case RequireCloudLogin, RequireCloudLoginAllowFreeTrialEnded, RequireNonAPIKeyCloudLogin:
			return true
		case RequireCloudLoginOrOnPremLogin, RequireNonAPIKeyCloudLoginOrOnPremLogin:
			return cfg.IsCloudLogin()
		}
	}
	return false
}
