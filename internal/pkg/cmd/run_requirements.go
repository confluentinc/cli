package cmd

import (
	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	RunRequirement = "run-requirement"

	RequireNonAPIKeyCloudLogin              = "non-api-key-cloud-login"
	RequireNonAPIKeyCloudLoginOrOnPremLogin = "non-api-key-cloud-login-or-on-prem-login"
	RequireCloudLogin                       = "cloud-login"
	RequireCloudLoginAllowFreeTrialEnded    = "cloud-login-allow-free-trial-ended"
	RequireCloudLoginOrOnPremLogin          = "cloud-login-or-on-prem-login"
	RequireOnPremLogin                      = "on-prem-login"
	RequireUpdatesEnabled                   = "updates-enabled"
)

var (
	requireUpdatesEnabledErr = errors.NewErrorWithSuggestions(
		"you must enable updates to use this command",
		"WARNING: To guarantee compatibility, enabling updates is not recommended for Confluent Platform users.\n"+`In ~/.confluent/config.json, set "disable_updates": false`,
	)
)

// ErrIfMissingRunRequirement returns an error when a command or its parent doesn't meet a requirement;
// for example, an on-prem command shouldn't be used by a cloud user.
func ErrIfMissingRunRequirement(cmd *cobra.Command, cfg *v1.Config) error {
	if cmd == nil {
		return nil
	}

	if requirement, ok := cmd.Annotations[RunRequirement]; ok {
		switch requirement {
		case RequireCloudLogin:
			if _, err := cfg.CheckIsCloudLogin(); err != nil {
				return err
			}
		case RequireCloudLoginAllowFreeTrialEnded:
			if _, err := cfg.CheckIsCloudLoginAllowFreeTrialEnded(); err != nil {
				return err
			}
		case RequireCloudLoginOrOnPremLogin:
			if _, err := cfg.CheckIsCloudLoginOrOnPremLogin(); err != nil {
				return err
			}
		case RequireNonAPIKeyCloudLogin:
			if _, err := cfg.CheckIsNonAPIKeyCloudLogin(); err != nil {
				return err
			}
		case RequireNonAPIKeyCloudLoginOrOnPremLogin:
			if _, err := cfg.CheckIsNonAPIKeyCloudLoginOrOnPremLogin(); err != nil {
				return err
			}
		case RequireOnPremLogin:
			if _, err := cfg.CheckIsOnPremLogin(); err != nil {
				return err
			}
		case RequireUpdatesEnabled:
			if cfg.DisableUpdates {
				return requireUpdatesEnabledErr
			}
		}
	}

	return ErrIfMissingRunRequirement(cmd.Parent(), cfg)
}

func CommandRequiresCloudAuth(cmd *cobra.Command, cfg *v1.Config) bool {
	if requirement, ok := cmd.Annotations[RunRequirement]; ok {
		switch requirement {
		case RequireCloudLogin:
			return true
		case RequireCloudLoginAllowFreeTrialEnded:
			return true
		case RequireNonAPIKeyCloudLogin:
			return true
		case RequireCloudLoginOrOnPremLogin:
			return cfg.IsCloudLogin()
		case RequireNonAPIKeyCloudLoginOrOnPremLogin:
			return cfg.IsCloudLogin()
		case RequireOnPremLogin:
			return false
		}
	}
	return false
}
