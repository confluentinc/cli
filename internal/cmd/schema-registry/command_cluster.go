package schemaregistry

import (
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type clusterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	logger          *log.Logger
	srClient        *srsdk.APIClient
	analyticsClient analytics.Client
}

func newClusterCommand(cfg *v1.Config, prerunner pcmd.PreRunner, srClient *srsdk.APIClient, logger *log.Logger, analyticsClient analytics.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "cluster",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	if cfg.IsCloudLogin() {
		cmd.Short = "Manage Schema Registry cluster."
		cmd.Long = "Manage the Schema Registry cluster for the current environment."
	} else {
		cmd.Short = "Manage Schema Registry clusters."
	}

	c := &clusterCommand{
		srClient:        srClient,
		logger:          logger,
		analyticsClient: analyticsClient,
	}

	if cfg.IsCloudLogin() {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner, nil)
	} else {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner, nil)
	}

	c.AddCommand(c.newDescribeCommand(cfg))
	c.AddCommand(c.newEnableCommand(cfg))
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newUpdateCommand(cfg))

	return c.Command
}
