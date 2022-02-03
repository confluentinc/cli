package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

const (
	singleZone = "single-zone"
	multiZone  = "multi-zone"
)

type clusterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	analyticsClient analytics.Client
}

func newClusterCommand(cfg *v1.Config, prerunner pcmd.PreRunner, analyticsClient analytics.Client) *clusterCommand {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Manage Kafka clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &clusterCommand{analyticsClient: analyticsClient}

	if cfg.IsCloudLogin() {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)
	} else {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)
	}

	c.AddCommand(c.newCreateCommand(cfg))
	c.AddCommand(c.newDeleteCommand(cfg))
	c.AddCommand(c.newDescribeCommand(cfg))
	c.AddCommand(c.newUpdateCommand(cfg))
	c.AddCommand(c.newUseCommand(cfg))

	if cfg.IsCloudLogin() {
		c.AddCommand(c.newListCommand())
	} else {
		c.AddCommand(c.newListCommandOnPrem())
	}

	return c
}

func (c *clusterCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteClusters(c.EnvironmentId(), c.Client)
}
