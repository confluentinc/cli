package kafka

import (
	"net/http"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

type replicaHumanOut struct {
	ClusterId          string `human:"Cluster"`
	TopicName          string `human:"Topic Name"`
	BrokerId           int32  `human:"Broker ID"`
	PartitionId        int32  `human:"Partition ID"`
	IsLeader           bool   `human:"Leader"`
	IsObserver         bool   `human:"Observer"`
	IsIsrEligible      bool   `human:"ISR Eligible"`
	IsInIsr            bool   `human:"In ISR"`
	IsCaughtUp         bool   `human:"Caught Up"`
	LogStartOffset     int64  `human:"Log Start Offset"`
	LogEndOffset       int64  `human:"Log End Offset"`
	LastCaughtUpTimeMs string `human:"Last Caught Up Time (ms)"`
	LastFetchTimeMs    string `human:"Last Fetch Time (ms)"`
	LinkName           string `human:"Link Name"`
}

type replicaSerializedOut struct {
	ClusterId          string `serialized:"cluster_id"`
	TopicName          string `serialized:"topic_name"`
	BrokerId           int32  `serialized:"broker_id"`
	PartitionId        int32  `serialized:"partition_id"`
	IsLeader           bool   `serialized:"is_leader"`
	IsObserver         bool   `serialized:"is_observer"`
	IsIsrEligible      bool   `serialized:"is_isr_eligible"`
	IsInIsr            bool   `serialized:"is_in_isr"`
	IsCaughtUp         bool   `serialized:"is_caught_up"`
	LogStartOffset     int64  `serialized:"log_start_offset"`
	LogEndOffset       int64  `serialized:"log_end_offset"`
	LastCaughtUpTimeMs int64  `serialized:"last_caught_up_time_ms"`
	LastFetchTimeMs    int64  `serialized:"last_fetch_time_ms"`
	LinkName           string `serialized:"link_name"`
}

func (c *replicaCommand) newStatusListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka replica statuses.",
		Long:  "List the statuses of partition replicas, filtered by topic and partition.",
		Args:  cobra.NoArgs,
		RunE:  c.statusList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List the replica statuses for partition 1 of topic "my-topic".`,
				Code: "confluent kafka replica status list --topic my-topic --partition 1",
			},
			examples.Example{
				Text: `List the replica statuses for topic "my-topic".`,
				Code: "confluent kafka replica status list --topic my-topic",
			},
		),
	}

	cmd.Flags().String("topic", "", "Topic name.")
	cmd.Flags().Int32("partition", 0, "Partition ID.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("topic"))

	return cmd
}

func (c *replicaCommand) statusList(cmd *cobra.Command, _ []string) error {
	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	var replicas kafkarestv3.ReplicaStatusDataList
	var resp *http.Response
	if cmd.Flags().Changed("partition") {
		partitionId, err := cmd.Flags().GetInt32("partition")
		if err != nil {
			return err
		}

		replicas, resp, err = restClient.ReplicaStatusApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicaStatusGet(restContext, clusterId, topic, partitionId)
		if err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
	} else {
		replicas, resp, err = restClient.ReplicaStatusApi.ClustersClusterIdTopicsTopicNamePartitionsReplicaStatusGet(restContext, clusterId, topic)
		if err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
	}

	list := output.NewList(cmd)
	if output.GetFormat(cmd) == output.Human {
		for _, replica := range replicas.Data {
			list.Add(&replicaHumanOut{
				ClusterId:          replica.ClusterId,
				TopicName:          replica.TopicName,
				BrokerId:           replica.BrokerId,
				PartitionId:        replica.PartitionId,
				IsLeader:           replica.IsLeader,
				IsObserver:         replica.IsObserver,
				IsIsrEligible:      replica.IsIsrEligible,
				IsInIsr:            replica.IsInIsr,
				IsCaughtUp:         replica.IsCaughtUp,
				LogStartOffset:     replica.LogStartOffset,
				LogEndOffset:       replica.LogEndOffset,
				LastCaughtUpTimeMs: utils.FormatUnixTime(replica.LastCaughtUpTimeMs),
				LastFetchTimeMs:    utils.FormatUnixTime(replica.LastFetchTimeMs),
				LinkName:           replica.LinkName,
			})
		}
	} else {
		for _, replica := range replicas.Data {
			list.Add(&replicaSerializedOut{
				ClusterId:          replica.ClusterId,
				TopicName:          replica.TopicName,
				BrokerId:           replica.BrokerId,
				PartitionId:        replica.PartitionId,
				IsLeader:           replica.IsLeader,
				IsObserver:         replica.IsObserver,
				IsIsrEligible:      replica.IsIsrEligible,
				IsInIsr:            replica.IsInIsr,
				IsCaughtUp:         replica.IsCaughtUp,
				LogStartOffset:     replica.LogStartOffset,
				LogEndOffset:       replica.LogEndOffset,
				LastCaughtUpTimeMs: replica.LastCaughtUpTimeMs,
				LastFetchTimeMs:    replica.LastFetchTimeMs,
				LinkName:           replica.LinkName,
			})
		}
	}
	return list.Print()
}
