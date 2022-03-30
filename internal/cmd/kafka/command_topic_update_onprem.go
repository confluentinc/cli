package kafka

import (
	"sort"

	"github.com/antihax/optional"
	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *authenticatedTopicCommand) newUpdateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <topic>",
		Short: "Update a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.onPremUpdate),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Modify the "my_topic" topic at specified cluster (providing Kafka REST Proxy endpoint) to have a retention period of 3 days (259200000 milliseconds).`,
				Code: "confluent kafka topic update my_topic --url http://localhost:8082 --config=\"retention.ms=259200000\"",
			}),
	}
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags
	cmd.Flags().StringSlice("config", nil, "A comma-separated list of topics configuration ('key=value') overrides for the topic being created.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *authenticatedTopicCommand) onPremUpdate(cmd *cobra.Command, args []string) error {
	// Parse Argument
	topicName := args[0]
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	} else if !output.IsValidOutputString(format) { // catch format flag
		return output.NewInvalidOutputFormatFlagError(format)
	}
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	// Update Config
	configStrings, err := cmd.Flags().GetStringSlice("config") // handle config parsing errors
	if err != nil {
		return err
	}
	configsMap, err := properties.ToMap(configStrings)
	if err != nil {
		return err
	}
	configs := make([]kafkarestv3.AlterConfigBatchRequestDataData, len(configsMap))
	i := 0
	for k, v := range configsMap {
		v2 := v
		configs[i] = kafkarestv3.AlterConfigBatchRequestDataData{
			Name:      k,
			Value:     &v2,
			Operation: nil,
		}
		i++
	}
	resp, err := restClient.ConfigsV3Api.UpdateKafkaTopicConfigBatch(restContext, clusterId, topicName,
		&kafkarestv3.UpdateKafkaTopicConfigBatchOpts{
			AlterConfigBatchRequestData: optional.NewInterface(kafkarestv3.AlterConfigBatchRequestData{Data: configs}),
		})
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	if format == output.Human.String() {
		// no errors (config update successful)
		utils.Printf(cmd, errors.UpdateTopicConfigMsg, topicName)
		// Print Updated Configs
		tableLabels := []string{"Name", "Value"}
		tableEntries := make([][]string, len(configs))
		for i, config := range configs {
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
	} else { //json or yaml
		sort.Slice(configs, func(i int, j int) bool {
			return configs[i].Name < configs[j].Name
		})
		err = output.StructuredOutput(format, configs)
		if err != nil {
			return err
		}
	}
	return nil
}
