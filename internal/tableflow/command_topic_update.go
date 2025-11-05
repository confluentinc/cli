package tableflow

import (
	"fmt"
	"strings"

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
	addErrorHandlingFlags(cmd)

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	// Deprecated
	cmd.Flags().String("record-failure-strategy", "", "DEPRECATED: Specify the record failure strategy, one of SUSPEND or SKIP.")

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

	errorHandling, err := cmd.Flags().GetString("error-handling")
	if err != nil {
		return err
	}

	logTarget, err := cmd.Flags().GetString("log-target")
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

	if cmd.Flags().Changed("error-handling") {
		if strings.ToUpper(errorHandling) == suspend {
			topicUpdate.Spec.Config.ErrorHandling = &tableflowv1.TableflowV1TableFlowTopicConfigsSpecErrorHandlingOneOf{
				TableflowV1ErrorHandlingSuspend: &tableflowv1.TableflowV1ErrorHandlingSuspend{
					Mode: suspend,
				},
			}
		} else if strings.ToUpper(errorHandling) == skip {
			topicUpdate.Spec.Config.ErrorHandling = &tableflowv1.TableflowV1TableFlowTopicConfigsSpecErrorHandlingOneOf{
				TableflowV1ErrorHandlingSkip: &tableflowv1.TableflowV1ErrorHandlingSkip{
					Mode: skip,
				},
			}
		} else if strings.ToUpper(errorHandling) == log {
			topicUpdate.Spec.Config.ErrorHandling = &tableflowv1.TableflowV1TableFlowTopicConfigsSpecErrorHandlingOneOf{
				TableflowV1ErrorHandlingLog: &tableflowv1.TableflowV1ErrorHandlingLog{
					Mode: log,
				},
			}
			if cmd.Flags().Changed("log-target") {
				topicUpdate.Spec.Config.ErrorHandling.TableflowV1ErrorHandlingLog.SetTarget(logTarget)
			}
		}
	}

	if cmd.Flags().Changed("log-target") && !cmd.Flags().Changed("error-handling") {
		// We must check for the edge case where the current error handling mode is *not* LOG, but the user is trying to update the log target anyway
		// We should not assume that the user wants to change the mode to LOG, so we check the current mode and do nothing if it is not LOG
		currentTopic, err := c.V2Client.GetTableflowTopic(environmentId, cluster.GetId(), args[0])
		if err != nil {
			return err
		}
		if strings.ToUpper(currentTopic.GetSpec().Config.GetErrorHandling().TableflowV1ErrorHandlingLog.GetMode()) == log {
			topicUpdate.Spec.Config.ErrorHandling = &tableflowv1.TableflowV1TableFlowTopicConfigsSpecErrorHandlingOneOf{
				TableflowV1ErrorHandlingLog: &tableflowv1.TableflowV1ErrorHandlingLog{
					Mode:   log,
					Target: tableflowv1.PtrString(logTarget),
				},
			}
		}
	}

	topic, err := c.V2Client.UpdateTableflowTopic(args[0], topicUpdate)
	if err != nil {
		return fmt.Errorf("Error with updating Tableflow topic: %w", err)
	}

	return printTopicTable(cmd, topic)
}
