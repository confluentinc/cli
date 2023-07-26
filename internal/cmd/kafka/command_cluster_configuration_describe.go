package kafka

import (
	"fmt"
	"strings"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (c *clusterCommand) newConfigurationDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Kafka cluster configuration.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.configurationDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe Kafka cluster configuration "auto.create.topics.enable"`,
				Code: "confluent kafka cluster configuration describe auto.create.topics.enable",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) configurationDescribe(cmd *cobra.Command, args []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	config, err := kafkaREST.CloudClient.GetKafkaClusterConfig(cluster.ID, args[0])
	if err != nil {
		return catchConfigurationNotFound(err, args[0])
	}

	table := output.NewTable(cmd)
	table.Add(&configurationOut{
		Name:     config.GetName(),
		Value:    config.GetValue(),
		ReadOnly: config.GetIsReadOnly(),
	})
	return table.Print()
}

func catchConfigurationNotFound(err error, configuration string) error {
	if err != nil && strings.Contains(err.Error(), "Not Found") {
		return fmt.Errorf(`configuration "%s" not found`, configuration)
	}
	return err
}
