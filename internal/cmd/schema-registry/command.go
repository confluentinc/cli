package schemaregistry

import (
	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
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

func New(cfg *v1.Config, prerunner pcmd.PreRunner, srClient *srsdk.APIClient, logger *log.Logger, analyticsClient analytics.Client) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:         "schema-registry",
			Short:       `Manage Schema Registry.`,
			Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
		}, prerunner)
	cmd := &command{
		AuthenticatedCLICommand: cliCmd,
		srClient:                srClient,
		logger:                  logger,
		prerunner:               prerunner,
		analyticsClient:         analyticsClient,
	}
	cmd.init(cfg)
	return cmd.Command
}

func (c *command) init(cfg *v1.Config) {
	c.AddCommand(NewClusterCommand(cfg, c.prerunner, c.srClient, c.logger, c.analyticsClient))
	c.AddCommand(NewExporterCommand(c.prerunner, c.srClient))
	c.AddCommand(NewSchemaCommand(c.prerunner, c.srClient))
	c.AddCommand(NewSubjectCommand(c.prerunner, c.srClient))
}
