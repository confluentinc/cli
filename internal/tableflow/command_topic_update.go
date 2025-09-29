package tableflow

import (
	"fmt"

	"github.com/spf13/cobra"

	tableflowv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/tableflow/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
)

func (c *command) newTopicUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <name>",
		Short:             "Update a topic.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validTopicArgs),
		RunE:              c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the refresh interval or retention time of Tableflow topic "my-tableflow-topic" related to Kafka cluster "lkc-123456".`,
				Code: "confluent tableflow topic update my-tableflow-topic --cluster lkc-123456 --retention-ms 432000000",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)

	cmd.Flags().String("retention-ms", "", "Specify the Tableflow table retention time in milliseconds.")
	cmd.Flags().String("table-formats", "", "Specify the table formats, one of DELTA or ICEBERG.")
	cmd.Flags().String("record-failure-strategy", "SUSPEND", "Specify the record failure strategy, one of SUSPEND or SKIP.")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	retentionMs, err := cmd.Flags().GetString("retention-ms")
	if err != nil {
		return err
	}

	tableFormats, err := cmd.Flags().GetString("table-formats")
	if err != nil {
		return err
	}
	tableFormatsSlice := []string{tableFormats}

	recordFailureStrategy, err := cmd.Flags().GetString("record-failure-strategy")
	if err != nil {
		return err
	}

	topicUpdate := tableflowv1.TableflowV1TableflowTopicUpdate{
		Spec: &tableflowv1.TableflowV1TableflowTopicSpecUpdate{
			Environment:  &tableflowv1.GlobalObjectReference{Id: environmentId},
			KafkaCluster: &tableflowv1.EnvScopedObjectReference{Id: cluster.GetId()},
			Config:       &tableflowv1.TableflowV1TableFlowTopicConfigsSpec{},
		},
	}

	if cmd.Flags().Changed("retention-ms") {
		topicUpdate.Spec.Config.SetRetentionMs(retentionMs)
	}

	if cmd.Flags().Changed("table-formats") {
		topicUpdate.Spec.SetTableFormats(tableFormatsSlice)
	}

	if cmd.Flags().Changed("record-failure-strategy") {
		topicUpdate.Spec.Config.SetRecordFailureStrategy(recordFailureStrategy)
	}

	topic, err := c.V2Client.UpdateTableflowTopic(args[0], topicUpdate)
	if err != nil {
		return fmt.Errorf("Error with updating Tableflow topic: %w", err)
	}

	return printTopicTable(cmd, topic)
}
