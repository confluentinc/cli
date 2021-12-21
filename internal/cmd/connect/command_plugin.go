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
	completableChildren []*cobra.Command
}

func newPluginCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "plugin",
		Short:       "Show plugins and their configurations.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &pluginCommand{AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	describeCmd := c.newDescribeCommand()

	c.AddCommand(describeCmd)
	c.AddCommand(c.newListCommand())

	c.completableChildren = []*cobra.Command{describeCmd}

	return c.Command
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
