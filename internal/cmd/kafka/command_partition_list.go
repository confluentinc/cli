package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *partitionCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka partitions.",
		Long:  "List the partitions that belong to a specified topic via Confluent Kafka REST.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List the partitions of topic "my_topic".`,
				Code: "confluent kafka partition list --topic my_topic",
			},
		),
	}

	cmd.Flags().String("topic", "", "Topic name to list partitions of.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("topic")

	return cmd
}

func (c *partitionCommand) list(cmd *cobra.Command, _ []string) error {
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
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
	partitionListResp, resp, err := restClient.PartitionV3Api.ListKafkaPartitions(restContext, clusterId, topic)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}
	partitionDatas := partitionListResp.Data

	outputWriter, err := output.NewListOutputWriter(cmd, []string{"ClusterId", "TopicName", "PartitionId", "LeaderId"}, []string{"Cluster ID", "Topic Name", "Partition ID", "Leader ID"}, []string{"cluster_id", "topic_name", "partition_id", "leader_id"})
	if err != nil {
		return err
	}
	for _, partition := range partitionDatas {
		s := &struct {
			ClusterId   string
			TopicName   string
			PartitionId int32
			LeaderId    int32
		}{
			ClusterId:   partition.ClusterId,
			TopicName:   partition.TopicName,
			PartitionId: partition.PartitionId,
			LeaderId:    parseLeaderId(partition.Leader),
		}
		outputWriter.AddElement(s)
	}

	return outputWriter.Out()
}
