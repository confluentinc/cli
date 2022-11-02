package kafka

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *authenticatedTopicCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <topic>",
		Short:             "Describe a Kafka topic.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "my_topic" topic.`,
				Code: "confluent kafka topic describe my_topic",
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *authenticatedTopicCommand) describe(cmd *cobra.Command, args []string) error {
	topicName := args[0]

	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	if !output.IsValidOutputString(outputOption) {
		return output.NewInvalidOutputFormatFlagError(outputOption)
	}

	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}
	err = c.provisioningClusterCheck(kafkaClusterConfig.ID)
	if err != nil {
		return err
	}

	if kafkaREST, _ := c.GetKafkaREST(); kafkaREST != nil {
		// Get topic config
		configsResp, httpResp, err := kafkaREST.CloudClient.ListKafkaTopicConfigs(kafkaClusterConfig.ID, topicName)

		if err != nil && httpResp != nil {
			// Kafka REST is available, but there was an error
			restErr, parseErr := kafkarest.ParseOpenAPIErrorCloud(err)
			if parseErr == nil {
				if restErr.Code == unknownTopicOrPartitionErrorCode {
					return fmt.Errorf(errors.UnknownTopicErrorMsg, topicName)
				}
			}
			return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
		}

		if err == nil && httpResp != nil {
			if httpResp.StatusCode != http.StatusOK {
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusErrorMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}

			// Kafka REST is available and there was no error. Fetch partition and config information.
			topicData := &topicData{
				TopicName: topicName,
				Config:    make(map[string]string),
			}

			for _, config := range configsResp.Data {
				topicData.Config[config.Name] = config.GetValue()
			}
			numPartitions, err := c.getNumPartitions(topicName)
			if err != nil {
				return err
			}
			topicData.Config[partitionCount] = strconv.Itoa(numPartitions)

			if outputOption == output.Human.String() {
				return printHumanDescribe(topicData)
			}

			return output.StructuredOutput(outputOption, topicData)
		}
	}
	// Kafka REST is not available, fallback to KafkaAPI
	cluster, err := dynamicconfig.KafkaCluster(c.Context)
	if err != nil {
		return err
	}

	topic := &schedv1.TopicSpecification{Name: topicName}
	resp, err := c.Client.Kafka.DescribeTopic(context.Background(), cluster, &schedv1.Topic{Spec: topic, Validate: false})
	if err != nil {
		return err
	}

	if outputOption == output.Human.String() {
		return printHumanTopicDescription(resp)
	} else {
		return printStructuredTopicDescription(resp, outputOption)
	}
}

func printHumanDescribe(topicData *topicData) error {
	configsTableLabels := []string{"Name", "Value"}
	configsTableEntries := make([][]string, len(topicData.Config))
	i := 0
	for name, value := range topicData.Config {
		configsTableEntries[i] = printer.ToRow(&struct {
			name  string
			value string
		}{name: name, value: value}, []string{"name", "value"})
		i++
	}
	sort.Slice(configsTableEntries, func(i int, j int) bool {
		return configsTableEntries[i][0] < configsTableEntries[j][0]
	})
	printer.RenderCollectionTable(configsTableEntries, configsTableLabels)
	return nil
}

func printHumanTopicDescription(resp *schedv1.TopicDescription) error {
	var entries [][]string
	titleRow := []string{"Name", "Value"}
	for _, entry := range resp.Config {
		record := &struct {
			Name  string
			Value string
		}{
			entry.Name,
			entry.Value,
		}
		entries = append(entries, printer.ToRow(record, titleRow))
	}
	partitionRecord := &struct {
		Name  string
		Value string
	}{
		partitionCount,
		strconv.Itoa(len(resp.Partitions)),
	}
	entries = append(entries, printer.ToRow(partitionRecord, titleRow))
	sort.Slice(entries, func(i, j int) bool {
		return entries[i][0] < entries[j][0]
	})
	printer.RenderCollectionTable(entries, titleRow)
	return nil
}

func printStructuredTopicDescription(resp *schedv1.TopicDescription, format string) error {
	structuredDisplay := &structuredDescribeDisplay{Config: make(map[string]string)}
	structuredDisplay.TopicName = resp.Name

	for _, entry := range resp.Config {
		structuredDisplay.Config[entry.Name] = entry.Value
	}
	structuredDisplay.Config[partitionCount] = strconv.Itoa(len(resp.Partitions))
	return output.StructuredOutput(format, structuredDisplay)
}
