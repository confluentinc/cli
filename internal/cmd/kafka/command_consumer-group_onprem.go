package kafka

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

type groupCommandOnPrem struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner           pcmd.PreRunner
}

type lagCommandOnPrem struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner 			pcmd.PreRunner
}

func NewGroupCommandOnPrem(prerunner pcmd.PreRunner) *cobra.Command {
	groupCmd := &groupCommandOnPrem{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(
			&cobra.Command{
				Use: "consumer-group",
				Short: "Manage Kafka consumer groups",
			}, prerunner, OnPremGroupSubcommandFlags),
		prerunner: prerunner,
	}
	groupCmd.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(groupCmd.AuthenticatedCLICommand))
	groupCmd.init()
	return groupCmd.Command
}

func (g *groupCommandOnPrem) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka consumer groups.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(g.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all consumer groups of a specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "ccloud kafka consumer-group list --url http://localhost:8082",
			},
		),
	}
	listCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	g.AddCommand(listCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <consumer-group>",
		Short: "Describe a Kafka consumer group.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(g.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the ``my_consumer_group`` consumer group of a specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "ccloud kafka consumer-group describe my_consumer_group --url http://localhost:8082",
			},
		),
	}
	describeCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	g.AddCommand(describeCmd)

	lagCmd := NewLagCommandOnPrem(g.prerunner)
	g.AddCommand(lagCmd.Command)
}

func (g *groupCommandOnPrem) list(cmd *cobra.Command, args []string) error {
	kafkaRest, err := g.GetKafkaREST()
	if err != nil {
		return err
	}
	restClient, restContext, err := getKafkaRestClientAndContext(cmd, kafkaRest)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	groupCmdResp, resp, err := restClient.ConsumerGroupApi.ClustersClusterIdConsumerGroupsGet(restContext, clusterId)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
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

func (g *groupCommandOnPrem) describe(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]
	kafkaRest, err := g.GetKafkaREST()
	if err != nil {
		return err
	}
	restClient, restContext, err := getKafkaRestClientAndContext(cmd, kafkaRest)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	groupCmdResp, resp, err := restClient.ConsumerGroupApi.ClustersClusterIdConsumerGroupsConsumerGroupIdGet(
		restContext,
		clusterId,
		consumerGroupId)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	groupCmdConsumersResp, resp, err := restClient.ConsumerGroupApi.ClustersClusterIdConsumerGroupsConsumerGroupIdConsumersGet(
		restContext,
		clusterId,
		consumerGroupId)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
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

func NewLagCommandOnPrem(prerunner pcmd.PreRunner) *lagCommandOnPrem {
	lagCmd := &lagCommandOnPrem{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(
			&cobra.Command{
				Use:   "lag",
				Short: "View consumer lag.",
			}, prerunner, OnPremLagSubcommandFlags),
		prerunner: prerunner,
	}
	lagCmd.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(lagCmd.AuthenticatedCLICommand))
	lagCmd.init()
	return lagCmd
}

func (lagCmd *lagCommandOnPrem) init() {
	summarizeLagCmd := &cobra.Command{
		Use:   "summarize <id>",
		Short: "Summarize consumer lag for a Kafka consumer group.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(lagCmd.summarizeLag),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Summarize the lag for the ``my_consumer_group`` consumer-group of a specified cluster (providing Kafka REST Proxy endpoint).",
				// ahu: should the examples include the other required flag(s)? --cluster
				Code: "ccloud kafka consumer-group lag summarize my_consumer_group --url http://localhost:8082",
			},
		),
	}
	summarizeLagCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	summarizeLagCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	summarizeLagCmd.Flags().SortFlags = false
	lagCmd.AddCommand(summarizeLagCmd)

	listLagCmd := &cobra.Command{
		Use:   "list <id>",
		Short: "List consumer lags for a Kafka consumer group.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(lagCmd.listLag),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all consumer lags for consumers in the ``my_consumer_group`` consumer-group of a specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "ccloud kafka consumer-group lag list my_consumer_group --url http://localhost:8082",
			},
		),
	}
	listLagCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	listLagCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listLagCmd.Flags().SortFlags = false
	lagCmd.AddCommand(listLagCmd)

	getLagCmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get consumer lag for a partition consumed by a Kafka consumer group.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(lagCmd.getLag),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Get the consumer lag for topic ``my_topic`` partition ``0`` consumed by consumer-group ``my_consumer_group`` of a specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "ccloud kafka consumer-group lag get my_consumer_group --topic my_topic --partition 0 --url http://localhost:8082",
			},
		),
	}
	// ahu: handle defaults
	getLagCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	getLagCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	getLagCmd.Flags().String("topic", "", "Topic name.")
	getLagCmd.Flags().Int32("partition", -1, "Partition ID.")
	check(getLagCmd.MarkFlagRequired("topic"))
	check(getLagCmd.MarkFlagRequired("partition"))
	getLagCmd.Flags().SortFlags = false
	lagCmd.AddCommand(getLagCmd)
}

func (lagCmd *lagCommandOnPrem) summarizeLag(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]
	kafkaRest, err := lagCmd.GetKafkaREST()
	if err != nil {
		return err
	}
	restClient, restContext, err := getKafkaRestClientAndContext(cmd, kafkaRest)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	lagSummaryResp, _, err :=
		restClient.ConsumerGroupApi.ClustersClusterIdConsumerGroupsConsumerGroupIdLagSummaryGet(
			restContext,
			clusterId,
			consumerGroupId)
	if err != nil {
		return err
	}
	return output.DescribeObject(
		cmd,
		convertLagSummaryToStruct(lagSummaryResp),
		lagSummaryFields,
		lagSummaryHumanRenames,
		lagSummaryStructuredRenames)
}

func (lagCmd *lagCommandOnPrem) listLag(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]
	kafkaRest, err := lagCmd.GetKafkaREST()
	if err != nil {
		return err
	}
	restClient, restContext, err := getKafkaRestClientAndContext(cmd, kafkaRest)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	lagSummaryResp, _, err :=
		restClient.ConsumerGroupApi.ClustersClusterIdConsumerGroupsConsumerGroupIdLagsGet(
			restContext,
			clusterId,
			consumerGroupId)
	if err != nil {
		return err
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

func (lagCmd *lagCommandOnPrem) getLag(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]
	topicName, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}
	partitionId, err := cmd.Flags().GetInt32("partition")
	if err != nil {
		return err
	}
	kafkaRest, err := lagCmd.GetKafkaREST()
	if err != nil {
		return err
	}
	restClient, restContext, err := getKafkaRestClientAndContext(cmd, kafkaRest)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	lagGetResp, _, err :=
		restClient.PartitionApi.ClustersClusterIdConsumerGroupsConsumerGroupIdLagsTopicNamePartitionsPartitionIdGet(
			restContext,
			clusterId,
			consumerGroupId,
			topicName,
			partitionId)
	if err != nil {
		return err
	}
	return output.DescribeObject(
		cmd,
		convertLagToStruct(lagGetResp),
		lagFields,
		lagGetHumanRenames,
		lagGetStructuredRenames)
}
