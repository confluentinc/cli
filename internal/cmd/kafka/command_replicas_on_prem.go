package kafka

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
	"net/http"
)

type replicaCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func NewReplicaCommand(prerunner pcmd.PreRunner) *cobra.Command {
	replicaCommand := &replicaCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(
			&cobra.Command{
				Use:         "replica",
				Short:       "Manage Kafka replicas.",
				Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
			}, prerunner, OnPremTopicSubcommandFlags),
	}
	replicaCommand.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(replicaCommand.AuthenticatedCLICommand))
	replicaCommand.init()
	return replicaCommand.Command
}

func (replicaCommand *replicaCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(replicaCommand.list),
		Short: "List Kafka replicas.",
		Long:  "List partition-replicas filtered by topic, partition, and broker via Confluent Kafka REST.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List the partitions for `my_topic`.",
				Code: "confluent kafka partition list --topic my_topic",
			},
		),
	}
	listCmd.Flags().String("topic", "", "Topic name.")
	listCmd.Flags().Int32("partition", -1, "Partition ID.")
	listCmd.Flags().Int32("broker", -1, "Broker ID.")
	listCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	replicaCommand.AddCommand(listCmd)
}

func (replicaCommand *replicaCommand) list(cmd *cobra.Command, _ []string) error {
	topic, partitionId, brokerId, err := validateFlagCombo(cmd)
	if err != nil {
		return err
	}
	restClient, restContext, err := initKafkaRest(replicaCommand.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return nil
	}
	var replicaDataList kafkarestv3.ReplicaDataList
	var resp *http.Response
	if partitionId != -1 && topic != "" {
		if brokerId != -1 {
			var replicaData kafkarestv3.ReplicaData
			replicaData, resp, err = restClient.ReplicaApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicasBrokerIdGet(restContext, clusterId, topic, partitionId, brokerId)
			replicaDataList.Data = append(replicaDataList.Data, replicaData)
		} else {
			replicaDataList, resp, err = restClient.ReplicaApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicasGet(restContext, clusterId, topic, partitionId)
		}
	} else {
		replicaDataList, resp, err = restClient.ReplicaApi.ClustersClusterIdBrokersBrokerIdPartitionReplicasGet(restContext, clusterId, brokerId)
	}
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	outputWriter, err := output.NewListOutputWriter(cmd, []string{"ClusterId", "BrokerId", "TopicName", "PartitionId", "IsLeader", "IsInSync"}, []string{"Cluster ID", "Broker ID", "Topic Name", "Partition ID", "Is Leader", "Is In Sync"}, []string{"cluster_id", "broker_id", "topic_name", "partition_id", "is_leader", "is_in_sync"})
	if err != nil {
		return err
	}
	for _, data := range replicaDataList.Data {
		outputWriter.AddElement(&data)
	}
	return outputWriter.Out()
}

func validateFlagCombo(cmd *cobra.Command) (string, int32, int32, error) {
	// valid flag combinations are topic+partition, topic+partition+broker, or just broker
	topicSet := cmd.Flags().Changed("topic")
	partitionSet := cmd.Flags().Changed("partition")
	brokerSet := cmd.Flags().Changed("broker")

	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return "", -1, -1, err
	}
	brokerId, err := cmd.Flags().GetInt32("broker")
	if err != nil {
		return "", -1, -1, err
	}
	partitionId, err := cmd.Flags().GetInt32("partition")
	if err != nil {
		return "", -1, -1, err
	}
	if !topicSet && !brokerSet && !partitionSet {
		return "", -1, -1, errors.NewErrorWithSuggestions(errors.MustEnterValidFlagComboErrorMsg, errors.ValidReplicaFlagsSuggestions)
	} else if (topicSet && !partitionSet) || (!topicSet && partitionSet) {
		return "", -1, -1, errors.NewErrorWithSuggestions(errors.MustSpecifyTopicAndPartitionErrorMsg, errors.ValidReplicaFlagsSuggestions)
	}
	return topic, partitionId, brokerId, nil
 }