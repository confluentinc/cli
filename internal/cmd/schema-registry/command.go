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
	cmd := &cobra.Command{
		Use:         "schema-registry",
		Short:       "Manage Schema Registry.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
		srClient:                srClient,
		logger:                  logger,
		prerunner:               prerunner,
		analyticsClient:         analyticsClient,
	}

	c.AddCommand(newClusterCommand(cfg, c.prerunner, c.srClient, c.logger, c.analyticsClient))
	c.AddCommand(newExporterCommand(c.prerunner, c.srClient))
	c.AddCommand(newSchemaCommand(c.prerunner, c.srClient))
	c.AddCommand(newSubjectCommand(c.prerunner, c.srClient))

	return c.Command
}
