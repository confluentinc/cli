package kafka

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/set"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type topicConfigurationOut struct {
	Name     string `human:"Name" serialized:"name"`
	Value    string `human:"Value" serialized:"value"`
	ReadOnly string `human:"Read-Only" serialized:"read_only"`
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

	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}
	err = c.provisioningClusterCheck(kafkaClusterConfig.ID)
	if err != nil {
		return err
	}

	if kafkaREST, _ := c.GetKafkaREST(); kafkaREST != nil && !dryRun {
		// num.partitions is read only but requires special handling
		_, hasNumPartitionsChanged := configMap["num.partitions"]
		if hasNumPartitionsChanged {
			delete(configMap, "num.partitions")
		}
		kafkaRestConfigs := toAlterConfigBatchRequestData(configMap)

		data := toAlterConfigBatchRequestData(configMap)
		httpResp, err := kafkaREST.CloudClient.UpdateKafkaTopicConfigBatch(kafkaClusterConfig.ID, topicName, data)

		if err != nil && httpResp != nil {
			// Kafka REST is available, but an error occurred
			restErr, parseErr := kafkarest.ParseOpenAPIErrorCloud(err)
			if parseErr == nil {
				if restErr.Code == unknownTopicOrPartitionErrorCode {
					return fmt.Errorf(errors.UnknownTopicErrorMsg, topicName)
				}
			}
			return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
		}

		if err == nil && httpResp != nil {
			if httpResp.StatusCode != http.StatusNoContent {
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusErrorMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}

			// Kafka REST is available and there was no error
			configsResp, httpResp, err := kafkaREST.CloudClient.ListKafkaTopicConfigs(kafkaClusterConfig.ID, topicName)
			if err != nil {
				return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
			} else if configsResp.Data == nil {
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

				readOnlyConfigs.Add("num.partitions")
				configsValues["num.partitions"] = strconv.Itoa(numPartitions)
				// Add num.partitions back into kafkaRestConfig for sorting & output
				partitionsKafkaRestConfig := kafkarestv3.AlterConfigBatchRequestDataData{Name: "num.partitions"}
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
					ReadOnly: strconv.FormatBool(readOnlyConfigs[config.Name]),
				})
			}
			return list.Print()
		}
	}

	// Kafka REST is not available, fallback to KafkaAPI
	cluster, err := dynamicconfig.KafkaCluster(c.Context)
	if err != nil {
		return err
	}

	topic := &schedv1.Topic{
		Spec: &schedv1.TopicSpecification{
			Name:    args[0],
			Configs: copyMap(configMap),
		},
		Validate: dryRun,
	}
	if err := c.PrivateClient.Kafka.UpdateTopic(context.Background(), cluster, topic); err != nil {
		return errors.CatchClusterNotReadyError(err, cluster.Id)
	}

	if output.GetFormat(cmd) == output.Human {
		utils.Printf(cmd, errors.UpdateTopicConfigMsg, args[0])
	}
	list := output.NewList(cmd)
	for name, value := range configMap {
		list.Add(&topicConfigurationOut{
			Name:  name,
			Value: value,
		})
	}
	list.Filter([]string{"Name", "Value"})
	return list.Print()
}
