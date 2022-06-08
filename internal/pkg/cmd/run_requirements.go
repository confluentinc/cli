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
	RequireCloudLoginOrOnPremLogin          = "cloud-login-or-on-prem-login"
	RequireOnPremLogin                      = "on-prem-login"
	RequireUpdatesEnabled                   = "updates-enabled"
)

const signupSuggestion = `If you need a Confluent Cloud account, sign up with "confluent cloud-signup".`

var (
	requireCloudLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Cloud to use this command",
		"Log in with \"confluent login\".\n"+signupSuggestion,
	)
	requireCloudLoginOrOnPremErr = errors.NewErrorWithSuggestions(
		"you must log in to use this command",
		"Log in with \"confluent login\".\n"+signupSuggestion,
	)
	requireNonAPIKeyCloudLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Cloud with a username and password to use this command",
		"Log in with \"confluent login\".\n"+signupSuggestion,
	)
	requireNonAPIKeyCloudLoginOrOnPremLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Cloud with a username and password or log in to Confluent Platform to use this command",
		"Log in with \"confluent login\" or \"confluent login --url <mds-url>\".\n"+signupSuggestion,
	)
	requireOnPremLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Platform to use this command",
		`Log in with "confluent login --url <mds-url>".`,
	)
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
			if !cfg.IsCloudLogin() {
				return requireCloudLoginErr
			}
		case RequireCloudLoginOrOnPremLogin:
			if !(cfg.IsCloudLogin() || cfg.IsOnPremLogin()) {
				return requireCloudLoginOrOnPremErr
			}
		case RequireNonAPIKeyCloudLogin:
			if !(cfg.CredentialType() != v1.APIKey && cfg.IsCloudLogin()) {
				return requireNonAPIKeyCloudLoginErr
			}
		case RequireNonAPIKeyCloudLoginOrOnPremLogin:
			if !(cfg.CredentialType() != v1.APIKey && cfg.IsCloudLogin() || cfg.IsOnPremLogin()) {
				return requireNonAPIKeyCloudLoginOrOnPremLoginErr
			}
		case RequireOnPremLogin:
			if !cfg.IsOnPremLogin() {
				return requireOnPremLoginErr
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
