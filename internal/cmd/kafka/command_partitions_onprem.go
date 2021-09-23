package kafka

import (
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
	"net/http"
	"strconv"
	"strings"
)

type partitionCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func NewPartitionCommandOnPrem(prerunner pcmd.PreRunner) *cobra.Command {
	partitionCommand := &partitionCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(
			&cobra.Command{
				Use:   "partition",
				Short: "Manage Kafka partitions.",
			}, prerunner, OnPremTopicSubcommandFlags),
	}
	partitionCommand.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(partitionCommand.AuthenticatedCLICommand))
	partitionCommand.init()
	return partitionCommand.Command
}

func (partitionCmd *partitionCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(partitionCmd.list),
		Short: "List Kafka partitions.",
		Long:  "List the partitions that belong to a specified topic via Confluent Kafka REST.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List the partitions for `my_topic`.",
				Code: "confluent kafka partition list --topic my_topic",
			},
		),
	}
	listCmd.Flags().String("topic", "", "Topic name to list partitions of.")
	listCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	check(listCmd.MarkFlagRequired("topic"))
	partitionCmd.AddCommand(listCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a Kafka partition",
		Long:  "Describe a Kafka partition via Confluent Kafka REST.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(partitionCmd.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe partition `1` for `my_topic`.",
				Code: "confluent kafka partition describe 1 --topic my_topic",
			}),
	}
	describeCmd.Flags().String("topic", "", "Topic name to list partitions of.")
	describeCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	check(describeCmd.MarkFlagRequired("topic"))
	partitionCmd.AddCommand(describeCmd)

	reassignmentsCmd := &cobra.Command{
		Use:   "get-reassignments [id]",
		Short: "Get ongoing replica reassignments.",
		Long:  "Get ongoing replica reassignments for a given cluster, topic, or partition via Confluent Kafka REST.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  pcmd.NewCLIRunE(partitionCmd.getReassignments),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Get all replica reassignments for the Kafka cluster",
				Code: "confluent kafka partition get-reassignments",
			},
			examples.Example{
				Text: "Get replica reassignments for `my_topic`",
				Code: "confluent kafka partition get-reassignments --topic my_topic",
			},
			examples.Example{
				Text: "Get replica reassignments for partition `1` of `my_topic`.",
				Code: "confluent kafka partition get-reassignments 1 --topic my_topic",
			}),
	}
	reassignmentsCmd.Flags().String("topic", "", "Topic name to search by.")
	reassignmentsCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	reassignmentsCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	reassignmentsCmd.Flags().SortFlags = false
	partitionCmd.AddCommand(reassignmentsCmd)
}

func (partitionCmd *partitionCommand) list(cmd *cobra.Command, _ []string) error {
	restClient, restContext, err := initKafkaRest(partitionCmd.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}
	partitionListResp, resp, err := restClient.PartitionApi.ClustersClusterIdTopicsTopicNamePartitionsGet(restContext, clusterId, topic)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	partitionDatas := partitionListResp.Data

	outputWriter, err := output.NewListOutputWriter(cmd, []string{"ClusterId", "TopicName", "PartitionId", "LeaderId"}, []string{"Cluster ID", "Topic Name", "Partition ID", "Leader ID"}, []string{"cluster_id", "topic_name", "partition_id", "leader_id"})
	if err != nil {
		return err
	}
	for _, partition := range partitionDatas {
		s := &struct {
			ClusterId string
			TopicName  string
			PartitionId int32
			LeaderId    int32
		}{
			ClusterId: partition.ClusterId,
			TopicName: partition.TopicName,
			PartitionId: partition.PartitionId,
			LeaderId: parseLeaderId(partition.Leader),
		}
		outputWriter.AddElement(s)
	}

	return outputWriter.Out()
}

func parseLeaderId(leader kafkarestv3.Relationship) int32 {
	index := strings.LastIndex(leader.Related,"/")
	idStr := leader.Related[index+1:]
	leaderId, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return -1 //shouldn't happen
	}
	return int32(leaderId)
}

func (partitionCmd *partitionCommand) describe(cmd *cobra.Command, args []string) error {
	partitionIdStr := args[0]
	i, err := strconv.ParseInt(partitionIdStr, 10, 32)
	if err != nil {
		return err // TODO custom error
	}
	partitionId := int32(i)
	restClient, restContext, err := initKafkaRest(partitionCmd.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}
	partitionGetResp, resp, err := restClient.PartitionApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdGet(restContext, clusterId, topic, partitionId)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	s := &struct {
		ClusterId string
		TopicName  string
		PartitionId int32
		LeaderId    int32
	}{
		ClusterId: partitionGetResp.ClusterId,
		TopicName: partitionGetResp.TopicName,
		PartitionId: partitionGetResp.PartitionId,
		LeaderId: parseLeaderId(partitionGetResp.Leader),
	}
	return output.DescribeObject(cmd, s, []string{"ClusterId", "TopicName", "PartitionId", "LeaderId"}, map[string]string{"ClusterId":"Cluster ID", "TopicName":"Topic Name", "PartitionId":"Partition ID", "LeaderId":"Leader ID"}, map[string]string{"ClusterId":"cluster_id", "TopicName":"topic_name", "PartitionId":"partition_id", "LeaderId":"leader_id"})
}

func (partitionCmd *partitionCommand) getReassignments(cmd *cobra.Command, args []string) error {
	restClient, restContext, err := initKafkaRest(partitionCmd.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}
	var reassignmentListResp kafkarestv3.ReassignmentDataList
	var resp *http.Response
	if len(args) > 0 {
		partitionIdStr := args[0] // todo maybe refactor out
		i, err := strconv.ParseInt(partitionIdStr, 10, 32)
		if err != nil {
			return err // TODO custom error
		}
		partitionId := int32(i)
		if topic == "" {
			return errors.New("must specify partition id and topic together")
		}
		var reassignmentGetResp kafkarestv3.ReassignmentData
		reassignmentGetResp, resp, err = restClient.PartitionApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReassignmentGet(restContext, clusterId, topic, partitionId)
		fmt.Println(reassignmentGetResp)
		if reassignmentGetResp.Kind != "" {
			reassignmentListResp.Data = []kafkarestv3.ReassignmentData{reassignmentGetResp}
		}
	} else if topic != "" {
		reassignmentListResp, resp, err = restClient.PartitionApi.ClustersClusterIdTopicsTopicNamePartitionsReassignmentGet(restContext, clusterId, topic)
	} else {
		reassignmentListResp, resp, err = restClient.PartitionApi.ClustersClusterIdTopicsPartitionsReassignmentGet(restContext, clusterId)
	}
	fmt.Println(reassignmentListResp)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, []string{"ClusterId", "TopicName", "PartitionId", "AddingReplicas", "RemovingReplicas"}, []string{"Cluster ID", "Topic Name", "Partition ID", "Adding Replicas", "Removing Replicas"}, []string{"cluster_id", "topic_name", "partition_id", "adding_replicas", "removing_replicas"})
	if err != nil {
		return err
	}
	for _, data := range reassignmentListResp.Data {
		s := &struct {
			ClusterId string
			TopicName  string
			PartitionId int32
			AddingReplicas []int32
			RemovingReplicas []int32
		}{
			ClusterId: data.ClusterId,
			TopicName: data.TopicName,
			PartitionId: data.PartitionId,
			AddingReplicas: data.AddingReplicas,
			RemovingReplicas: data.RemovingReplicas,
		}
		outputWriter.AddElement(s)
	}

	return outputWriter.Out()
}

