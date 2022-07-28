package kafka

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/antihax/optional"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/set"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type updateRow struct {
	Name     string
	Value    string
	ReadOnly string
}

var (
	topicListLabels           = []string{"Name", "Value", "ReadOnly"}
	topicListHumanLabels      = []string{"Name", "Value", "Read-Only"}
	topicListStructuredLabels = []string{"name", "value", "read_only"}
)

func (c *authenticatedTopicCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <topic>",
		Short:             "Update a Kafka topic.",
		Args:              cobra.ExactArgs(1),
		RunE:              c.update,
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Modify the "my_topic" topic to have a retention period of 3 days (259200000 milliseconds).`,
				Code: `confluent kafka topic update my_topic --config="retention.ms=259200000"`,
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides with form "key=value".`)
	cmd.Flags().Bool("dry-run", false, "Execute request without committing changes to Kafka.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *authenticatedTopicCommand) update(cmd *cobra.Command, args []string) error {
	topicName := args[0]

	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}
	configMap, err := properties.ConfigFlagToMap(configs)
	if err != nil {
		return err
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	if !output.IsValidOutputString(outputOption) {
		return output.NewInvalidOutputFormatFlagError(outputOption)
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST != nil && !dryRun {
		// num.partitions is read only but requires special handling
		_, numPartChange := configMap["num.partitions"]
		if numPartChange {
			delete(configMap, "num.partitions")
		}
		kafkaRestConfigs := toAlterConfigBatchRequestData(configMap)

		kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
		if err != nil {
			return err
		}
		lkc := kafkaClusterConfig.ID

		httpResp, err := kafkaREST.Client.ConfigsV3Api.UpdateKafkaTopicConfigBatch(kafkaREST.Context, lkc, topicName,
			&kafkarestv3.UpdateKafkaTopicConfigBatchOpts{
				AlterConfigBatchRequestData: optional.NewInterface(kafkarestv3.AlterConfigBatchRequestData{Data: kafkaRestConfigs}),
			})

		if err != nil && httpResp != nil {
			// Kafka REST is available, but an error occurred
			restErr, parseErr := parseOpenAPIError(err)
			if parseErr == nil {
				if restErr.Code == KafkaRestUnknownTopicOrPartitionErrorCode {
					return fmt.Errorf(errors.UnknownTopicErrorMsg, topicName)
				}
			}
			return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
		}

		if err == nil && httpResp != nil {
			if httpResp.StatusCode != http.StatusNoContent {
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}

			// Kafka REST is available and there was no error
			configsResp, httpResp, err := kafkaREST.Client.ConfigsV3Api.ListKafkaTopicConfigs(kafkaREST.Context, lkc, topicName)
			if err != nil {
				return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
			} else if configsResp.Data == nil {
				return errors.NewErrorWithSuggestions(errors.EmptyResponseMsg, errors.InternalServerErrorSuggestions)
			}
			readOnlyConfigs := set.New()
			configsValues := make(map[string]string)
			for _, conf := range configsResp.Data {
				if conf.IsReadOnly {
					readOnlyConfigs.Add(conf.Name)
				}
				configsValues[conf.Name] = *conf.Value
			}

			// Output writing preparation
			outputWriter, err := output.NewListOutputWriter(cmd, topicListLabels, topicListHumanLabels, topicListStructuredLabels)
			if err != nil {
				return err
			}
			if numPartChange {
				partitionsResp, httpResp, err := kafkaREST.Client.PartitionV3Api.ListKafkaPartitions(kafkaREST.Context, lkc, topicName)
				if err != nil && httpResp != nil {
					restErr, parseErr := parseOpenAPIError(err)
					if parseErr == nil {
						if restErr.Code == KafkaRestUnknownTopicOrPartitionErrorCode {
							return fmt.Errorf(errors.UnknownTopicErrorMsg, topicName)
						}
					}
					return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
				}

				readOnlyConfigs.Add("num.partitions")
				configsValues["num.partitions"] = strconv.Itoa(len(partitionsResp.Data))
				// Add num.partitions back into kafkaRestConfig for sorting & output
				partitionsKafkaRestConfig := kafkarestv3.AlterConfigBatchRequestDataData{
					Name:      "num.partitions",
					Value:     nil,
					Operation: nil,
				}
				kafkaRestConfigs = append(kafkaRestConfigs, partitionsKafkaRestConfig)
			}
			sort.Slice(kafkaRestConfigs, func(i, j int) bool {
				return kafkaRestConfigs[i].Name < kafkaRestConfigs[j].Name
			})

			// Write current state of relevant config settings
			utils.Printf(cmd, errors.UpdateTopicConfigRESTMsg, topicName)
			for _, config := range kafkaRestConfigs {
				row := updateRow{
					Name:     config.Name,
					Value:    configsValues[config.Name],
					ReadOnly: strconv.FormatBool(readOnlyConfigs[config.Name]),
				}
				outputWriter.AddElement(row)
			}
			return outputWriter.Out()
		}
	}

	// Kafka REST is not available, fallback to KafkaAPI
	cluster, err := dynamicconfig.KafkaCluster(c.Context)
	if err != nil {
		return err
	}

	topic := &schedv1.TopicSpecification{Name: args[0], Configs: copyMap(configMap)}

	err = c.Client.Kafka.UpdateTopic(context.Background(), cluster, &schedv1.Topic{Spec: topic, Validate: dryRun})
	if err != nil {
		return errors.CatchClusterNotReadyError(err, cluster.Id)
	}
	utils.Printf(cmd, errors.UpdateTopicConfigMsg, args[0])
	var entries [][]string
	titleRow := []string{"Name", "Value"}
	for name, value := range configMap {
		record := &struct {
			Name  string
			Value string
		}{
			name,
			value,
		}
		entries = append(entries, printer.ToRow(record, titleRow))
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i][0] < entries[j][0]
	})
	printer.RenderCollectionTable(entries, titleRow)
	return nil
}
