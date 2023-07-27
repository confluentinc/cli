package kafka

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *consumerGroupCommand) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <consumer-group>",
		Short: "Describe a Kafka consumer group.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describeOnPrem,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerGroupCommand) describeOnPrem(cmd *cobra.Command, args []string) error {
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	consumerGroup, resp, err := restClient.ConsumerGroupV3Api.GetKafkaConsumerGroup(restContext, clusterId, args[0])
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	table := output.NewTable(cmd)
	table.Add(&consumerGroupOut{
		ClusterId:         consumerGroup.ClusterId,
		ConsumerGroupId:   consumerGroup.ConsumerGroupId,
		Coordinator:       getStringBrokerOnPrem(consumerGroup.Coordinator),
		IsSimple:          consumerGroup.IsSimple,
		PartitionAssignor: consumerGroup.PartitionAssignor,
		State:             consumerGroup.State,
	})
	return table.Print()
}

func getStringBrokerOnPrem(relationship kafkarestv3.Relationship) string {
	// relationship.Related will look like ".../v3/clusters/{cluster_id}/brokers/{broker_id}
	splitString := strings.SplitAfter(relationship.Related, "brokers/")
	// if relationship was an empty string or did not contain "brokers/"
	if len(splitString) < 2 {
		return ""
	}
	// returning brokerId
	return splitString[1]
}
