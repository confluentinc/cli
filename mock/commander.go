package mock

import (
	"os"

	"github.com/confluentinc/ccloud-sdk-go"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/version"
)

type Commander struct {
	FlagResolver cmd.FlagResolver
	Client       *ccloud.Client
	MDSClient    *mds.APIClient
	Version      *version.Version
	Config       *v3.Config
}

var _ cmd.PreRunner = (*Commander)(nil)

func NewPreRunnerMock(client *ccloud.Client, mdsClient *mds.APIClient, cfg *v3.Config) cmd.PreRunner {
	flagResolverMock := &cmd.FlagResolverImpl{
		Prompt: &Prompt{},
		Out:    os.Stdout,
	}
	return &Commander{
		FlagResolver: flagResolverMock,
		Client:       client,
		MDSClient:    mdsClient,
		Config:       cfg,
	}
}

func (c *Commander) Anonymous(command *cmd.CLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if command != nil {
			command.Version = c.Version
			command.Config.Resolver = c.FlagResolver
			command.Config.Config = c.Config
		}
		return nil
	}
}

func (c *Commander) Authenticated(command *cmd.AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := c.Anonymous(command.CLICommand)(cmd, args)
		if err != nil {
			return err
		}
		c.setClient(command)
		ctx, err := command.Config.Context(cmd)
		if err != nil {
			return err
		}
		if ctx == nil {
			return &errors.NoContextError{CLIName: c.Config.CLIName}
		}
		command.Context = ctx
		command.State, err = ctx.AuthenticatedState(cmd)
		if err == nil {
			return err
		}
		return nil
	}
}

func (c *Commander) AuthenticatedWithMDS(command *cmd.AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := c.Anonymous(command.CLICommand)(cmd, args)
		if err != nil {
			return err
		}
		c.setClient(command)
		ctx, err := command.Config.Context(cmd)
		if err != nil {
			return err
		}
		if ctx == nil {
			return &errors.NoContextError{CLIName: c.Config.CLIName}
		}
		command.Context = ctx
		if !ctx.HasMDSLogin() {
			return &errors.NotLoggedInError{CLIName: c.Config.CLIName}
		}
		command.State = ctx.State
		return nil
	}
}

func (c *Commander) HasAPIKey(command *cmd.HasAPIKeyCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := c.Anonymous(command.CLICommand)(cmd, args)
		if err != nil {
			return err
		}
		ctx, err := command.Config.Context(cmd)
		if err != nil {
			return err
		}
		if ctx == nil {
			return &errors.NoContextError{CLIName: c.Config.CLIName}
		}
		command.Context = ctx
		return nil
	}
}

func (c *Commander) setClient(command *cmd.AuthenticatedCLICommand) {
	command.Client = c.Client
	command.MDSClient = c.MDSClient
	command.Config.Client = c.Client
}
