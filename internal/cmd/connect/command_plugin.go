package connect

import (
	"github.com/spf13/cobra"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type pluginCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newPluginCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage plugins for managed connectors.",
	}

	c := &pluginCommand{}

	if cfg.IsCloudLogin() {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)

		cmd.AddCommand(c.newListCommand())
	} else {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)

		cmd.AddCommand(c.newInstallCommand())
	}

	cmd.AddCommand(c.newDescribeCommand())

	return cmd
}

func (c *pluginCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteConnectorPlugins()
}

func (c *pluginCommand) autocompleteConnectorPlugins() []string {
	plugins, err := c.getPlugins()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(plugins))
	for i, plugin := range plugins {
		suggestions[i] = plugin.Class
	}
	return suggestions
}

func (c *pluginCommand) getPlugins() ([]connectv1.InlineResponse2002, error) {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return nil, err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListConnectorPlugins(environmentId, kafkaCluster.ID)
}
