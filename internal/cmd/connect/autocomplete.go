package connect

import (
	"context"
	"fmt"

	"github.com/c-bata/go-prompt"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	opv1 "github.com/confluentinc/cc-structs/operator/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/shell/completer"
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

func (c *command) Cmd() *cobra.Command {
	return c.Command
}

func (c *command) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}

func (c *command) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	connectors, err := c.fetchConnectors()
	if err != nil {
		return suggestions
	}
	for _, conn := range connectors {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        conn.Id.Id,
			Description: conn.Info.Name,
		})
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

func (c *command) ServerCompletableFlagChildren() map[string][]*cobra.Command {
	return c.completableFlagChildren
}

func (c *command) ServerFlagComplete() map[string]func() []prompt.Suggest {
	return map[string]func() []prompt.Suggest{
		"cluster": completer.ClusterFlagServerCompleterFunc(c.Client, c.EnvironmentId()),
	}
}

func (c *pluginCommand) Cmd() *cobra.Command {
	return c.Command
}

func (c *pluginCommand) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	plugins, err := c.getPlugins()
	if err != nil {
		return suggestions
	}
	for _, conn := range plugins {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        conn.Class,
			Description: conn.Type,
		})
	}
	return suggestions
}

func (c *pluginCommand) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}
