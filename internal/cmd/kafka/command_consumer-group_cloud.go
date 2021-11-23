package kafka

import (
	"fmt"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/shell/completer"
	"github.com/confluentinc/cli/internal/pkg/utils"

	"github.com/c-bata/go-prompt"
	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/go-printer/tables"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	groupListFields           = []string{"ClusterId", "ConsumerGroupId", "IsSimple", "State"}
	groupListHumanLabels      = []string{"Cluster", "ConsumerGroup", "Simple", "State"}
	groupListStructuredLabels = []string{"cluster", "consumer_group", "simple", "state"}
	// groupDescribe vars and struct used for human output
	groupDescribeFields       = []string{"ClusterId", "ConsumerGroupId", "Coordinator", "IsSimple", "PartitionAssignor", "State"}
	groupDescribeHumanRenames = map[string]string{
		"ClusterId":       "Cluster",
		"ConsumerGroupId": "ConsumerGroup",
		"IsSimple":        "Simple"}
	groupDescribeConsumersFields     = []string{"ConsumerGroupId", "ConsumerId", "InstanceId", "ClientId"}
	groupDescribeConsumerTableLabels = []string{"Consumer Group", "Consumer", "Instance", "Client"}
	lagSummaryFields                 = []string{"ClusterId", "ConsumerGroupId", "TotalLag", "MaxLag", "MaxLagConsumerId", "MaxLagInstanceId", "MaxLagClientId", "MaxLagTopicName", "MaxLagPartitionId"}
	lagSummaryHumanRenames           = map[string]string{
		"ClusterId":         "Cluster",
		"ConsumerGroupId":   "ConsumerGroup",
		"MaxLagConsumerId":  "MaxLagConsumer",
		"MaxLagInstanceId":  "MaxLagInstance",
		"MaxLagClientId":    "MaxLagClient",
		"MaxLagTopicName":   "MaxLagTopic",
		"MaxLagPartitionId": "MaxLagPartition"}
	lagSummaryStructuredRenames = map[string]string{
		"ClusterId":         "cluster",
		"ConsumerGroupId":   "consumer_group",
		"TotalLag":          "total_lag",
		"MaxLag":            "max_lag",
		"MaxLagConsumerId":  "max_lag_consumer",
		"MaxLagInstanceId":  "max_lag_instance",
		"MaxLagClientId":    "max_lag_client",
		"MaxLagTopicName":   "max_lag_topic",
		"MaxLagPartitionId": "max_lag_partition"}
	lagFields               = []string{"ClusterId", "ConsumerGroupId", "Lag", "LogEndOffset", "CurrentOffset", "ConsumerId", "InstanceId", "ClientId", "TopicName", "PartitionId"}
	lagListHumanLabels      = []string{"Cluster", "ConsumerGroup", "Lag", "LogEndOffset", "CurrentOffset", "Consumer", "Instance", "Client", "Topic", "Partition"}
	lagListStructuredLabels = []string{"cluster", "consumer_group", "lag", "log_end_offset", "current_offset", "consumer", "instance", "client", "topic", "partition"}
	lagGetHumanRenames      = map[string]string{
		"ClusterId":       "Cluster",
		"ConsumerGroupId": "ConsumerGroup",
		"ConsumerId":      "Consumer",
		"InstanceId":      "Instance",
		"ClientId":        "Client",
		"TopicName":       "Topic",
		"PartitionId":     "Partition"}
	lagGetStructuredRenames = map[string]string{
		"ClusterId":       "cluster",
		"ConsumerGroupId": "consumer_group",
		"Lag":             "lag",
		"LogEndOffset":    "log_end_offset",
		"CurrentOffset":   "current_offset",
		"ConsumerId":      "consumer",
		"InstanceId":      "instance",
		"ClientId":        "client",
		"TopicName":       "topic",
		"PartitionId":     "partition"}
)

type groupCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner               pcmd.PreRunner
	serverCompleter         completer.ServerSideCompleter
	completableChildren     []*cobra.Command
	completableFlagChildren map[string][]*cobra.Command
	lagCmd                  *lagCommand
}

type consumerData struct {
	ConsumerGroupId string `json:"consumer_group" yaml:"consumer_group"`
	ConsumerId      string `json:"consumer" yaml:"consumer"`
	InstanceId      string `json:"instance" yaml:"instance"`
	ClientId        string `json:"client" yaml:"client"`
}

type groupData struct {
	ClusterId         string         `json:"cluster" yaml:"cluster"`
	ConsumerGroupId   string         `json:"consumer_group" yaml:"consumer_group"`
	Coordinator       string         `json:"coordinator" yaml:"coordinator"`
	IsSimple          bool           `json:"simple" yaml:"simple"`
	PartitionAssignor string         `json:"partition_assignor" yaml:"partition_assignor"`
	State             string         `json:"state" yaml:"state"`
	Consumers         []consumerData `json:"consumers" yaml:"consumers"`
}

type groupDescribeStruct struct {
	ClusterId         string
	ConsumerGroupId   string
	Coordinator       string
	IsSimple          bool
	PartitionAssignor string
	State             string
}

type lagCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner           pcmd.PreRunner
	completableChildren []*cobra.Command
	*groupCommand
}

type lagSummaryStruct struct {
	ClusterId         string
	ConsumerGroupId   string
	TotalLag          int64
	MaxLag            int64
	MaxLagConsumerId  string
	MaxLagInstanceId  string
	MaxLagClientId    string
	MaxLagTopicName   string
	MaxLagPartitionId int32
}

type lagDataStruct struct {
	ClusterId       string
	ConsumerGroupId string
	Lag             int64
	LogEndOffset    int64
	CurrentOffset   int64
	ConsumerId      string
	InstanceId      string
	ClientId        string
	TopicName       string
	PartitionId     int32
}

func NewGroupCommand(prerunner pcmd.PreRunner, serverCompleter completer.ServerSideCompleter) *groupCommand {
	command := &cobra.Command{
		Use:    "consumer-group",
		Short:  "Manage Kafka consumer groups.",
		Hidden: true,
	}
	groupCmd := &groupCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(command, prerunner, GroupSubcommandFlags),
		prerunner:                     prerunner,
		serverCompleter:               serverCompleter,
	}
	groupCmd.init()
	return groupCmd
}

func (g *groupCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka consumer groups.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(g.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all consumer groups.",
				Code: "confluent kafka consumer-group list",
			},
		),
		Hidden: true,
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	g.AddCommand(listCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <consumer-group>",
		Short: "Describe a Kafka consumer group.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(g.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the `my-consumer-group` consumer group.",
				Code: "confluent kafka consumer-group describe my-consumer-group",
			},
		),
		Hidden: true,
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	g.AddCommand(describeCmd)

	lagCmd := NewLagCommand(g.prerunner, g)
	g.AddCommand(lagCmd.Command)
	g.lagCmd = lagCmd

	g.completableChildren = append(lagCmd.completableChildren, listCmd, describeCmd)
	g.completableFlagChildren = map[string][]*cobra.Command{
		"cluster": append(lagCmd.completableChildren, listCmd, describeCmd),
	}
}

func (g *groupCommand) list(cmd *cobra.Command, _ []string) error {
	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(g.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}
	groupCmdResp, httpResp, err :=
		kafkaREST.Client.ConsumerGroupApi.ClustersClusterIdConsumerGroupsGet(
			kafkaREST.Context,
			lkc)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}
	outputWriter, err := output.NewListOutputWriter(cmd, groupListFields, groupListHumanLabels, groupListStructuredLabels)
	if err != nil {
		return err
	}
	for _, groupData := range groupCmdResp.Data {
		outputWriter.AddElement(&groupData)
	}
	return outputWriter.Out()
}

func (g *groupCommand) describe(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]

	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(g.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}
	groupCmdResp, httpResp, err :=
		kafkaREST.Client.ConsumerGroupApi.ClustersClusterIdConsumerGroupsConsumerGroupIdGet(
			kafkaREST.Context,
			lkc,
			consumerGroupId)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}
	groupCmdConsumersResp, httpResp, err :=
		kafkaREST.Client.ConsumerGroupApi.ClustersClusterIdConsumerGroupsConsumerGroupIdConsumersGet(
			kafkaREST.Context,
			lkc,
			consumerGroupId)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}
	groupData := getGroupData(groupCmdResp, groupCmdConsumersResp)
	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	if outputOption == output.Human.String() {
		return printGroupHumanDescribe(cmd, groupData)
	}
	return output.StructuredOutputForCommand(cmd, outputOption, groupData)
}

func getGroupData(groupCmdResp kafkarestv3.ConsumerGroupData, groupCmdConsumersResp kafkarestv3.ConsumerDataList) *groupData {
	groupData := &groupData{
		ClusterId:         groupCmdResp.ClusterId,
		ConsumerGroupId:   groupCmdResp.ConsumerGroupId,
		Coordinator:       getStringBroker(groupCmdResp.Coordinator),
		IsSimple:          groupCmdResp.IsSimple,
		PartitionAssignor: groupCmdResp.PartitionAssignor,
		State:             getStringState(groupCmdResp.State),
		Consumers:         make([]consumerData, len(groupCmdConsumersResp.Data)),
	}
	// populate Consumers list
	for i, consumerResp := range groupCmdConsumersResp.Data {
		instanceId := ""
		if consumerResp.InstanceId != nil {
			instanceId = *consumerResp.InstanceId
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

func getStringState(state kafkarestv3.ConsumerGroupState) string {
	return fmt.Sprintf("%+v", state)
}

func printGroupHumanDescribe(cmd *cobra.Command, groupData *groupData) error {
	// printing non-consumer information in table format first
	err := tables.RenderTableOut(convertGroupToDescribeStruct(groupData), groupDescribeFields, groupDescribeHumanRenames, cmd.OutOrStdout())
	if err != nil {
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

// todo: remove *groupCommand from params after fixing ServerComplete
func NewLagCommand(prerunner pcmd.PreRunner, groupCmd *groupCommand) *lagCommand {
	cliCmd := pcmd.NewAuthenticatedStateFlagCommand(
		&cobra.Command{
			Use:    "lag",
			Short:  "View consumer lag.",
			Hidden: true,
		}, prerunner, LagSubcommandFlags)
	lagCmd := &lagCommand{
		AuthenticatedStateFlagCommand: cliCmd,
		prerunner:                     prerunner,
		groupCommand:                  groupCmd,
	}
	lagCmd.init()
	return lagCmd
}

func (lagCmd *lagCommand) init() {
	summarizeLagCmd := &cobra.Command{
		Use:   "summarize <consumer-group>",
		Short: "Summarize consumer lag for a Kafka consumer group.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(lagCmd.summarizeLag),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Summarize the lag for the `my-consumer-group` consumer-group.",
				Code: "confluent kafka consumer-group lag summarize my-consumer-group",
			},
		),
		Hidden: true,
	}
	summarizeLagCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	lagCmd.AddCommand(summarizeLagCmd)

	listLagCmd := &cobra.Command{
		Use:   "list <consumer-group>",
		Short: "List consumer lags for a Kafka consumer group.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(lagCmd.listLag),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all consumer lags for consumers in the `my-consumer-group` consumer-group.",
				Code: "confluent kafka consumer-group lag list my-consumer-group",
			},
		),
		Hidden: true,
	}
	listLagCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	lagCmd.AddCommand(listLagCmd)

	getLagCmd := &cobra.Command{
		Use:   "get <consumer-group>",
		Short: "Get consumer lag for a Kafka topic partition.",
		Long:  "Get consumer lag for a Kafka topic partition consumed by a consumer group.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(lagCmd.getLag),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Get the consumer lag for topic `my-topic` partition `0` consumed by consumer-group `my-consumer-group`.",
				Code: "confluent kafka consumer-group lag get my-consumer-group --topic my-topic --partition 0",
			},
		),
		Hidden: true,
	}
	getLagCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	getLagCmd.Flags().String("topic", "", "Topic name.")
	getLagCmd.Flags().Int32("partition", 0, "Partition ID.")
	check(getLagCmd.MarkFlagRequired("topic"))
	check(getLagCmd.MarkFlagRequired("partition"))
	lagCmd.AddCommand(getLagCmd)

	lagCmd.completableChildren = []*cobra.Command{summarizeLagCmd, listLagCmd, getLagCmd}
}

func (lagCmd *lagCommand) summarizeLag(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]

	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(lagCmd.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}
	lagSummaryResp, httpResp, err :=
		kafkaREST.Client.ConsumerGroupApi.ClustersClusterIdConsumerGroupsConsumerGroupIdLagSummaryGet(
			kafkaREST.Context,
			lkc,
			consumerGroupId)

	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	return output.DescribeObject(
		cmd,
		convertLagSummaryToStruct(lagSummaryResp),
		lagSummaryFields,
		lagSummaryHumanRenames,
		lagSummaryStructuredRenames)
}

func convertLagSummaryToStruct(lagSummaryData kafkarestv3.ConsumerGroupLagSummaryData) *lagSummaryStruct {
	maxLagInstanceId := ""
	if lagSummaryData.MaxLagInstanceId != nil {
		maxLagInstanceId = *lagSummaryData.MaxLagInstanceId
	}
	return &lagSummaryStruct{
		ClusterId:         lagSummaryData.ClusterId,
		ConsumerGroupId:   lagSummaryData.ConsumerGroupId,
		TotalLag:          lagSummaryData.TotalLag,
		MaxLag:            lagSummaryData.MaxLag,
		MaxLagConsumerId:  lagSummaryData.MaxLagConsumerId,
		MaxLagInstanceId:  maxLagInstanceId,
		MaxLagClientId:    lagSummaryData.MaxLagClientId,
		MaxLagTopicName:   lagSummaryData.MaxLagTopicName,
		MaxLagPartitionId: lagSummaryData.MaxLagPartitionId,
	}
}

func (lagCmd *lagCommand) listLag(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]

	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(lagCmd.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}
	lagSummaryResp, httpResp, err :=
		kafkaREST.Client.ConsumerGroupApi.ClustersClusterIdConsumerGroupsConsumerGroupIdLagsGet(
			kafkaREST.Context,
			lkc,
			consumerGroupId)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}
	outputWriter, err := output.NewListOutputWriter(cmd, lagFields, lagListHumanLabels, lagListStructuredLabels)
	if err != nil {
		return err
	}
	for _, lagData := range lagSummaryResp.Data {
		outputWriter.AddElement(convertLagToStruct(lagData))
	}
	return outputWriter.Out()
}

func (lagCmd *lagCommand) getLag(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]
	topicName, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}
	partitionId, err := cmd.Flags().GetInt32("partition")
	if err != nil {
		return err
	}

	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(lagCmd.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}
	lagGetResp, httpResp, err :=
		kafkaREST.Client.PartitionApi.ClustersClusterIdConsumerGroupsConsumerGroupIdLagsTopicNamePartitionsPartitionIdGet(
			kafkaREST.Context,
			lkc,
			consumerGroupId,
			topicName,
			partitionId)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}
	return output.DescribeObject(
		cmd,
		convertLagToStruct(lagGetResp),
		lagFields,
		lagGetHumanRenames,
		lagGetStructuredRenames)
}

func convertLagToStruct(lagData kafkarestv3.ConsumerLagData) *lagDataStruct {
	instanceId := ""
	if lagData.InstanceId != nil {
		instanceId = *lagData.InstanceId
	}

	return &lagDataStruct{
		ClusterId:       lagData.ClusterId,
		ConsumerGroupId: lagData.ConsumerGroupId,
		Lag:             lagData.Lag,
		LogEndOffset:    lagData.LogEndOffset,
		CurrentOffset:   lagData.CurrentOffset,
		ConsumerId:      lagData.ConsumerId,
		InstanceId:      instanceId,
		ClientId:        lagData.ClientId,
		TopicName:       lagData.TopicName,
		PartitionId:     lagData.PartitionId,
	}
}

func (g *groupCommand) Cmd() *cobra.Command {
	return g.Command
}

func (g *groupCommand) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	consumerGroupDataList, err := listConsumerGroups(g.AuthenticatedStateFlagCommand, g.Command)
	if err != nil {
		return suggestions
	}
	for _, groupData := range consumerGroupDataList.Data {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        groupData.ConsumerGroupId,
			Description: groupData.ConsumerGroupId,
		})
	}
	return suggestions
}

func (g *groupCommand) ServerCompletableFlagChildren() map[string][]*cobra.Command {
	return g.completableFlagChildren
}

func (g *groupCommand) ServerFlagComplete() map[string]func() []prompt.Suggest {
	return map[string]func() []prompt.Suggest{
		"cluster": completer.ClusterFlagServerCompleterFunc(g.Client, g.EnvironmentId()),
		// todo: add Topic and Partition flag completion
	}
}

func (g *groupCommand) ServerCompletableChildren() []*cobra.Command {
	return g.completableChildren
}

func (lagCmd *lagCommand) Cmd() *cobra.Command {
	return lagCmd.Command
}

// HACK: using groupCommand's ServerComplete until we can figure out why calling listConsumerGroups on lagCmd is
// producing segfaults. I believe we just need to figure out why Authenticated (prerunner.go) is being called on
// groupCommand.AuthenticatedCLICommand instead of lagCmd.AuthenticatedCLICommand
func (lagCmd *lagCommand) ServerComplete() []prompt.Suggest {
	return lagCmd.groupCommand.ServerComplete()
}

func (lagCmd *lagCommand) ServerCompletableChildren() []*cobra.Command {
	return lagCmd.completableChildren
}

func listConsumerGroups(flagCmd *pcmd.AuthenticatedStateFlagCommand, cobraCmd *cobra.Command) (*kafkarestv3.ConsumerGroupDataList, error) {
	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(flagCmd, cobraCmd)
	if err != nil {
		return nil, err
	}
	groupCmdResp, httpResp, err :=
		kafkaREST.Client.ConsumerGroupApi.ClustersClusterIdConsumerGroupsGet(
			kafkaREST.Context,
			lkc)
	if err != nil {
		return nil, kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}
	return &groupCmdResp, nil
}
