package mock

import (
	"os"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/mds-sdk-go"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/version"
)

type Commander struct {
	FlagResolver cmd.FlagResolver
	Client       *ccloud.Client
	MDSClient    *mds.APIClient
	Version *version.Version
}

var _ cmd.PreRunner = (*Commander)(nil)

func NewPreRunnerMock(client *ccloud.Client, mdsClient *mds.APIClient) cmd.PreRunner {
	flagResolverMock := &cmd.FlagResolverImpl{
		Prompt: &Prompt{},
		Out:    os.Stdout,
	}
	return &Commander{
		FlagResolver: flagResolverMock,
		Client:       client,
		MDSClient:    mdsClient,
	}
}

func (c *Commander) Anonymous(command *cmd.CLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if command != nil {
			c.setClient(command)
			command.Version = c.Version
			command.Config.Resolver = c.FlagResolver
		}
		return nil
	}
}

func (c *Commander) Authenticated(command *cmd.AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := c.Anonymous(command.CLICommand)(cmd, args)
		panicIfErr(err)
		c.setClient(command.CLICommand)
		ctx, err := command.Config.Context(cmd)
		panicIfErr(err)
		command.Context = ctx
		command.State, err = ctx.AuthenticatedState(cmd)
		panicIfErr(err)
		return nil
	}
}

func (c *Commander) HasAPIKey(command *cmd.HasAPIKeyCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := c.Anonymous(command.CLICommand)(cmd, args)
		panicIfErr(err)
		c.setClient(command.CLICommand)
		ctx, err := command.Config.Context(cmd)
		panicIfErr(err)
		command.Context = ctx 
		return nil
	}
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (c *Commander) setClient(command *cmd.CLICommand) {
	command.Client = c.Client
	command.MDSClient = c.MDSClient
	command.Config.Client = c.Client
}
