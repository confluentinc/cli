package kafka

import (
	"strings"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *consumerGroupCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <consumer-group>",
		Short:             "Describe a Kafka consumer group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "my-consumer-group" consumer group.`,
				Code: "confluent kafka consumer-group describe my-consumer-group",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerGroupCommand) describe(cmd *cobra.Command, args []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	consumerGroupData, httpResp, err := kafkaREST.CloudClient.GetKafkaConsumerGroup(cluster.ID, args[0])
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	table := output.NewTable(cmd)
	table.Add(&consumerGroupOut{
		ClusterId:         consumerGroupData.GetClusterId(),
		ConsumerGroupId:   consumerGroupData.GetConsumerGroupId(),
		Coordinator:       getStringBroker(consumerGroupData.GetCoordinator()),
		IsSimple:          consumerGroupData.GetIsSimple(),
		PartitionAssignor: consumerGroupData.GetPartitionAssignor(),
		State:             consumerGroupData.GetState(),
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
