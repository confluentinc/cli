package kafka

import (
	"strings"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *consumerGroupCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <consumer-group>",
		Short:             "Describe a Kafka consumer group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerGroupCommand) describe(cmd *cobra.Command, args []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	consumerGroup, err := kafkaREST.CloudClient.GetKafkaConsumerGroup(cluster.ID, args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&consumerGroupOut{
		ClusterId:         consumerGroup.GetClusterId(),
		ConsumerGroupId:   consumerGroup.GetConsumerGroupId(),
		Coordinator:       getStringBroker(consumerGroup.GetCoordinator()),
		IsSimple:          consumerGroup.GetIsSimple(),
		PartitionAssignor: consumerGroup.GetPartitionAssignor(),
		State:             consumerGroup.GetState(),
	})
	return table.Print()
}

func getStringBroker(relationship kafkarestv3.Relationship) string {
	// relationship.Related will look like ".../v3/clusters/{cluster_id}/brokers/{broker_id}
	splitString := strings.SplitAfter(relationship.Related, "brokers/")
	// if relationship was an empty string or did not contain "brokers/"
	if len(splitString) < 2 {
		return ""
	}
	// returning brokerId
	return splitString[1]
}
