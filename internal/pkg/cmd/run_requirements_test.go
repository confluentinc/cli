package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	testserver "github.com/confluentinc/cli/test/test-server"
)

var (
	noContextCfg = new(v1.Config)

	regularOrgContextState = &v1.ContextState{
		Auth: &v1.AuthConfig{
			Organization: testserver.RegularOrg,
		},
	}

	endOfFreeTrialSuspendedOrgContextState = &v1.ContextState{
		Auth: &v1.AuthConfig{
			Organization: testserver.SuspendedOrg(orgv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL),
		},
	}

	normalSuspendedOrgContextState = &v1.ContextState{
		Auth: &v1.AuthConfig{
			Organization: testserver.SuspendedOrg(orgv1.SuspensionEventType_SUSPENSION_EVENT_CUSTOMER_INITIATED_ORG_DEACTIVATION),
		},
	}

	cloudCfg = func(contextState *v1.ContextState) *v1.Config {
		return &v1.Config{
			Contexts: map[string]*v1.Context{"cloud": {
				PlatformName: testserver.TestCloudURL.String(),
				State:        contextState,
			}},
			CurrentContext: "cloud",
			IsTest:         true,
		}
	}

	apiKeyCloudCfg = &v1.Config{
		Contexts: map[string]*v1.Context{"cloud": {
			PlatformName: testserver.TestCloudURL.String(),
			Credential:   &v1.Credential{CredentialType: v1.APIKey},
			State:        regularOrgContextState,
		}},
		CurrentContext: "cloud",
		IsTest:         true,
	}

	nonAPIKeyCloudCfg = &v1.Config{
		Contexts: map[string]*v1.Context{"cloud": {
			PlatformName: testserver.TestCloudURL.String(),
			Credential:   &v1.Credential{CredentialType: v1.Username},
			State:        regularOrgContextState,
		}},
		CurrentContext: "cloud",
		IsTest:         true,
	}

	onPremCfg = &v1.Config{
		Contexts: map[string]*v1.Context{"on-prem": {
			Credential:   new(v1.Credential),
			PlatformName: "https://example.com",
			State:        &v1.ContextState{AuthToken: "token"},
		}},
		CurrentContext: "on-prem",
	}

	updatesDisabledCfg = &v1.Config{DisableUpdates: true}

	updatesEnabledCfg = &v1.Config{DisableUpdates: false}
)

func TestErrIfMissingRunRequirement_NoError(t *testing.T) {
	for _, test := range []struct {
		req string
		cfg *v1.Config
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
		{RequireUpdatesEnabled, updatesEnabledCfg},
	} {
		cmd := &cobra.Command{Annotations: map[string]string{RunRequirement: test.req}}
		err := ErrIfMissingRunRequirement(cmd, test.cfg)
		require.NoError(t, err)
	}
}

func TestErrIfMissingRunRequirement_Error(t *testing.T) {
	for _, test := range []struct {
		req string
		cfg *v1.Config
		err error
	}{
		{RequireCloudLogin, onPremCfg, requireCloudLoginErr},
		{RequireCloudLogin, cloudCfg(endOfFreeTrialSuspendedOrgContextState), requireCloudLoginFreeTrialEndedOrgUnsuspendedErr},
		{RequireCloudLogin, cloudCfg(normalSuspendedOrgContextState), requireCloudLoginOrgUnsuspendedErr},
		{RequireCloudLoginAllowFreeTrialEnded, onPremCfg, requireCloudLoginErr},
		{RequireCloudLoginAllowFreeTrialEnded, cloudCfg(normalSuspendedOrgContextState), requireCloudLoginOrgUnsuspendedErr},
		{RequireCloudLoginOrOnPremLogin, noContextCfg, requireCloudLoginOrOnPremErr},
		{RequireCloudLoginOrOnPremLogin, noContextCfg, requireCloudLoginOrOnPremErr},
		{RequireNonAPIKeyCloudLogin, apiKeyCloudCfg, requireNonAPIKeyCloudLoginErr},
		{RequireNonAPIKeyCloudLoginOrOnPremLogin, apiKeyCloudCfg, requireNonAPIKeyCloudLoginOrOnPremLoginErr},
		{RequireNonAPIKeyCloudLoginOrOnPremLogin, apiKeyCloudCfg, requireNonAPIKeyCloudLoginOrOnPremLoginErr},
		{RequireOnPremLogin, cloudCfg(regularOrgContextState), requireOnPremLoginErr},
		{RequireUpdatesEnabled, updatesDisabledCfg, requireUpdatesEnabledErr},
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
	require.Equal(t, err, requireCloudLoginErr)
}
