package internal

import (
	"runtime"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	pversion "github.com/confluentinc/cli/v4/pkg/version"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

var (
	regularOrgContextState   = &config.ContextState{Auth: &config.AuthConfig{Organization: testserver.RegularOrg}}
	suspendedOrgContextState = func(eventType ccloudv1.SuspensionEventType) *config.ContextState {
		return &config.ContextState{Auth: &config.AuthConfig{Organization: testserver.SuspendedOrg(eventType)}}
	}
)

func TestHelp_NoContext(t *testing.T) {
	cfg := new(config.Config)

	out, err := runWithConfig(cfg)
	require.NoError(t, err)

	commands := []string{
		"cloud-signup", "completion", "context", "help", "kafka", "local", "login", "logout", "secret", "update",
		"version",
	}
	if runtime.GOOS == "windows" {
		commands = slices.DeleteFunc(commands, func(s string) bool { return s == "local" })
	}

	for _, command := range commands {
		require.Contains(t, out, command)
	}
}

func TestHelp_CloudSuspendedOrg(t *testing.T) {
	cfg := &config.Config{
		Contexts: map[string]*config.Context{"cloud": {
			PlatformName: "confluent.cloud",
			State:        suspendedOrgContextState(ccloudv1.SuspensionEventType_SUSPENSION_EVENT_CUSTOMER_INITIATED_ORG_DEACTIVATION),
		}},
		CurrentContext: "cloud",
	}

	out, err := runWithConfig(cfg)
	require.NoError(t, err)

	commands := []string{
		"cloud-signup", "completion", "context", "help", "kafka", "local", "login", "logout", "prompt", "shell", "update", "version",
	}
	if runtime.GOOS == "windows" {
		commands = slices.DeleteFunc(commands, func(s string) bool { return s == "local" })
	}

	for _, command := range commands {
		require.Contains(t, out, command)
	}
}

func TestHelp_CloudEndOfFreeTrialSuspendedOrg(t *testing.T) {
	cfg := &config.Config{
		Contexts: map[string]*config.Context{"cloud": {
			PlatformName: "confluent.cloud",
			State:        suspendedOrgContextState(ccloudv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL),
		}},
		CurrentContext: "cloud",
	}

	out, err := runWithConfig(cfg)
	require.NoError(t, err)

	// Note that users can still run `confluent billing payment update` or `confluent billing promo add` if their organization is suspended
	// but only due to end of free trial
	commands := []string{
		"billing", "cloud-signup", "completion", "context", "help", "kafka", "local", "login", "logout", "prompt", "shell", "update", "version",
	}
	if runtime.GOOS == "windows" {
		commands = slices.DeleteFunc(commands, func(s string) bool { return s == "local" })
	}

	for _, command := range commands {
		require.Contains(t, out, command)
	}

	// check that some top level cloud commands are not included (each of these top level command corresponds to a
	// different run requirement)
	cloudCommands := []string{"api-key", "audit-log", "cluster", "connect", "service-quota"}
	for _, command := range cloudCommands {
		require.NotContains(t, out, command)
	}

	cmd := NewConfluentCommand(cfg)

	out, err = pcmd.ExecuteCommand(cmd, "billing", "payment", "--help")
	require.NoError(t, err)
	require.Contains(t, out, "update")
	require.Contains(t, out, "describe")

	out, err = pcmd.ExecuteCommand(cmd, "billing", "promo", "--help")
	require.NoError(t, err)
	require.Contains(t, out, "add")
	require.Contains(t, out, "list")

	out, err = pcmd.ExecuteCommand(cmd, "billing", "--help")
	require.NoError(t, err)
	require.NotContains(t, out, "cost")
	require.NotContains(t, out, "price")
}

func TestHelp_Cloud(t *testing.T) {
	cfg := &config.Config{
		Contexts: map[string]*config.Context{"cloud": {
			PlatformName: "confluent.cloud",
			State:        regularOrgContextState,
		}},
		CurrentContext: "cloud",
	}

	out, err := runWithConfig(cfg)
	require.NoError(t, err)

	commands := []string{
		"ai", "api-key", "audit-log", "billing", "cloud-signup", "completion", "context", "connect", "environment", "help",
		"iam", "kafka", "ksql", "login", "logout", "prompt", "schema-registry", "shell", "update", "version",
	}

	for _, command := range commands {
		require.Contains(t, out, command)
	}

	cmd := NewConfluentCommand(cfg)

	out, err = pcmd.ExecuteCommand(cmd, "billing", "--help")
	require.NoError(t, err)
	require.Contains(t, out, "cost")
	require.Contains(t, out, "payment")
	require.Contains(t, out, "price")
	require.Contains(t, out, "promo")
}

func TestHelp_CloudWithAPIKey(t *testing.T) {
	cfg := &config.Config{
		Contexts: map[string]*config.Context{
			"cloud-with-api-key": {
				PlatformName: "confluent.cloud",
				Credential:   &config.Credential{CredentialType: config.APIKey},
				State:        regularOrgContextState,
			},
		},
		CurrentContext: "cloud-with-api-key",
	}

	out, err := runWithConfig(cfg)
	require.NoError(t, err)

	commands := []string{
		"audit-log", "billing", "cloud-signup", "completion", "context", "help", "kafka", "login", "logout", "update",
		"version",
	}

	for _, command := range commands {
		require.Contains(t, out, command)
	}

	cmd := NewConfluentCommand(cfg)

	out, err = pcmd.ExecuteCommand(cmd, "billing", "--help")
	require.NoError(t, err)
	require.Contains(t, out, "payment")
	require.Contains(t, out, "promo")
	require.NotContains(t, out, "cost")
	require.NotContains(t, out, "price")
}

func TestHelp_OnPrem(t *testing.T) {
	cfg := &config.Config{
		Contexts:       map[string]*config.Context{"on-prem": {PlatformName: "https://example.com"}},
		CurrentContext: "on-prem",
	}

	out, err := runWithConfig(cfg)
	require.NoError(t, err)

	commands := []string{
		"audit-log", "cloud-signup", "cluster", "completion", "context", "connect", "help", "iam", "kafka", "ksql",
		"local", "login", "logout", "schema-registry", "secret", "update", "version",
	}
	if runtime.GOOS == "windows" {
		commands = slices.DeleteFunc(commands, func(s string) bool { return s == "local" })
	}

	for _, command := range commands {
		require.Contains(t, out, command)
	}
}

func runWithConfig(cfg *config.Config) (string, error) {
	cfg.IsTest = true
	cfg.Version = new(pversion.Version)

	cmd := NewConfluentCommand(cfg)
	return pcmd.ExecuteCommand(cmd, "help")
}
