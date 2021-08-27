package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/config"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	testserver "github.com/confluentinc/cli/test/test-server"
)

var (
	noContextCfg = new(v3.Config)

	cloudCfg = &v3.Config{
		Contexts:       map[string]*v3.Context{"cloud": {PlatformName: testserver.TestCloudURL.String()}},
		CurrentContext: "cloud",
		IsTest:         true,
	}

	apiKeyCloudCfg = &v3.Config{
		BaseConfig: &config.BaseConfig{Params: &config.Params{CLIName: "ccloud"}}, // TODO: Remove CLIName
		Contexts: map[string]*v3.Context{"cloud": {
			PlatformName: testserver.TestCloudURL.String(),
			Credential:   &v2.Credential{CredentialType: v2.APIKey},
		}},
		CurrentContext: "cloud",
		IsTest:         true,
	}

	nonAPIKeyCloudCfg = &v3.Config{
		BaseConfig: &config.BaseConfig{Params: &config.Params{CLIName: "ccloud"}}, // TODO: Remove CLIName
		Contexts: map[string]*v3.Context{"cloud": {
			PlatformName: testserver.TestCloudURL.String(),
			Credential:   &v2.Credential{CredentialType: v2.Username},
		}},
		CurrentContext: "cloud",
		IsTest:         true,
	}

	onPremCfg = &v3.Config{
		BaseConfig: &config.BaseConfig{Params: &config.Params{CLIName: "confluent"}}, // TODO: Remove CLIName
		Contexts: map[string]*v3.Context{"on-prem": {
			Credential:   new(v2.Credential),
			PlatformName: "https://example.com",
			State:        &v2.ContextState{AuthToken: "token"},
		}},
		CurrentContext: "on-prem",
	}

	updatesDisabledCfg = &v3.Config{DisableUpdates: true}

	updatesEnabledCfg = &v3.Config{DisableUpdates: false}
)

func TestErrIfMissingRunRequirement_NoError(t *testing.T) {
	for _, test := range []struct {
		req string
		cfg *v3.Config
	}{
		{RequireCloudLogin, cloudCfg},
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
		cfg *v3.Config
		err error
	}{
		{RequireCloudLogin, onPremCfg, requireCloudLoginErr},
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
