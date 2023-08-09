package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/internal/pkg/config"
	testserver "github.com/confluentinc/cli/test/test-server"
)

var (
	noContextCfg = new(config.Config)

	regularOrgContextState                 = &config.ContextState{Auth: &config.AuthConfig{Organization: testserver.RegularOrg}}
	endOfFreeTrialSuspendedOrgContextState = &config.ContextState{Auth: &config.AuthConfig{Organization: testserver.SuspendedOrg(ccloudv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL)}}
	normalSuspendedOrgContextState         = &config.ContextState{Auth: &config.AuthConfig{Organization: testserver.SuspendedOrg(ccloudv1.SuspensionEventType_SUSPENSION_EVENT_CUSTOMER_INITIATED_ORG_DEACTIVATION)}}

	cloudCfg = func(contextState *config.ContextState) *config.Config {
		return &config.Config{
			Contexts: map[string]*config.Context{"cloud": {
				PlatformName: testserver.TestCloudUrl.String(),
				State:        contextState,
			}},
			CurrentContext: "cloud",
			IsTest:         true,
		}
	}

	apiKeyCloudCfg = &config.Config{
		Contexts: map[string]*config.Context{"cloud": {
			PlatformName: testserver.TestCloudUrl.String(),
			Credential:   &config.Credential{CredentialType: config.APIKey},
			State:        regularOrgContextState,
		}},
		CurrentContext: "cloud",
		IsTest:         true,
	}

	nonAPIKeyCloudCfg = &config.Config{
		Contexts: map[string]*config.Context{"cloud": {
			PlatformName: testserver.TestCloudUrl.String(),
			Credential:   &config.Credential{CredentialType: config.Username},
			State:        regularOrgContextState,
		}},
		CurrentContext: "cloud",
		IsTest:         true,
	}

	onPremCfg = &config.Config{
		Contexts: map[string]*config.Context{"on-prem": {
			Credential:   new(config.Credential),
			PlatformName: "https://example.com",
			State:        &config.ContextState{AuthToken: "token"},
		}},
		CurrentContext: "on-prem",
	}
)

func TestErrIfMissingRunRequirement_NoError(t *testing.T) {
	for _, test := range []struct {
		req string
		cfg *config.Config
	}{
		{RequireCloudLogin, cloudCfg(regularOrgContextState)},
		{RequireCloudLoginAllowFreeTrialEnded, cloudCfg(regularOrgContextState)},
		{RequireCloudLoginAllowFreeTrialEnded, cloudCfg(endOfFreeTrialSuspendedOrgContextState)},
		{RequireCloudLoginOrOnPremLogin, cloudCfg(regularOrgContextState)},
		{RequireCloudLoginOrOnPremLogin, onPremCfg},
		{RequireNonAPIKeyCloudLogin, nonAPIKeyCloudCfg},
		{RequireNonAPIKeyCloudLoginOrOnPremLogin, nonAPIKeyCloudCfg},
		{RequireNonAPIKeyCloudLoginOrOnPremLogin, onPremCfg},
		{RequireOnPremLogin, onPremCfg},
	} {
		cmd := &cobra.Command{Annotations: map[string]string{RunRequirement: test.req}}
		err := ErrIfMissingRunRequirement(cmd, test.cfg)
		require.NoError(t, err)
	}
}

func TestErrIfMissingRunRequirement_Error(t *testing.T) {
	for _, test := range []struct {
		req string
		cfg *config.Config
		err error
	}{
		{RequireCloudLogin, onPremCfg, config.RequireCloudLoginErr},
		{RequireCloudLogin, cloudCfg(endOfFreeTrialSuspendedOrgContextState), config.RequireCloudLoginFreeTrialEndedOrgUnsuspendedErr},
		{RequireCloudLogin, cloudCfg(normalSuspendedOrgContextState), config.RequireCloudLoginOrgUnsuspendedErr},
		{RequireCloudLoginAllowFreeTrialEnded, onPremCfg, config.RequireCloudLoginErr},
		{RequireCloudLoginAllowFreeTrialEnded, cloudCfg(normalSuspendedOrgContextState), config.RequireCloudLoginOrgUnsuspendedErr},
		{RequireCloudLoginOrOnPremLogin, noContextCfg, config.RequireCloudLoginOrOnPremErr},
		{RequireCloudLoginOrOnPremLogin, noContextCfg, config.RequireCloudLoginOrOnPremErr},
		{RequireNonAPIKeyCloudLogin, apiKeyCloudCfg, config.RequireNonAPIKeyCloudLoginErr},
		{RequireNonAPIKeyCloudLoginOrOnPremLogin, apiKeyCloudCfg, config.RequireNonAPIKeyCloudLoginOrOnPremLoginErr},
		{RequireNonAPIKeyCloudLoginOrOnPremLogin, apiKeyCloudCfg, config.RequireNonAPIKeyCloudLoginOrOnPremLoginErr},
		{RequireOnPremLogin, cloudCfg(regularOrgContextState), config.RunningOnPremCommandInCloudErr},
	} {
		cmd := &cobra.Command{Annotations: map[string]string{RunRequirement: test.req}}
		err := ErrIfMissingRunRequirement(cmd, test.cfg)
		require.Error(t, err)
		require.Equal(t, test.err, err)
	}
}

func TestErrIfMissingRunRequirement_Root(t *testing.T) {
	err := ErrIfMissingRunRequirement(&cobra.Command{}, nil)
	require.NoError(t, err)
}

func TestErrIfMissingRunRequirement_Subcommand(t *testing.T) {
	a := &cobra.Command{Annotations: map[string]string{RunRequirement: RequireCloudLogin}}
	b := &cobra.Command{}
	a.AddCommand(b)

	err := ErrIfMissingRunRequirement(b, onPremCfg)
	require.Error(t, err)
	require.Equal(t, err, config.RequireCloudLoginErr)
}
