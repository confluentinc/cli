package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *partitionCommand) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a Kafka partition.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describeOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe partition "1" for topic "my_topic".`,
				Code: "confluent kafka partition describe 1 --topic my_topic",
			},
		),
	}

	cmd.Flags().String("topic", "", "Topic name to list partitions of.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("topic"))

	return cmd
}

func (c *partitionCommand) describeOnPrem(cmd *cobra.Command, args []string) error {
	partitionId, err := partitionIdFromArg(args)
	if err != nil {
		return err
	}

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}

	partition, resp, err := restClient.PartitionV3Api.GetKafkaPartition(restContext, clusterId, topic, partitionId)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	table := output.NewTable(cmd)
	table.Add(&partitionOut{
		ClusterId:   partition.ClusterId,
		TopicName:   partition.TopicName,
		PartitionId: partition.PartitionId,
		LeaderId:    parseLeaderId(partition.Leader.Related),
	})
	return table.Print()
}
