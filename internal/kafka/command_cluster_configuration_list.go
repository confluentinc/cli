package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *clusterCommand) newConfigurationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List updated Kafka cluster configurations.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.configurationList,
	}

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) configurationList(cmd *cobra.Command, _ []string) error {
	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	configs, err := kafkaREST.CloudClient.ListKafkaClusterConfigs()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, config := range configs {
		list.Add(&configurationOut{
			Name:     config.GetName(),
			Value:    config.GetValue(),
			ReadOnly: config.GetIsReadOnly(),
		})
	}
	return list.Print()
}
