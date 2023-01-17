package kafka

import (
	"fmt"
	"strconv"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/set"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type topicConfigurationOut struct {
	Name     string `human:"Name" serialized:"name"`
	Value    string `human:"Value" serialized:"value"`
	ReadOnly bool   `human:"Read-Only" serialized:"read_only"`
}

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
				Code: `confluent kafka topic update my_topic --config "retention.ms=259200000"`,
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides with form "key=value".`)
	cmd.Flags().Bool("dry-run", false, "Run the command without committing changes to Kafka.")
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

	kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if err := c.provisioningClusterCheck(kafkaClusterConfig.ID); err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	// num.partitions is read-only but requires special handling
	_, hasNumPartitionsChanged := configMap[numPartitionsKey]
	if hasNumPartitionsChanged {
		delete(configMap, numPartitionsKey)
	}
	kafkaRestConfigs := toAlterConfigBatchRequestData(configMap)

	data := toAlterConfigBatchRequestData(configMap)
	data.ValidateOnly = &dryRun

	httpResp, err := kafkaREST.CloudClient.UpdateKafkaTopicConfigBatch(kafkaClusterConfig.ID, topicName, data)
	if err != nil {
		restErr, parseErr := kafkarest.ParseOpenAPIErrorCloud(err)
		if parseErr == nil {
			if restErr.Code == ccloudv2.UnknownTopicOrPartitionErrorCode {
				return fmt.Errorf(errors.UnknownTopicErrorMsg, topicName)
			}
		}
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	if dryRun {
		utils.Printf(cmd, errors.UpdatedResourceMsg, resource.Topic, topicName)
		return nil
	}

	configsResp, err := kafkaREST.CloudClient.ListKafkaTopicConfigs(kafkaClusterConfig.ID, topicName)
	if err != nil {
		return err
	}
	if configsResp.Data == nil {
		return errors.NewErrorWithSuggestions(errors.EmptyResponseErrorMsg, errors.InternalServerErrorSuggestions)
	}

	readOnlyConfigs := set.New()
	configsValues := make(map[string]string)
	for _, conf := range configsResp.Data {
		if conf.IsReadOnly {
			readOnlyConfigs.Add(conf.Name)
		}
		configsValues[conf.Name] = conf.GetValue()
	}

	if hasNumPartitionsChanged {
		numPartitions, err := c.getNumPartitions(topicName)
		if err != nil {
			return err
		}

		readOnlyConfigs.Add(numPartitionsKey)
		configsValues[numPartitionsKey] = strconv.Itoa(numPartitions)
		// Add num.partitions back into kafkaRestConfig for sorting & output
		partitionsKafkaRestConfig := kafkarestv3.AlterConfigBatchRequestDataData{Name: numPartitionsKey}
		kafkaRestConfigs.Data = append(kafkaRestConfigs.Data, partitionsKafkaRestConfig)
	}

	// Write current state of relevant config settings
	if output.GetFormat(cmd) == output.Human {
		utils.ErrPrintf(cmd, errors.UpdateTopicConfigRestMsg, topicName)
	}

	list := output.NewList(cmd)
	for _, config := range kafkaRestConfigs.Data {
		list.Add(&topicConfigurationOut{
			Name:     config.Name,
			Value:    configsValues[config.Name],
			ReadOnly: readOnlyConfigs[config.Name],
		})
	}
	return list.Print()
}
