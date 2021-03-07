package schema_registry

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/log"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	logger          *log.Logger
	srClient        *srsdk.APIClient
	prerunner       pcmd.PreRunner
	analyticsClient analytics.Client
}

func New(cliName string, prerunner pcmd.PreRunner, srClient *srsdk.APIClient, logger *log.Logger, analyticsClient analytics.Client) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "schema-registry",
			Short: `Manage Schema Registry.`,
		}, prerunner)
	cmd := &command{
		AuthenticatedCLICommand: cliCmd,
		srClient:                srClient,
		logger:                  logger,
		prerunner:               prerunner,
		analyticsClient:         analyticsClient,
	}
	cmd.init(cliName)
	return cmd.Command
}

func (c *command) init(cliName string) {
	if cliName == "ccloud" {
		c.AddCommand(NewClusterCommand(cliName, c.prerunner, c.srClient, c.logger, c.analyticsClient))
		c.AddCommand(NewSubjectCommand(cliName, c.prerunner, c.srClient))
		c.AddCommand(NewSchemaCommand(cliName, c.prerunner, c.srClient))
	} else {
		c.AddCommand(NewClusterCommandOnPrem(c.prerunner))
	}
}
