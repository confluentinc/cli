package cmd

import (
	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

const RunRequirement = "run-requirement"

const (
	RequireCloudLogin                       = "cloud-login"
	RequireCloudLoginAllowFreeTrialEnded    = "cloud-login-allow-free-trial-ended"
	RequireCloudLoginOrOnPremLogin          = "cloud-login-or-on-prem-login"
	RequireNonAPIKeyCloudLogin              = "non-api-key-cloud-login"
	RequireNonAPIKeyCloudLoginOrOnPremLogin = "non-api-key-cloud-login-or-on-prem-login"
	RequireNonCloudLogin                    = "non-cloud-login"
	RequireOnPremLogin                      = "on-prem-login"
)

// ErrIfMissingRunRequirement returns an error when a command or its parent doesn't meet a requirement;
// for example, an on-prem command shouldn't be used by a cloud user.
func ErrIfMissingRunRequirement(cmd *cobra.Command, cfg *v1.Config) error {
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
		case RequireNonCloudLogin:
			f = cfg.CheckIsNonCloudLogin
		case RequireOnPremLogin:
			f = cfg.CheckIsOnPremLogin
		}

		if err := f(); err != nil {
			return err
		}
	}

	return ErrIfMissingRunRequirement(cmd.Parent(), cfg)
}

func CommandRequiresCloudAuth(cmd *cobra.Command, cfg *v1.Config) bool {
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
