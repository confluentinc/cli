package mock

import (
	"os"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/mds-sdk-go"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type Commander struct {
	FlagResolver cmd.FlagResolver
	Client       *ccloud.Client
	MDSClient    *mds.APIClient
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

func (c *Commander) Anonymous(cfg *config.Config, command *cmd.CLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := c.FlagResolver.ResolveContextFlag(cmd, cfg)
		if command != nil {
			c.setClient(command)
		}
		panicIfErr(err)
		return nil
	}
}

func (c *Commander) Authenticated(cfg *config.Config, command *cmd.CLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := c.Anonymous(cfg, command)(cmd, args)
		panicIfErr(err)
		err = c.FlagResolver.ResolveFlags(cmd, cfg, c.Client)
		panicIfErr(err)
		c.setClient(command)
		return nil
	}
}

func (c *Commander) HasAPIKey(cfg *config.Config, command *cmd.CLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := c.Anonymous(cfg, command)(cmd, args)
		panicIfErr(err)
		err = c.FlagResolver.ResolveFlags(cmd, cfg, c.Client)
		panicIfErr(err)
		c.setClient(command)
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
}
