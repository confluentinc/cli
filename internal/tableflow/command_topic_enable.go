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

func (c *command) newTopicEnableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "enable <name>",
		Aliases: []string{"create"},
		Short:   "Enable a topic.",
		Args:    cobra.ExactArgs(1),
		RunE:    c.enable,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Enable a BYOS Tableflow topic related to a Kafka cluster.",
				Code: "confluent tableflow topic enable my-tableflow-topic --cluster lkc-123456 --retention-ms 604800000 --storage-type BYOS --provider-integration cspi-stgce89r7 --bucket-name bucket_1",
			},
			examples.Example{
				Text: "Enable a confluent managed Tableflow topic related to a Kafka cluster.",
				Code: "confluent tableflow topic enable my-tableflow-topic --cluster lkc-123456 --retention-ms 604800000 --storage-type MANAGED",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)

	cmd.Flags().String("retention-ms", "604800000", "Specify the max age of snapshots (Iceberg) or versions (Delta) (snapshot/version expiration) to keep on the table in milliseconds for the Tableflow enabled topic.")
	cmd.Flags().String("storage-type", "MANAGED", "Specify the storage type of the Kafka cluster, one of MANAGED, BYOS or AzureDataLakeStorageGen2.")
	cmd.Flags().String("provider-integration", "", "Specify the provider integration id.")
	cmd.Flags().String("bucket-name", "", "Specify the name of the AWS S3 bucket.")
	cmd.Flags().String("table-formats", "ICEBERG", "Specify the table formats, one of DELTA or ICEBERG.")
	cmd.Flags().String("record-failure-strategy", "SUSPEND", "Specify the record failure strategy, one of SUSPEND or SKIP.")
	cmd.Flags().String("storage-account-name", "", "Specify the storage account name for Azure Data Lake.")
	cmd.Flags().String("container-name", "", "Specify the container name for Azure Data Lake.")
	addErrorHandlingFlags(cmd)

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	// Deprecated
	cmd.Flags().String("record-failure-strategy", "", "DEPRECATED: Specify the record failure strategy, one of SUSPEND or SKIP.")

	return cmd
}

func (c *command) enable(cmd *cobra.Command, args []string) error {
	name := args[0]

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

	storageType, err := cmd.Flags().GetString("storage-type")
	if err != nil {
		return err
	}

	providerIntegration, err := cmd.Flags().GetString("provider-integration")
	if err != nil {
		return err
	}

	bucketName, err := cmd.Flags().GetString("bucket-name")
	if err != nil {
		return err
	}

	tableFormats, err := cmd.Flags().GetString("table-formats")
	if err != nil {
		return err
	}
	tableFormatsSlice := []string{tableFormats}

	storageAccountName, err := cmd.Flags().GetString("storage-account-name")
	if err != nil {
		return err
	}

	containerName, err := cmd.Flags().GetString("container-name")
	if err != nil {
		return err
	}

	createTopic := tableflowv1.TableflowV1TableflowTopic{

		Spec: &tableflowv1.TableflowV1TableflowTopicSpec{
			DisplayName:  &name,
			TableFormats: &tableFormatsSlice,
			Environment:  &tableflowv1.GlobalObjectReference{Id: environmentId},
			KafkaCluster: &tableflowv1.EnvScopedObjectReference{Id: cluster.GetId()},
		},
	}

	createTopic.Spec.Config = &tableflowv1.TableflowV1TableFlowTopicConfigsSpec{
		RetentionMs: tableflowv1.PtrString(retentionMs),
	}

	if cmd.Flags().Changed("record-failure-strategy") {
		createTopic.Spec.Config.SetRecordFailureStrategy(recordFailureStrategy)
	}

	if cmd.Flags().Changed("error-handling") {
		if strings.ToUpper(errorHandling) == suspend {
			createTopic.Spec.Config.ErrorHandling = &tableflowv1.TableflowV1TableFlowTopicConfigsSpecErrorHandlingOneOf{
				TableflowV1ErrorHandlingSuspend: &tableflowv1.TableflowV1ErrorHandlingSuspend{
					Mode: suspend,
				},
			}
		} else if strings.ToUpper(errorHandling) == skip {
			createTopic.Spec.Config.ErrorHandling = &tableflowv1.TableflowV1TableFlowTopicConfigsSpecErrorHandlingOneOf{
				TableflowV1ErrorHandlingSkip: &tableflowv1.TableflowV1ErrorHandlingSkip{
					Mode: skip,
				},
			}
		} else if strings.ToUpper(errorHandling) == log {
			createTopic.Spec.Config.ErrorHandling = &tableflowv1.TableflowV1TableFlowTopicConfigsSpecErrorHandlingOneOf{
				TableflowV1ErrorHandlingLog: &tableflowv1.TableflowV1ErrorHandlingLog{
					Mode: log,
				},
			}
			if cmd.Flags().Changed("log-target") {
				createTopic.Spec.Config.ErrorHandling.TableflowV1ErrorHandlingLog.SetTarget(logTarget)
			}
		}
	}

	if strings.ToUpper(storageType) == "BYOS" {
		if !cmd.Flags().Changed("provider-integration") || !cmd.Flags().Changed("bucket-name") {
			return fmt.Errorf("provider-integration and bucket-name flags are required when storage-type is BYOS.")
		}

		createTopic.Spec.Storage = &tableflowv1.TableflowV1TableflowTopicSpecStorageOneOf{
			TableflowV1ByobAwsSpec: &tableflowv1.TableflowV1ByobAwsSpec{
				Kind:                  "ByobAws",
				BucketName:            bucketName,
				ProviderIntegrationId: providerIntegration,
			},
		}
	} else if strings.ToUpper(storageType) == "MANAGED" {
		createTopic.Spec.Storage = &tableflowv1.TableflowV1TableflowTopicSpecStorageOneOf{
			TableflowV1ManagedStorageSpec: &tableflowv1.TableflowV1ManagedStorageSpec{
				Kind: "Managed",
			},
		}
	} else if strings.ToUpper(storageType) == "AZUREDATALAKESTORAGEGEN2" {
		if !cmd.Flags().Changed("provider-integration") || !cmd.Flags().Changed("storage-account-name") || !cmd.Flags().Changed("container-name") {
			return fmt.Errorf("provider-integration, storage-account-name and container-name flags are required when storage-type is AzureDataLakeStorageGen2.")
		}
		createTopic.Spec.Storage = &tableflowv1.TableflowV1TableflowTopicSpecStorageOneOf{
			TableflowV1AzureAdlsSpec: &tableflowv1.TableflowV1AzureAdlsSpec{
				Kind:                  "AzureDataLakeStorageGen2",
				StorageAccountName:    storageAccountName,
				ContainerName:         containerName,
				ProviderIntegrationId: providerIntegration,
			},
		}
	} else {
		return fmt.Errorf("Unrecognized Storage Type: %s", storageType)
	}

	topic, err := c.V2Client.CreateTableflowTopic(createTopic)
	if err != nil {
		return err
	}

	return printTopicTable(cmd, topic)
}
