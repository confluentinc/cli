package tableflow

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newTopicListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Tableflow topics related to one Kafka cluster.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	topics, err := c.V2Client.ListTableflowTopics(environmentId, cluster.GetId())
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, topic := range topics {
		storageType, err := getStorageType(topic)
		if err != nil {
			return err
		}

		strStatus := getCatalogSyncStatuses(topic.Status.GetCatalogSyncStatuses())
		strFormats := getFailingTableFormats(topic.Status.GetFailingTableFormats())

		out := &topicOut{
			KafkaCluster:          topic.GetSpec().KafkaCluster.GetId(),
			TopicName:             topic.Spec.GetDisplayName(),
			EnableCompaction:      topic.GetSpec().Config.GetEnableCompaction(),   // should be read-only & true
			EnablePartitioning:    topic.GetSpec().Config.GetEnablePartitioning(), // should be read-only & true
			TableFormats:          strings.Join(topic.Spec.GetTableFormats(), ""),
			Environment:           topic.GetSpec().Environment.GetId(),
			RetentionMs:           topic.GetSpec().Config.GetRetentionMs(),
			RecordFailureStrategy: topic.GetSpec().Config.GetRecordFailureStrategy(),
			StorageType:           storageType,
			Suspended:             topic.Spec.GetSuspended(),
			Phase:                 topic.Status.GetPhase(),
			CatalogSyncStatus:     strStatus,
			FailingTableFormat:    strFormats,
			ErrorMessage:          topic.Status.GetErrorMessage(),
			WriteMode:             topic.Status.GetWriteMode(),
		}

		if storageType == byos {
			out.BucketName = topic.Spec.Storage.TableflowV1ByobAwsSpec.GetBucketName()
			out.BucketRegion = topic.Spec.Storage.TableflowV1ByobAwsSpec.GetBucketRegion()
			out.ProviderIntegrationId = topic.Spec.Storage.TableflowV1ByobAwsSpec.GetProviderIntegrationId()
			out.TablePath = topic.Spec.Storage.TableflowV1ByobAwsSpec.GetTablePath()
		} else if storageType == managed {
			out.TablePath = topic.Spec.Storage.TableflowV1ManagedStorageSpec.GetTablePath()
		}

		list.Add(out)
	}

	return list.PrintWithAutoWrap(false)
}
