package cmd

import (
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/mds-sdk-go"
	"github.com/spf13/cobra"

	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/version"
)

type CLICommand struct {
	*cobra.Command
	Config    *DynamicConfig
	Version   *version.Version
	Prerunner PreRunner
}
type AuthenticatedCLICommand struct {
	*CLICommand
	Client    *ccloud.Client
	MDSClient *mds.APIClient
	Context   *DynamicContext
	State     *v2.ContextState
}

type HasAPIKeyCLICommand struct {
	*CLICommand
	Context *DynamicContext
}

func (a *AuthenticatedCLICommand) AuthToken() string {
	return a.State.AuthToken
}
func (a *AuthenticatedCLICommand) EnvironmentId() string {
	return a.State.Auth.Account.Id
}

func NewAuthenticatedCLICommand(command *cobra.Command, cfg *v2.Config, prerunner PreRunner) *AuthenticatedCLICommand {
	cmd := &AuthenticatedCLICommand{
		CLICommand: NewCLICommand(command, cfg, prerunner),
		Context:    nil,
		State:      nil,
	}
	command.PersistentPreRunE = prerunner.Authenticated(cmd)
	cmd.Command = command
	return cmd
}

func NewAuthenticatedWithMDSCLICommand(command *cobra.Command, cfg *v2.Config, prerunner PreRunner) *AuthenticatedCLICommand {
	cmd := &AuthenticatedCLICommand{
		CLICommand: NewCLICommand(command, cfg, prerunner),
		Context:    nil,
		State:      nil,
	}
	command.PersistentPreRunE = prerunner.AuthenticatedWithMDS(cmd)
	cmd.Command = command
	return cmd
}

func NewHasAPIKeyCLICommand(command *cobra.Command, cfg *v2.Config, prerunner PreRunner) *HasAPIKeyCLICommand {
	cmd := &HasAPIKeyCLICommand{
		CLICommand: NewCLICommand(command, cfg, prerunner),
		Context:    nil,
	}
	command.PersistentPreRunE = prerunner.HasAPIKey(cmd)
	cmd.Command = command
	return cmd
}

func NewAnonymousCLICommand(command *cobra.Command, cfg *v2.Config, prerunner PreRunner) *CLICommand {
	cmd := NewCLICommand(command, cfg, prerunner)
	command.PersistentPreRunE = prerunner.Anonymous(cmd)
	cmd.Command = command
	return cmd
}

func NewCLICommand(command *cobra.Command, cfg *v2.Config, prerunner PreRunner) *CLICommand {
	return &CLICommand{
		Config:    NewDynamicConfig(cfg, nil, nil),
		Command:   command,
		Prerunner: prerunner,
	}
}

func (a *AuthenticatedCLICommand) AddCommand(command *cobra.Command) {
	command.PersistentPreRunE = a.PersistentPreRunE
	a.Command.AddCommand(command)
}

func (h *HasAPIKeyCLICommand) AddCommand(command *cobra.Command) {
	command.PersistentPreRunE = h.PersistentPreRunE
	h.Command.AddCommand(command)
}
