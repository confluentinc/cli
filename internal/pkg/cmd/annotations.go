package cmd

const DoNotTrack = "do-not-track-analytics"

const (
	RunRequirement = "run-requirement"

	RequireNonAPIKeyCloudLogin              = "non-api-key-cloud-login"
	RequireNonAPIKeyCloudLoginOrOnPremLogin = "non-api-key-cloud-login-or-on-prem-login"
	RequireCloudLogin                       = "cloud-login"
	RequireCloudLoginOrOnPremLogin          = "cloud-login-or-on-prem-login"
	RequireOnPremLogin                      = "on-prem-login"
	RequireUpdatesEnabled                   = "updates-enabled"
)
