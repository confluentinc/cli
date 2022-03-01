package kafka

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/antihax/optional"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
)

func (c *authenticatedTopicCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <topic>",
		Short:             "Update a Kafka topic.",
		Args:              cobra.ExactArgs(1),
		RunE:              pcmd.NewCLIRunE(c.update),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Modify the "my_topic" topic to have a retention period of 3 days (259200000 milliseconds).`,
				Code: `confluent kafka topic update my_topic --config="retention.ms=259200000"`,
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}
	cmd.Flags().StringSlice("config", nil, "A comma-separated list of topics. Configuration ('key=value') overrides for the topic being created.")
	cmd.Flags().Bool("dry-run", false, "Execute request without committing changes to Kafka.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *authenticatedTopicCommand) update(cmd *cobra.Command, args []string) error {
	topicName := args[0]

	configStrings, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}
	configsMap, err := utils.ToMap(configStrings)
	if err != nil {
		return err
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST != nil && !dryRun {
		kafkaRestConfigs := toAlterConfigBatchRequestData(configsMap)

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
			utils.Printf(cmd, errors.UpdateTopicConfigMsg, topicName)
			tableLabels := []string{"Name", "Value"}
			tableEntries := make([][]string, len(kafkaRestConfigs))
			for i, config := range kafkaRestConfigs {
				tableEntries[i] = printer.ToRow(
					&struct {
						Name  string
						Value string
					}{Name: config.Name, Value: *config.Value}, []string{"Name", "Value"})
			}
			sort.Slice(tableEntries, func(i int, j int) bool {
				return tableEntries[i][0] < tableEntries[j][0]
			})
			printer.RenderCollectionTable(tableEntries, tableLabels)
			return nil
		}
	}

	// Kafka REST is not available, fallback to KafkaAPI
	cluster, err := pcmd.KafkaCluster(c.Context)
	if err != nil {
		return err
	}

	topic := &schedv1.TopicSpecification{Name: args[0], Configs: make(map[string]string)}

	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}

	configMap, err := utils.ToMap(configs)
	if err != nil {
		return err
	}
	topic.Configs = copyMap(configMap)

	err = c.Client.Kafka.UpdateTopic(context.Background(), cluster, &schedv1.Topic{Spec: topic, Validate: dryRun})
	if err != nil {
		err = errors.CatchClusterNotReadyError(err, cluster.Id)
		return err
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
