package kafka

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *consumerCommand) newGroupDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <group>",
		Short:             "Describe a Kafka consumer group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validGroupArgs),
		RunE:              c.describe,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerCommand) describe(cmd *cobra.Command, args []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	group, err := kafkaREST.CloudClient.GetKafkaConsumerGroup(args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&consumerGroupOut{
		ClusterId:         group.GetClusterId(),
		ConsumerGroupId:   group.GetConsumerGroupId(),
		Coordinator:       getStringBroker(group.GetCoordinator().Related),
		IsSimple:          group.GetIsSimple(),
		PartitionAssignor: group.GetPartitionAssignor(),
		State:             group.GetState(),
	})
	return table.Print()
}

func getStringBroker(relationship string) string {
	// relationship will look like ".../v3/clusters/{cluster_id}/brokers/{broker_id}
	splitString := strings.SplitAfter(relationship, "brokers/")
	// if relationship was an empty string or did not contain "brokers/"
	if len(splitString) < 2 {
		return ""
	}
	// returning brokerId
	return splitString[1]
}
