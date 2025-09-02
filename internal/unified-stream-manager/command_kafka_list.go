package unifiedstreammanager

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newKafkaListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka clusters.",
		Args:  cobra.NoArgs,
		RunE:  c.listKafka,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List Kafka clusters.",
				Code: "confluent unified-stream-manager kafka list",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) listKafka(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	clusters, err := c.V2Client.ListUsmKafkaClusters(environmentId)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, cluster := range clusters {
		out := &kafkaOut{
			Id:                              cluster.GetId(),
			Name:                            cluster.GetDisplayName(),
			ConfluentPlatformKafkaClusterId: cluster.GetConfluentPlatformKafkaClusterId(),
			Cloud:                           cluster.GetCloud(),
			Region:                          cluster.GetRegion(),
			Environment:                     cluster.Environment.GetId(),
		}

		list.Add(out)
	}

	return list.Print()
}
