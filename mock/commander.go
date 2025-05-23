package mock

import (
	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	servicequotav1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/version"
)

type Commander struct {
	Client            *ccloudv1.Client
	V2Client          *ccloudv2.Client
	MDSClient         *mdsv1.APIClient
	MDSv2Client       *mdsv2alpha1.APIClient
	KafkaRESTProvider *pcmd.KafkaRESTProvider
	QuotasClient      *servicequotav1.APIClient
	Version           *version.Version
	Config            *config.Config
}

var _ pcmd.PreRunner = (*Commander)(nil)

func NewPreRunnerMock(client *ccloudv1.Client, v2Client *ccloudv2.Client, mdsClient *mdsv1.APIClient, kafkaRESTProvider *pcmd.KafkaRESTProvider, cfg *config.Config) pcmd.PreRunner {
	return &Commander{
		Client:            client,
		V2Client:          v2Client,
		MDSClient:         mdsClient,
		KafkaRESTProvider: kafkaRESTProvider,
		Config:            cfg,
	}
}

func (c *Commander) Anonymous(command *pcmd.CLICommand, _ bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if command != nil {
			command.Version = c.Version
			command.Config = c.Config
		}
		return nil
	}
}

func (c *Commander) Authenticated(command *pcmd.AuthenticatedCLICommand) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := c.Anonymous(command.CLICommand, true)(cmd, args); err != nil {
			return err
		}
		c.setClient(command)

		ctx := command.Config.Context()
		if !ctx.HasLogin() {
			return new(errors.NotLoggedInError)
		}
		command.Context = ctx

		return nil
	}
}

func (c *Commander) AuthenticatedWithMDS(command *pcmd.AuthenticatedCLICommand) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := c.Anonymous(command.CLICommand, true)(cmd, args); err != nil {
			return err
		}
		c.setClient(command)

		ctx := command.Config.Context()
		if !ctx.HasLogin() {
			return new(errors.NotLoggedInError)
		}
		command.Context = ctx

		return nil
	}
}

// UseKafkaRest - The PreRun function registered by the mock prerunner for UseKafkaRestCLICommand
func (c *Commander) InitializeOnPremKafkaRest(command *pcmd.AuthenticatedCLICommand) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := c.AuthenticatedWithMDS(command)(cmd, args); err != nil {
			return err
		}
		command.KafkaRESTProvider = c.KafkaRESTProvider
		return nil
	}
}

func (c *Commander) ParseFlagsIntoContext(command *pcmd.CLICommand) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		return command.Config.Context().ParseFlagsIntoContext(cmd)
	}
}

func (c *Commander) setClient(command *pcmd.AuthenticatedCLICommand) {
	command.Client = c.Client
	command.V2Client = c.V2Client
	//command.MDSClient = c.MDSClient
	command.MDSv2Client = c.MDSv2Client
	command.KafkaRESTProvider = c.KafkaRESTProvider
}
