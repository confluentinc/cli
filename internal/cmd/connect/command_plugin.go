package connect

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	opv1 "github.com/confluentinc/cc-structs/operator/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type pluginCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func newPluginCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "plugin",
		Short:       "Show plugins and their configurations.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &pluginCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newListCommand())

	return c.Command
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

func (c *pluginCommand) getPlugins() ([]*opv1.ConnectorPluginInfo, error) {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return nil, err
	}

	connector := &schedv1.Connector{
		AccountId:      c.EnvironmentId(),
		KafkaClusterId: kafkaCluster.ID,
	}

	return c.Client.Connect.GetPlugins(context.Background(), connector, "")
}
