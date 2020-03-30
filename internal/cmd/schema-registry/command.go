package schema_registry

import (
	"github.com/spf13/cobra"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/log"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	logger    *log.Logger
	srClient  *srsdk.APIClient
	prerunner pcmd.PreRunner
}

func New(prerunner pcmd.PreRunner, config *v3.Config, srClient *srsdk.APIClient, logger *log.Logger) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "schema-registry",
			Short: `Manage Schema Registry.`,
		},
		config, prerunner)
	cmd := &command{
		AuthenticatedCLICommand: cliCmd,
		srClient:                srClient,
		logger:                  logger,
		prerunner:               prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewClusterCommand(c.Config.Config, c.prerunner, c.srClient, c.logger))
	c.AddCommand(NewSubjectCommand(c.Config.Config, c.prerunner, c.srClient))
	c.AddCommand(NewSchemaCommand(c.Config.Config, c.prerunner, c.srClient))
}
