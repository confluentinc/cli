package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
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
			Organization: testserver.SuspendedOrg(ccloudv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL),
		},
	}

	normalSuspendedOrgContextState = &v1.ContextState{
		Auth: &v1.AuthConfig{
			Organization: testserver.SuspendedOrg(ccloudv1.SuspensionEventType_SUSPENSION_EVENT_CUSTOMER_INITIATED_ORG_DEACTIVATION),
		},
	}

	cloudCfg = func(contextState *v1.ContextState) *v1.Config {
		return &v1.Config{
			Contexts: map[string]*v1.Context{"cloud": {
				PlatformName: testserver.TestCloudUrl.String(),
				State:        contextState,
			}},
			CurrentContext: "cloud",
			IsTest:         true,
		}
	}

	apiKeyCloudCfg = &v1.Config{
		Contexts: map[string]*v1.Context{"cloud": {
			PlatformName: testserver.TestCloudUrl.String(),
			Credential:   &v1.Credential{CredentialType: v1.APIKey},
			State:        regularOrgContextState,
		}},
		CurrentContext: "cloud",
		IsTest:         true,
	}

	nonAPIKeyCloudCfg = &v1.Config{
		Contexts: map[string]*v1.Context{"cloud": {
			PlatformName: testserver.TestCloudUrl.String(),
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
		{RequireCloudLogin, onPremCfg, v1.RequireCloudLoginErr},
		{RequireCloudLogin, cloudCfg(endOfFreeTrialSuspendedOrgContextState), v1.RequireCloudLoginFreeTrialEndedOrgUnsuspendedErr},
		{RequireCloudLogin, cloudCfg(normalSuspendedOrgContextState), v1.RequireCloudLoginOrgUnsuspendedErr},
		{RequireCloudLoginAllowFreeTrialEnded, onPremCfg, v1.RequireCloudLoginErr},
		{RequireCloudLoginAllowFreeTrialEnded, cloudCfg(normalSuspendedOrgContextState), v1.RequireCloudLoginOrgUnsuspendedErr},
		{RequireCloudLoginOrOnPremLogin, noContextCfg, v1.RequireCloudLoginOrOnPremErr},
		{RequireCloudLoginOrOnPremLogin, noContextCfg, v1.RequireCloudLoginOrOnPremErr},
		{RequireNonAPIKeyCloudLogin, apiKeyCloudCfg, v1.RequireNonAPIKeyCloudLoginErr},
		{RequireNonAPIKeyCloudLoginOrOnPremLogin, apiKeyCloudCfg, v1.RequireNonAPIKeyCloudLoginOrOnPremLoginErr},
		{RequireNonAPIKeyCloudLoginOrOnPremLogin, apiKeyCloudCfg, v1.RequireNonAPIKeyCloudLoginOrOnPremLoginErr},
		{RequireOnPremLogin, cloudCfg(regularOrgContextState), v1.RequireOnPremLoginErr},
		{RequireUpdatesEnabled, updatesDisabledCfg, v1.RequireUpdatesEnabledErr},
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
	require.Equal(t, err, v1.RequireCloudLoginErr)
}
