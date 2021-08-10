package cmd

import (
	"github.com/spf13/cobra"

	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
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

var (
	requireCloudLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Cloud to use this command",
		`Log in with "confluent login"`,
	)
	requireCloudLoginOrOnPremErr = errors.NewErrorWithSuggestions(
		"you must log in to use this command",
		`Log in with "confluent login"`,
	)
	requireNonAPIKeyCloudLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Cloud with a username and password to use this command",
		`Log in with "confluent login"`,
	)
	requireNonAPIKeyCloudLoginOrOnPremLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Cloud with a username and password or log in to Confluent Platform to use this command",
		`Log in with "confluent login" or "confluent login --url <mds-url>"`,
	)
	requireOnPremLoginErr = errors.NewErrorWithSuggestions(
		"you must log in to Confluent Platform to use this command",
		`Log in with "confluent login --url <mds-url>"`,
	)
	requireUpdatesEnabledErr = errors.NewErrorWithSuggestions(
		"you must enable updates to use this command",
		`In ~/.confluent/config.json, set "disable_updates": false`,
	)
)

// ErrIfMissingRunRequirement returns an error when a command or its parent doesn't meet a requirement;
// for example, an on-prem command shouldn't be used by a cloud user.
func ErrIfMissingRunRequirement(cmd *cobra.Command, cfg *v3.Config) error {
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
			if !(cfg.CredentialType() != v2.APIKey && cfg.IsCloudLogin()) {
				return requireNonAPIKeyCloudLoginErr
			}
		case RequireNonAPIKeyCloudLoginOrOnPremLogin:
			if !(cfg.CredentialType() != v2.APIKey && cfg.IsCloudLogin() || cfg.IsOnPremLogin()) {
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
