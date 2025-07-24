package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

var availabilitiesToModel = map[string]string{
	"single-zone": "SINGLE_ZONE",
	"multi-zone":  "MULTI_ZONE",
	"low":         "LOW",
	"high":        "HIGH",
}

var availabilitiesToFreightModel = map[string]string{
	"single-zone": "LOW",  // TODO: This mapping is deprecated and will be removed in v5
	"multi-zone":  "HIGH", // TODO: This mapping is deprecated and will be removed in v5
	"low":         "LOW",
	"high":        "HIGH",
}

type clusterCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newClusterCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage Kafka clusters.",
	}

	c := &clusterCommand{}

	if cfg.IsCloudLogin() {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)
	} else {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
	}

	cmd.AddCommand(c.newConfigurationCommand(cfg, prerunner))
	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newUpdateCommand())
	cmd.AddCommand(c.newUseCommand())

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newListCommand())
		cmd.AddCommand(c.newEndpointCommand(cfg))
	} else {
		cmd.AddCommand(c.newListCommandOnPrem())
	}

	return cmd
}

func (c *clusterCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *clusterCommand) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	clusters, err := c.V2Client.ListKafkaClusters(environmentId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters))
	for i, cluster := range clusters {
		suggestions[i] = fmt.Sprintf("%s\t%s", cluster.GetId(), cluster.Spec.GetDisplayName())
	}
	return suggestions
}
