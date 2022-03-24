package kafka

import (
	"net/http"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (replicaCommand *replicaCommand) newListCommand() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(replicaCommand.list),
		Short: "List Kafka replica statuses.",
		Long:  "List partition-replicas statuses filtered by topic and partition via Confluent Kafka REST.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List the replica statuses for partition 1 of "my_topic".`,
				Code: "confluent kafka replica list --topic my_topic --partition 1",
			},
			examples.Example{
				Text: `List the replicas statuses for topic "my_topic".`,
				Code: "confluent kafka replica list --topic my_topic",
			},
		),
	}
	listCmd.Flags().String("topic", "", "Topic name.")
	listCmd.Flags().Int32("partition", -1, "Partition ID.")
	listCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(listCmd)
	_ = listCmd.MarkFlagRequired("topic")
	replicaCommand.AddCommand(listCmd)
	return listCmd
}

func (replicaCommand *replicaCommand) list(cmd *cobra.Command, _ []string) error {
	topic, partitionId, err := readFlagValues(cmd)
	if err != nil {
		return err
	}
	restClient, restContext, err := initKafkaRest(replicaCommand.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	var replicaStatusDataList kafkarestv3.ReplicaStatusDataList
	var resp *http.Response
	if partitionId != -1 {
		replicaStatusDataList, resp, err = restClient.ReplicaStatusApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicaStatusGet(restContext, clusterId, topic, partitionId)
	} else {
		replicaStatusDataList, resp, err = restClient.ReplicaStatusApi.ClustersClusterIdTopicsTopicNamePartitionsReplicaStatusGet(restContext, clusterId, topic)
	}
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	outputWriter, err := output.NewListOutputWriter(cmd, replicaStatusListFields, replicaHumanFields, camelToSnake(replicaStatusListFields))
	if err != nil {
		return err
	}
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	humanOutput := format == output.Human.String()
	for _, data := range replicaStatusDataList.Data {
		if humanOutput {
			d := &struct {
				ClusterId          string
				TopicName          string
				BrokerId           int32
				PartitionId        int32
				IsLeader           bool
				IsObserver         bool
				IsIsrEligible      bool
				IsInIsr            bool
				IsCaughtUp         bool
				LogStartOffset     int64
				LogEndOffset       int64
				LastCaughtUpTimeMs string
				LastFetchTimeMs    string
				LinkName           string
			}{
				ClusterId:          data.ClusterId,
				TopicName:          data.TopicName,
				BrokerId:           data.BrokerId,
				PartitionId:        data.PartitionId,
				IsLeader:           data.IsLeader,
				IsObserver:         data.IsObserver,
				IsIsrEligible:      data.IsIsrEligible,
				IsInIsr:            data.IsInIsr,
				IsCaughtUp:         data.IsCaughtUp,
				LogStartOffset:     data.LogStartOffset,
				LogEndOffset:       data.LogEndOffset,
				LastCaughtUpTimeMs: utils.FormatUnixTime(data.LastCaughtUpTimeMs),
				LastFetchTimeMs:    utils.FormatUnixTime(data.LastFetchTimeMs),
				LinkName:           data.LinkName,
			}
			outputWriter.AddElement(d)
		} else {
			outputWriter.AddElement(&data)
		}
	}
	return outputWriter.Out()
}

func readFlagValues(cmd *cobra.Command) (string, int32, error) {
	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return "", -1, err
	}
	partition, err := cmd.Flags().GetInt32("partition")
	if err != nil {
		return "", -1, err
	}
	return topic, partition, nil
}
