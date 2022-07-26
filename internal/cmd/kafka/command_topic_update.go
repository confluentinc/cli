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
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/set"
	"github.com/confluentinc/cli/internal/pkg/utils"
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
			configsReadOnly := set.New()
			readOnlyValue := make(map[string]string)
			for _, conf := range configsResp.Data {
				if conf.IsReadOnly {
					configsReadOnly.Add(conf.Name)
					readOnlyValue[conf.Name] = *conf.Value
				}
			}

			utils.Printf(cmd, errors.UpdateTopicConfigMsg, topicName)
			tableLabels := []string{"Name", "Value", "Read Only"}
			tableEntries := make([][]string, len(kafkaRestConfigs))
			for i, config := range kafkaRestConfigs {
				valString := *config.Value
				readOnlyString := "No"
				if configsReadOnly[config.Name] {
					valString = readOnlyValue[config.Name]
					readOnlyString = "Yes"
				}
				tableEntries[i] = printer.ToRow(
					&struct {
						Name     string
						Value    string
						ReadOnly string
					}{Name: config.Name, Value: valString, ReadOnly: readOnlyString}, []string{"Name", "Value", "ReadOnly"})
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

				tableEntries = append(tableEntries, printer.ToRow(
					&struct {
						Name     string
						Value    string
						ReadOnly string
					}{Name: "num.partitions", Value: strconv.Itoa(len(partitionsResp.Data)), ReadOnly: "Yes"}, []string{"Name", "Value", "ReadOnly"}))
			}
			sort.Slice(tableEntries, func(i int, j int) bool {
				return tableEntries[i][0] < tableEntries[j][0]
			})
			printer.RenderCollectionTable(tableEntries, tableLabels)
			return nil
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
