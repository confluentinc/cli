package kafka

import (
	"context"
	"fmt"
	"net/http"

	"github.com/antihax/optional"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
)

func (a *authenticatedTopicCommand) newCreateCommand() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create <topic>",
		Short: "Create a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(a.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a topic named `my_topic` with default options.",
				Code: "confluent kafka topic create my_topic",
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}
	createCmd.Flags().Int32("partitions", 6, "Number of topic partitions.")
	createCmd.Flags().StringSlice("config", nil, "A comma-separated list of configuration overrides ('key=value') for the topic being created.")
	createCmd.Flags().Bool("dry-run", false, "Run the command without committing changes to Kafka.")
	createCmd.Flags().Bool("if-not-exists", false, "Exit gracefully if topic already exists.")
	pcmd.AddContextFlag(createCmd, a.CLICommand)

	return createCmd
}

func (a *authenticatedTopicCommand) create(cmd *cobra.Command, args []string) error {
	topicName := args[0]

	numPartitions, err := cmd.Flags().GetInt32("partitions")
	if err != nil {
		return err
	}

	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}
	topicConfigsMap, err := utils.ToMap(configs)
	if err != nil {
		return err
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	ifNotExistsFlag, err := cmd.Flags().GetBool("if-not-exists")
	if err != nil {
		return err
	}

	kafkaREST, _ := a.GetKafkaREST()
	if kafkaREST != nil && !dryRun {
		topicConfigs := make([]kafkarestv3.CreateTopicRequestDataConfigs, len(topicConfigsMap))
		i := 0
		for k, v := range topicConfigsMap {
			val := v
			topicConfigs[i] = kafkarestv3.CreateTopicRequestDataConfigs{
				Name:  k,
				Value: &val,
			}
			i++
		}

		kafkaClusterConfig, err := a.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
		if err != nil {
			return err
		}
		lkc := kafkaClusterConfig.ID

		_, httpResp, err := kafkaREST.Client.TopicApi.ClustersClusterIdTopicsPost(kafkaREST.Context, lkc, &kafkarestv3.ClustersClusterIdTopicsPostOpts{
			CreateTopicRequestData: optional.NewInterface(kafkarestv3.CreateTopicRequestData{
				TopicName:         topicName,
				PartitionsCount:   numPartitions,
				ReplicationFactor: defaultReplicationFactor,
				Configs:           topicConfigs,
			}),
		})

		if err != nil && httpResp != nil {
			// Kafka REST is available, but there was an error
			restErr, parseErr := parseOpenAPIError(err)
			if parseErr == nil {
				if restErr.Code == KafkaRestBadRequestErrorCode {
					// Ignore or pretty print topic exists error
					if !ifNotExistsFlag {
						return errors.NewErrorWithSuggestions(
							fmt.Sprintf(errors.TopicExistsErrorMsg, topicName, lkc),
							fmt.Sprintf(errors.TopicExistsSuggestions, lkc, lkc))
					}
					return nil
				}
			}
			return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
		}

		if err == nil && httpResp != nil {
			if httpResp.StatusCode != http.StatusCreated {
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.KafkaRestUnexpectedStatusMsg, httpResp.Request.URL, httpResp.StatusCode),
					errors.InternalServerErrorSuggestions)
			}
			// Kafka REST is available and there was no error
			utils.Printf(cmd, errors.CreatedTopicMsg, topicName)
			return nil
		}
	}

	// Kafka REST is not available, fall back to KafkaAPI

	cluster, err := pcmd.KafkaCluster(a.Context)
	if err != nil {
		return err
	}

	topic := &schedv1.Topic{
		Spec: &schedv1.TopicSpecification{
			Configs: make(map[string]string)},
		Validate: false,
	}

	topic.Spec.Name = topicName
	topic.Spec.NumPartitions = numPartitions
	topic.Spec.ReplicationFactor = defaultReplicationFactor
	topic.Validate = dryRun
	topic.Spec.Configs = topicConfigsMap

	if err := a.Client.Kafka.CreateTopic(context.Background(), cluster, topic); err != nil {
		err = errors.CatchTopicExistsError(err, cluster.Id, topic.Spec.Name, ifNotExistsFlag)
		err = errors.CatchClusterNotReadyError(err, cluster.Id)
		return err
	}
	utils.Printf(cmd, errors.CreatedTopicMsg, topic.Spec.Name)
	return nil
}
