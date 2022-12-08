package kafka

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *authenticatedTopicCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <topic>",
		Short: "Create a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a topic named "my_topic" with default options.`,
				Code: "confluent kafka topic create my_topic",
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.Flags().Uint32("partitions", 0, "Number of topic partitions.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the topic being created.`)
	cmd.Flags().Bool("dry-run", false, "Run the command without committing changes to Kafka.")
	cmd.Flags().Bool("if-not-exists", false, "Exit gracefully if topic already exists.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *authenticatedTopicCommand) create(cmd *cobra.Command, args []string) error {
	topicName := args[0]

	partitions, err := cmd.Flags().GetUint32("partitions")
	if err != nil {
		return err
	}

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

	ifNotExists, err := cmd.Flags().GetBool("if-not-exists")
	if err != nil {
		return err
	}

	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if err := c.provisioningClusterCheck(kafkaClusterConfig.ID); err != nil {
		return err
	}

	if kafkaREST, _ := c.GetKafkaREST(); kafkaREST != nil && !dryRun {
		topicConfigs := make([]kafkarestv3.CreateTopicRequestDataConfigs, len(configMap))
		i := 0
		for key, val := range configMap {
			v := val
			topicConfigs[i] = kafkarestv3.CreateTopicRequestDataConfigs{
				Name:  key,
				Value: *kafkarestv3.NewNullableString(&v),
			}
			i++
		}

		data := kafkarestv3.CreateTopicRequestData{
			TopicName: topicName,
			Configs:   &topicConfigs,
		}

		if cmd.Flags().Changed("partitions") {
			data.PartitionsCount = utils.Int32Ptr(int32(partitions))
		}

		_, httpResp, err := kafkaREST.CloudClient.CreateKafkaTopic(kafkaClusterConfig.ID, data)
		if err != nil && httpResp != nil {
			// Kafka REST is available, but there was an error
			restErr, parseErr := kafkarest.ParseOpenAPIErrorCloud(err)
			if parseErr == nil {
				if restErr.Code == badRequestErrorCode {
					// Ignore or pretty print topic exists error
					if strings.Contains(restErr.Message, "already exists") {
						if !ifNotExists {
							return errors.NewErrorWithSuggestions(
								fmt.Sprintf(errors.TopicExistsErrorMsg, topicName, kafkaClusterConfig.ID),
								fmt.Sprintf(errors.TopicExistsSuggestions, kafkaClusterConfig.ID, kafkaClusterConfig.ID))
						}
						return nil
					}

					// Print partition limit error w/ suggestion
					if strings.Contains(restErr.Message, "partitions will exceed") {
						return errors.NewErrorWithSuggestions(restErr.Message, errors.ExceedPartitionLimitSuggestions)
					}
				}
			}
			return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
		}

		if err == nil && httpResp != nil {
			if httpResp.StatusCode != http.StatusCreated {
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusErrorMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}
			// Kafka REST is available and there was no error
			utils.Printf(cmd, errors.CreatedResourceMsg, resource.Topic, topicName)
			return nil
		}
	}

	// Kafka REST is not available, fall back to KafkaAPI

	cluster, err := dynamicconfig.KafkaCluster(c.Context)
	if err != nil {
		return err
	}

	topic := &schedv1.Topic{
		Spec: &schedv1.TopicSpecification{
			Name:              topicName,
			Configs:           configMap,
			ReplicationFactor: 3,
			NumPartitions:     6,
		},
		Validate: dryRun,
	}

	if cmd.Flags().Changed("partitions") {
		topic.Spec.NumPartitions = int32(partitions)
	}

	if err := c.PrivateClient.Kafka.CreateTopic(context.Background(), cluster, topic); err != nil {
		err = errors.CatchTopicExistsError(err, cluster.Id, topic.Spec.Name, ifNotExists)
		err = errors.CatchClusterNotReadyError(err, cluster.Id)
		return err
	}

	utils.Printf(cmd, errors.CreatedResourceMsg, resource.Topic, topic.Spec.Name)
	return nil
}
