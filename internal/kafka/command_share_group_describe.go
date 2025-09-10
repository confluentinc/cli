package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *shareCommand) newGroupDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <group>",
		Short:             "Describe a Kafka share group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validGroupArgs),
		RunE:              c.groupDescribe,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *shareCommand) groupDescribe(cmd *cobra.Command, args []string) error {
	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	shareGroup, err := kafkaREST.CloudClient.GetKafkaShareGroup(args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&shareGroupOut{
		Cluster:                 shareGroup.GetClusterId(),
		ShareGroup:              shareGroup.GetShareGroupId(),
		Coordinator:             getStringBrokerFromShareGroup(shareGroup),
		State:                   shareGroup.GetState(),
		ConsumerCount:           shareGroup.GetConsumerCount(),
		PartitionCount:          shareGroup.GetPartitionCount(),
		TopicSubscriptions: formatAssignedTopicPartitions(shareGroup.GetAssignedTopicPartitions()),
	})
	return table.Print()
}
