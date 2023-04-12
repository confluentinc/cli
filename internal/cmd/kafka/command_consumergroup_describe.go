package kafka

import (
	"strings"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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
				Text: `Describe the "my-consumer-group" consumer group:`,
				Code: "confluent kafka consumer-group describe my-consumer-group",
			},
		),
		Hidden: true,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerGroupCommand) describe(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]

	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(c.AuthenticatedCLICommand)
	if err != nil {
		return err
	}

	groupCmdResp, httpResp, err := kafkaREST.CloudClient.GetKafkaConsumerGroup(lkc, consumerGroupId)
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	groupCmdConsumersResp, httpResp, err := kafkaREST.CloudClient.ListKafkaConsumers(lkc, consumerGroupId)
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	groupData := getGroupData(groupCmdResp, groupCmdConsumersResp)

	table := output.NewTable(cmd)
	table.Add(convertGroupToDescribeStruct(groupData))
	if err := table.Print(); err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		output.Println()
		output.Println("Consumers")
		output.Println()

		list := output.NewList(cmd)
		for _, consumer := range groupData.Consumers {
			list.Add(&consumer)
		}
		return list.Print()
	}

	return nil
}

func getGroupData(groupCmdResp kafkarestv3.ConsumerGroupData, groupCmdConsumersResp kafkarestv3.ConsumerDataList) *groupData {
	groupData := &groupData{
		ClusterId:         groupCmdResp.ClusterId,
		ConsumerGroupId:   groupCmdResp.ConsumerGroupId,
		Coordinator:       getStringBroker(groupCmdResp.Coordinator),
		IsSimple:          groupCmdResp.IsSimple,
		PartitionAssignor: groupCmdResp.PartitionAssignor,
		State:             groupCmdResp.State,
		Consumers:         make([]consumerData, len(groupCmdConsumersResp.Data)),
	}

	// Populate consumers list
	for i, consumerResp := range groupCmdConsumersResp.Data {
		groupData.Consumers[i] = consumerData{
			ConsumerGroupId: groupCmdResp.ConsumerGroupId,
			ConsumerId:      consumerResp.ConsumerId,
			InstanceId:      consumerResp.GetInstanceId(),
			ClientId:        consumerResp.ClientId,
		}
	}

	return groupData
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

func convertGroupToDescribeStruct(groupData *groupData) *consumerGroupOut {
	return &consumerGroupOut{
		ClusterId:         groupData.ClusterId,
		ConsumerGroupId:   groupData.ConsumerGroupId,
		Coordinator:       groupData.Coordinator,
		IsSimple:          groupData.IsSimple,
		PartitionAssignor: groupData.PartitionAssignor,
		State:             groupData.State,
	}
}
