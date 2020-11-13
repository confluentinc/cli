package mock

import (
	"os"

	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"

	"github.com/confluentinc/ccloud-sdk-go"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	pmock "github.com/confluentinc/cli/internal/pkg/mock"
	"github.com/confluentinc/cli/internal/pkg/version"
)

type Commander struct {
	FlagResolver pcmd.FlagResolver
	Client       *ccloud.Client
	MDSClient    *mds.APIClient
	MDSv2Client  *mdsv2alpha1.APIClient
	Version      *version.Version
	Config       *v3.Config
}

var _ pcmd.PreRunner = (*Commander)(nil)

func NewPreRunnerMock(client *ccloud.Client, mdsClient *mds.APIClient, cfg *v3.Config) pcmd.PreRunner {
	flagResolverMock := &pcmd.FlagResolverImpl{
		Prompt: &pmock.Prompt{},
		Out:    os.Stdout,
	}
	return &Commander{
		FlagResolver: flagResolverMock,
		Client:       client,
		MDSClient:    mdsClient,
		Config:       cfg,
	}
}

func NewPreRunnerMdsV2Mock(client *ccloud.Client, mdsClient *mdsv2alpha1.APIClient, cfg *v3.Config) pcmd.PreRunner {
	flagResolverMock := &pcmd.FlagResolverImpl{
		Prompt: &pmock.Prompt{},
		Out:    os.Stdout,
	}
	return &Commander{
		FlagResolver: flagResolverMock,
		Client:       client,
		MDSv2Client:  mdsClient,
		Config:       cfg,
	}
}

func (c *Commander) Anonymous(command *pcmd.CLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if command != nil {
			command.Version = c.Version
			command.Config.Resolver = c.FlagResolver
			command.Config.Config = c.Config
		}
		return nil
	}
}

func (c *Commander) Authenticated(command *pcmd.AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error {
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

func (c *Commander) AuthenticatedWithMDS(command *pcmd.AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error {
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

func (c *Commander) HasAPIKey(command *pcmd.HasAPIKeyCLICommand) func(cmd *cobra.Command, args []string) error {
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

func (c *Commander) setClient(command *pcmd.AuthenticatedCLICommand) {
	command.Client = c.Client
	command.MDSClient = c.MDSClient
	command.MDSv2Client = c.MDSv2Client
	command.Config.Client = c.Client
}
