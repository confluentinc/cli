package tableflow

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	tableflowv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/tableflow/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
)

const (
	byos    = "BYOS"
	managed = "MANAGED"
	azure   = "AzureDataLakeStorageGen2"
)

type topicOut struct {
	KafkaCluster          string `human:"Kafka Cluster" serialized:"kafka_cluster"`
	TopicName             string `human:"Topic Name" serialized:"topic_name"`
	EnableCompaction      bool   `human:"Enable Compaction" serialized:"enable_compaction"`
	EnablePartitioning    bool   `human:"Enable Partitioning" serialized:"enable_partitioning"`
	Environment           string `human:"Environment" serialized:"environment"`
	RecordFailureStrategy string `human:"Record Failure Strategy" serialized:"record_failure_strategy"`
	RetentionMs           string `human:"Retention Ms" serialized:"retention_ms"`
	StorageType           string `human:"Storage Type" serialized:"storage_type"`
	ProviderIntegrationId string `human:"Provider Integration ID,omitempty" serialized:"provider_integration_id,omitempty"`
	BucketName            string `human:"Bucket Name,omitempty" serialized:"bucket_name,omitempty"`
	BucketRegion          string `human:"Bucket Region,omitempty" serialized:"bucket_region,omitempty"`
	ContainerName         string `human:"Container Name,omitempty" serialized:"container_name,omitempty"`
	StorageAccountName    string `human:"Storage Account Name,omitempty" serialized:"storage_account_name,omitempty"`
	StorageRegion         string `human:"Storage Region,omitempty" serialized:"storage_region ,omitempty"`
	Suspended             bool   `human:"Suspended" serialized:"suspended"`
	TableFormats          string `human:"Table Formats" serialized:"table_formats"`
	TablePath             string `human:"Table Path" serialized:"table_path"`
	Phase                 string `human:"Phase" serialized:"phase"`
	ErrorMessage          string `human:"Error Message,omitempty" serialized:"error_message,omitempty"`
	WriteMode             string `human:"Write Mode,omitempty" serialized:"write_mode,omitempty"`
}

func (c *command) newTopicCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "Manage Tableflow topics.",
	}

	cmd.AddCommand(c.newTopicDescribeCommand())
	cmd.AddCommand(c.newTopicDisableCommand())
	cmd.AddCommand(c.newTopicEnableCommand())
	cmd.AddCommand(c.newTopicListCommand())
	cmd.AddCommand(c.newTopicUpdateCommand())

	return cmd
}

func (c *command) validTopicArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validTopicArgsMultiple(cmd, args)
}

func (c *command) validTopicArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteTopics()
}

func (c *command) autocompleteTopics() []string {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return nil
	}

	topics, err := c.V2Client.ListTableflowTopics(environmentId, cluster.GetId())
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(topics))
	for i, topic := range topics {
		suggestions[i] = topic.Spec.GetDisplayName()
	}
	return suggestions
}

func getStorageType(topic tableflowv1.TableflowV1TableflowTopic) (string, error) {
	config := topic.Spec.GetStorage()

	if config.TableflowV1ByobAwsSpec != nil {
		return byos, nil
	}

	if config.TableflowV1ManagedStorageSpec != nil {
		return managed, nil
	}

	if config.TableflowV1AzureAdlsSpec != nil {
		return azure, nil
	}

	return "", fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "config")
}

func printTopicTable(cmd *cobra.Command, topic tableflowv1.TableflowV1TableflowTopic) error {
	storageType, err := getStorageType(topic)
	if err != nil {
		return err
	}

	if topic.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec not found")
	}
	if topic.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status not found")
	}

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
	} else if storageType == azure {
		out.ProviderIntegrationId = topic.Spec.Storage.TableflowV1AzureAdlsSpec.GetProviderIntegrationId()
		out.ContainerName = topic.Spec.Storage.TableflowV1AzureAdlsSpec.GetContainerName()
		out.StorageAccountName = topic.Spec.Storage.TableflowV1AzureAdlsSpec.GetStorageAccountName()
		out.StorageRegion = topic.Spec.Storage.TableflowV1AzureAdlsSpec.GetStorageRegion()
		out.TablePath = topic.Spec.Storage.TableflowV1AzureAdlsSpec.GetTablePath()
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
