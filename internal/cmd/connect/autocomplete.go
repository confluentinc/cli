package connect

import (
	"context"
	"fmt"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	opv1 "github.com/confluentinc/cc-structs/operator/v1"
	"github.com/spf13/cobra"
)

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteConnectors()
}

func (c *command) autocompleteConnectors() []string {
	connectors, err := c.fetchConnectors()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(connectors))
	i := 0
	for _, connector := range connectors {
		suggestions[i] = fmt.Sprintf("%s\t%s", connector.Id.Id, connector.Info.Name)
		i++
	}
	return suggestions
}

func (c *command) fetchConnectors() (map[string]*opv1.ConnectorExpansion, error) {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return nil, err
	}

	connector := &schedv1.Connector{
		AccountId:      c.EnvironmentId(),
		KafkaClusterId: kafkaCluster.ID,
	}

	return c.Client.Connect.ListWithExpansions(context.Background(), connector, "status,info,id")
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

func (c *pluginCommand) Cmd() *cobra.Command {
	return c.Command
}
