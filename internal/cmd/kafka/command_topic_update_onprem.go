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
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *authenticatedTopicCommand) newUpdateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <topic>",
		Short: "Update a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.onPremUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Modify the "my_topic" topic at specified cluster (providing Kafka REST Proxy endpoint) to have a retention period of 3 days (259200000 milliseconds).`,
				Code: `confluent kafka topic update my_topic --url http://localhost:8082 --config "retention.ms=259200000"`,
			},
		),
	}
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of topics configuration ("key=value") overrides for the topic being created.`)
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
	configs, err := cmd.Flags().GetStringSlice("config") // handle config parsing errors
	if err != nil {
		return err
	}
	configMap, err := properties.ConfigFlagToMap(configs)
	if err != nil {
		return err
	}

	data := make([]kafkarestv3.AlterConfigBatchRequestDataData, len(configMap))
	i := 0
	for key, val := range configMap {
		v := val
		data[i] = kafkarestv3.AlterConfigBatchRequestDataData{
			Name:  key,
			Value: &v,
		}
		i++
	}
	resp, err := restClient.ConfigsV3Api.UpdateKafkaTopicConfigBatch(restContext, clusterId, topicName,
		&kafkarestv3.UpdateKafkaTopicConfigBatchOpts{
			AlterConfigBatchRequestData: optional.NewInterface(kafkarestv3.AlterConfigBatchRequestData{Data: data}),
		})
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}
	if format == output.Human.String() {
		// no errors (config update successful)
		utils.Printf(cmd, errors.UpdateTopicConfigMsg, topicName)
		// Print Updated Configs
		tableLabels := []string{"Name", "Value"}
		tableEntries := make([][]string, len(data))
		for i, config := range data {
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
		sort.Slice(data, func(i int, j int) bool {
			return data[i].Name < data[j].Name
		})
		err = output.StructuredOutput(format, data)
		if err != nil {
			return err
		}
	}
	return nil
}
