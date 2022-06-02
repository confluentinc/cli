package cmd

import (
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	testserver "github.com/confluentinc/cli/test/test-server"
)

var (
	noContextCfg = new(v1.Config)

	state = &v1.ContextState{
		Auth: &v1.AuthConfig{
			Organization: &orgv1.Organization{
				Id:   321,
				Name: "test-org",
			},
		},
	}

	cloudCfg = &v1.Config{
		Contexts: map[string]*v1.Context{"cloud": {
			PlatformName: testserver.TestCloudURL.String(),
			State:        state,
		}},
		CurrentContext: "cloud",
		IsTest:         true,
	}

	apiKeyCloudCfg = &v1.Config{
		Contexts: map[string]*v1.Context{"cloud": {
			PlatformName: testserver.TestCloudURL.String(),
			Credential:   &v1.Credential{CredentialType: v1.APIKey},
			State:        state,
		}},
		CurrentContext: "cloud",
		IsTest:         true,
	}

	nonAPIKeyCloudCfg = &v1.Config{
		Contexts: map[string]*v1.Context{"cloud": {
			PlatformName: testserver.TestCloudURL.String(),
			Credential:   &v1.Credential{CredentialType: v1.Username},
			State:        state,
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
		{RequireCloudLogin, cloudCfg},
		{RequireLenientCloudLogin, cloudCfg},
		{RequireCloudLoginOrOnPremLogin, cloudCfg},
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
		{RequireLenientCloudLogin, onPremCfg, requireLenientCloudLoginErr},
		{RequireCloudLoginOrOnPremLogin, noContextCfg, requireCloudLoginOrOnPremErr},
		{RequireCloudLoginOrOnPremLogin, noContextCfg, requireCloudLoginOrOnPremErr},
		{RequireNonAPIKeyCloudLogin, apiKeyCloudCfg, requireNonAPIKeyCloudLoginErr},
		{RequireNonAPIKeyCloudLoginOrOnPremLogin, apiKeyCloudCfg, requireNonAPIKeyCloudLoginOrOnPremLoginErr},
		{RequireNonAPIKeyCloudLoginOrOnPremLogin, apiKeyCloudCfg, requireNonAPIKeyCloudLoginOrOnPremLoginErr},
		{RequireOnPremLogin, cloudCfg, requireOnPremLoginErr},
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
