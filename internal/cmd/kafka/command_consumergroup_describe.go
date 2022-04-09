package kafka

import (
	cloudkafkarest "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"strings"

	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/go-printer/tables"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	groupDescribeFields              = []string{"ClusterId", "ConsumerGroupId", "Coordinator", "IsSimple", "PartitionAssignor", "State"}
	groupDescribeHumanRenames        = map[string]string{"ClusterId": "Cluster", "ConsumerGroupId": "Consumer Group", "IsSimple": "Simple"}
	groupDescribeConsumersFields     = []string{"ConsumerGroupId", "ConsumerId", "InstanceId", "ClientId"}
	groupDescribeConsumerTableLabels = []string{"Consumer Group", "Consumer", "Instance", "Client"}
)

func (c *consumerGroupCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <consumer-group>",
		Short:             "Describe a Kafka consumer group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              pcmd.NewCLIRunE(c.describe),
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

	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return err
	}

	groupCmdResp, httpResp, err := kafkaREST.Client.ConsumerGroupV3Api.GetKafkaConsumerGroup(kafkaREST.Context, lkc, consumerGroupId).Execute()
	if err != nil {
		return kafkaRestError(pcmd.GetCloudKafkaRestBaseUrl(kafkaREST.Client), err, httpResp)
	}

	groupCmdConsumersResp, httpResp, err := kafkaREST.Client.ConsumerGroupV3Api.ListKafkaConsumers(kafkaREST.Context, lkc, consumerGroupId).Execute()
	if err != nil {
		return kafkaRestError(pcmd.GetCloudKafkaRestBaseUrl(kafkaREST.Client), err, httpResp)
	}

	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	groupData := getGroupData(groupCmdResp, groupCmdConsumersResp)

	if outputOption == output.Human.String() {
		return printGroupHumanDescribe(cmd, groupData)
	}

	return output.StructuredOutputForCommand(cmd, outputOption, groupData)
}

func getGroupData(groupCmdResp cloudkafkarest.ConsumerGroupData, groupCmdConsumersResp cloudkafkarest.ConsumerDataList) *groupData {
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
		instanceId := ""
		if consumerResp.InstanceId.IsSet() {
			instanceId = *consumerResp.InstanceId.Get()
		}
		consumerData := consumerData{
			ConsumerGroupId: groupCmdResp.ConsumerGroupId,
			ConsumerId:      consumerResp.ConsumerId,
			InstanceId:      instanceId,
			ClientId:        consumerResp.ClientId,
		}
		groupData.Consumers[i] = consumerData
	}

	return groupData
}

func getStringBroker(relationship cloudkafkarest.Relationship) string {
	// relationship.Related will look like ".../v3/clusters/{cluster_id}/brokers/{broker_id}
	splitString := strings.SplitAfter(relationship.Related, "brokers/")
	// if relationship was an empty string or did not contain "brokers/"
	if len(splitString) < 2 {
		return ""
	}
	// returning brokerId
	return splitString[1]
}

func printGroupHumanDescribe(cmd *cobra.Command, groupData *groupData) error {
	// printing non-consumer information in table format first
	if err := tables.RenderTableOut(convertGroupToDescribeStruct(groupData), groupDescribeFields, groupDescribeHumanRenames, cmd.OutOrStdout()); err != nil {
		return err
	}

	utils.Print(cmd, "\nConsumers\n\n")

	// printing consumer information in list table format
	consumerTableEntries := make([][]string, len(groupData.Consumers))
	for i, consumer := range groupData.Consumers {
		consumerTableEntries[i] = printer.ToRow(&consumer, groupDescribeConsumersFields)
	}

	printer.RenderCollectionTable(consumerTableEntries, groupDescribeConsumerTableLabels)
	return nil
}

func convertGroupToDescribeStruct(groupData *groupData) *groupDescribeStruct {
	return &groupDescribeStruct{
		ClusterId:         groupData.ClusterId,
		ConsumerGroupId:   groupData.ConsumerGroupId,
		Coordinator:       groupData.Coordinator,
		IsSimple:          groupData.IsSimple,
		PartitionAssignor: groupData.PartitionAssignor,
		State:             groupData.State,
	}
}
