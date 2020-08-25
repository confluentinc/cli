package kafka

// confluent kafka topic <commands>
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	purl "net/url"
	"sort"
	"strings"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafka"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
)

// Info needed to complete kafka topic ...
type topicCommand struct {
	*pcmd.UseKafkaRestCLICommand
	prerunner pcmd.PreRunner
}

// Return the command to be registered to the kafka topic slot
func NewTopicCommandOnPrem(prerunner pcmd.PreRunner) *cobra.Command {
	topicCmd := &topicCommand{
		UseKafkaRestCLICommand: pcmd.NewUseKafkaRestCLICommand(
			&cobra.Command{
				Use:   "topic",
				Short: "Manage Kafka topics.",
			}, prerunner),
		prerunner: prerunner,
	}

	topicCmd.init()
	return topicCmd.Command
}

// Register each of the verbs and expected args
func (topicCmd *topicCommand) init() {
	// Register list command
	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(topicCmd.listTopics),
		Short: "List Kafka topics.",
		Example: examples.BuildExampleString(
			examples.Example{
				// on-prem examples are ccloud examples + "of a specified cluster (providing Kafka REST Proxy endpoint)."
				Text: "List all topics of a specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "confluent kafka topic list --url http://localhost:8082",
			},
		),
	}
	listCmd.Flags().String("url", "", "Base URL of REST Proxy Endpoint of Kafka Cluster.")
	check(listCmd.MarkFlagRequired("url")) // TODO: unset url as required
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	topicCmd.AddCommand(listCmd)

	// Register create command
	// Create a Kafka topic.
	// Usage:
	//   confluent kafka topic create <topic> [flags]
	// Flags:
	//   --url string Base URL of REST Proxy Endpoint of Kafka Cluster. (Required) (replace: cluster string      Kafka cluster ID)
	//   --partitions uint32   Number of topic partitions. (default 6)
	//   --replication-factor uint32 Number of replicas. (default 3)  (new)
	//   --config strings      A comma-separated list of topics configuration ('key=value') overrides for the topic being created.
	//   --if-not-exists       Exit gracefully if topic already exists.
	createCmd := &cobra.Command{
		Use:   "create <topic>",
		Short: "Create a Kafka topic.",
		Args:  cobra.ExactArgs(1), // <topic>
		RunE:  pcmd.NewCLIRunE(topicCmd.createTopic),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a topic named ``my_topic`` with default options at specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "confluent kafka topic create my_topic --url http://localhost:8082",
			},
			// TODO: this is not included in the ccloud examples, but I like how it demonstrates how the "StringSlice" is inputted.
			examples.Example{
				Text: "Create a topic named ``my_topic_2`` using the config flag.", // TODO: better text
				Code: "confluent kafka topic create my_topic_2 --url http://localhost:8082 --config cleanup.policy=compact,compression.type=gzip",
			}),
	}
	createCmd.Flags().String("url", "", "Base URL of REST Proxy Endpoint of Kafka Cluster.")
	check(listCmd.MarkFlagRequired("url"))
	createCmd.Flags().Int32("partitions", 6, "Number of topic partitions.")
	createCmd.Flags().Int32("replication-factor", 3, "Number of replicas.")
	createCmd.Flags().StringSlice("config", nil, "A comma-separated list of topic configuration ('key=value') overrides for the topic being created.")
	createCmd.Flags().Bool("if-not-exists", false, "Exit gracefully if topic already exists.")
	createCmd.Flags().SortFlags = false
	topicCmd.AddCommand(createCmd)

	// Register delete command
	// Delete a Kafka topic.
	// Usage:
	//     confluent kafka topic delete <topic> [flags]
	// Examples:
	//     Delete the topics ``my_topic`` and ``my_topic_avro``. Use this command carefully as data loss can occur.
	//     confluent kafka topic delete my_topic
	//     confluent kafka topic delete my_topic_avro
	// Flags:
	//     --url Base URL of REST Proxy Endpoint of Kafka Cluster. (Required) (replace: cluster string Kafka cluster ID)
	deleteCmd := &cobra.Command{
		Use:   "delete <topic>",
		Short: "Delete a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(topicCmd.deleteTopic),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete the topic ``my_topic`` at specified cluster (providing Kafka REST Proxy endpoint). Use this command carefully as data loss can occur.",
				Code: "confluent kafka topic delete my_topic --url http://localhost:8082",
			}),
	}
	deleteCmd.Flags().String("url", "", "Base URL of REST Proxy Endpoint of Kafka Cluster.")
	check(deleteCmd.MarkFlagRequired("url"))
	deleteCmd.Flags().SortFlags = false
	topicCmd.AddCommand(deleteCmd)

	// Register update command
	// Update a Kafka topic.
	// Usage:
	// 	confluent kafka topic update <topic> [flags]
	// Examples:
	// 	Modify the ``my_topic`` topic at specified cluster (providing Kafka REST Proxy endpoint) to have a retention period of 3 days (259200000 milliseconds).
	// 		confluent kafka topic update my_topic --url http://localhost:8082 --config="retention.ms=259200000"
	// Flags:
	// 	--url string   Base URL of REST Proxy Endpoint of Kafka Cluster.
	// 	--config strings   A comma-separated list of topics configuration ('key=value') overrides for the topic being created.
	updateCmd := &cobra.Command{
		Use:   "update <topic>",
		Short: "Update a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(topicCmd.updateTopicConfig),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Modify the ``my_topic`` topic at specified cluster (providing Kafka REST Proxy endpoint) to have a retention period of 3 days (259200000 milliseconds).",
				Code: "confluent kafka topic update my_topic --url http://localhost:8082 --config=\"retention.ms=259200000\"",
			}),
	}
	updateCmd.Flags().String("url", "", "Base URL of REST Proxy Endpoint of Kafka Cluster.")
	check(updateCmd.MarkFlagRequired("url"))
	updateCmd.Flags().StringSlice("config", nil, "A comma-separated list of topics configuration ('key=value') overrides for the topic being created.")
	updateCmd.Flags().SortFlags = false
	topicCmd.AddCommand(updateCmd)

	// Register describe command
	// Describe a Kafka topic.
	// Usage:
	// confluent kafka topic describe <topic> [flags]
	// Examples:
	// Describe the ``my_topic`` topic.
	// confluent kafka topic describe my_topic
	// Flags:
	//  --cluster string   Kafka cluster ID.
	// -o, --output string    Specify the output format as "human", "json", or "yaml". (default "human")
	describeCmd := &cobra.Command{
		Use:   "describe <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(topicCmd.describeTopic),
		Short: "Describe a Kafka topic.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the ``my_topic`` topic at specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "confluent kafka topic describe my_topic --url http://localhost:8082",
			},
		),
	}
	describeCmd.Flags().String("url", "", "Base URL of REST Proxy Endpoint of Kafka Cluster.")
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	topicCmd.AddCommand(describeCmd)
}

func setServerURL(client *kafkarestv3.APIClient, url string) {
	client.ChangeBasePath(strings.Trim(url, "/") + "/v3")
}

type kafkaRestV3Error struct {
	Code    int    `json:"error_code"`
	Message string `json:"message"`
}

func handleCommonKafkaRestClientErrors(url string, kafkaRestClient *kafkarestv3.APIClient, resp *http.Response, err error) error {
	switch err.(type) {
	case *purl.Error: // Handle errors with request url
		if e, ok := err.(*purl.Error); ok {
			// TODO: Currently this error exposes implementation detail
			// TODO: add message: cluster may not be ready?
			return errors.Errorf(errors.InvalidFlagValueWithWrappedErrorErrorMsg, url, "url", e.Err)
		}
	case kafkarestv3.GenericOpenAPIError:
		if openAPIError, ok := err.(kafkarestv3.GenericOpenAPIError); ok {
			var decodedError kafkaRestV3Error
			err = json.Unmarshal(openAPIError.Body(), &decodedError)
			if err != nil {
				return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
			}
			return fmt.Errorf("Kafka REST Proxy backend error:\n\t%v", decodedError.Message)
		}
	}
	return err
}

// listTopics - Registered as RunE of kafka topic list
// * @param cmd: cobra.Command matched by command line arguments
// * @param args: The rest of command line arguments (os.args[1:]
func (topicCmd *topicCommand) listTopics(cmd *cobra.Command, args []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil { // require the flag
		return err
	}

	setServerURL(topicCmd.KafkaRestClient, url)
	kafkaRestClient := topicCmd.KafkaRestClient

	// Get Cluster Id
	clusters, resp, err := kafkaRestClient.ClusterApi.ClustersGet(context.Background())
	if err != nil {
		return handleCommonKafkaRestClientErrors(url, kafkaRestClient, resp, err)
	}
	clusterId := clusters.Data[0].ClusterId

	// Get Topics
	topicGetResp, resp, err := kafkaRestClient.TopicApi.ClustersClusterIdTopicsGet(context.Background(), clusterId)
	if err != nil {
		return handleCommonKafkaRestClientErrors(url, kafkaRestClient, resp, err)
	}
	topicDatas := topicGetResp.Data

	// Create and populate output writer
	outputWriter, err := output.NewListOutputWriter(cmd, []string{"TopicName"}, []string{"Name"}, []string{"name"})
	if err != nil {
		return err
	}
	for _, topicData := range topicDatas {
		outputWriter.AddElement(&topicData)
	}

	return outputWriter.Out()
}

// Register create command
// Create a Kafka topic.
// Usage:
//   confluent kafka topic create <topic> [flags]
// Flags:
//   --url string Base URL of REST Proxy Endpoint of Kafka Cluster. (Required) (replace: cluster string      Kafka cluster ID)
//   --partitions uint32   Number of topic partitions. (default 6)
//   --replication-factor uint32 Number of replicas. (default 3)  (new)
//   --config strings      A comma-separated list of topics configuration ('key=value') overrides for the topic being created.
//   --if-not-exists       Exit gracefully if topic already exists.
func (topicCmd *topicCommand) createTopic(cmd *cobra.Command, args []string) error {
	// Parse arguments
	topicName := args[0]
	// check required argument
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}

	// Setup APIClient
	setServerURL(topicCmd.KafkaRestClient, url)
	kafkaRestClient := topicCmd.KafkaRestClient

	// Get Cluster Id
	clusters, resp, err := kafkaRestClient.ClusterApi.ClustersGet(context.Background())
	if err != nil {
		return handleCommonKafkaRestClientErrors(url, kafkaRestClient, resp, err) // handle url errors
	} else if clusters.Data == nil || len(clusters.Data) == 0 {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
	}
	clusterId := clusters.Data[0].ClusterId

	// Parse remaining arguments
	numPartitions, err := cmd.Flags().GetInt32("partitions")
	if err != nil {
		return err
	}
	replicationFactor, err := cmd.Flags().GetInt32("replication-factor")
	if err != nil {
		return err
	}
	topicConfigStrings, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}
	ifNotExists, err := cmd.Flags().GetBool("if-not-exists")
	if err != nil {
		return err
	}

	topicConfigsMap, err := kafka.ToMap(topicConfigStrings)
	if err != nil {
		return err
	}
	topicConfigs := make([]kafkarestv3.CreateTopicRequestDataConfigs, len(topicConfigsMap))
	i := 0
	for k, v := range topicConfigsMap {
		v2 := v // create a local copy to use pointer
		topicConfigs[i] = kafkarestv3.CreateTopicRequestDataConfigs{
			Name:  k,
			Value: &v2,
		}
		i++
	}

	// Create new topic
	_, resp, err = kafkaRestClient.TopicApi.ClustersClusterIdTopicsPost(context.Background(), clusterId, &kafkarestv3.ClustersClusterIdTopicsPostOpts{
		CreateTopicRequestData: optional.NewInterface(kafkarestv3.CreateTopicRequestData{
			TopicName:         topicName,
			PartitionsCount:   numPartitions,
			ReplicationFactor: replicationFactor,
			Configs:           topicConfigs,
		}),
	})
	if err != nil {
		// catch topic exists error
		if openAPIError, ok := err.(kafkarestv3.GenericOpenAPIError); ok {
			var decodedError kafkaRestV3Error
			err2 := json.Unmarshal(openAPIError.Body(), &decodedError)
			if err2 != nil {
				return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
			}
			if decodedError.Message == fmt.Sprintf("Topic '%s' already exists.", topicName) {
				if ifNotExists == false {
					return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.TopicExistsOnPremErrorMsg, topicName), errors.TopicExistsOnPremSuggestions)
				} // ignore error if ifNotExists flag is set
				return nil
			}
		}
		return handleCommonKafkaRestClientErrors(url, kafkaRestClient, resp, err) // catch all other errors
	}
	// no error if topic is created successfully.
	pcmd.ErrPrintf(cmd, errors.CreatedTopicMsg, topicName) // TODO: why print to StdErr
	return nil
}

// Register delete command
// Delete a Kafka topic.
// Usage:
//     confluent kafka topic delete <topic> [flags]
// Flags:
//     --url Base URL of REST Proxy Endpoint of Kafka Cluster. (Required) (replace: cluster string Kafka cluster ID)
func (topicCmd *topicCommand) deleteTopic(cmd *cobra.Command, args []string) error {
	// Parse arguments
	topicName := args[0]
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}

	// Get ClusterId
	setServerURL(topicCmd.KafkaRestClient, url)
	kafkaRestClient := topicCmd.KafkaRestClient
	clustersData, resp, err := kafkaRestClient.ClusterApi.ClustersGet(context.Background())
	if err != nil {
		return handleCommonKafkaRestClientErrors(url, kafkaRestClient, resp, err) // checks for error in URL
	} else if clustersData.Data == nil || len(clustersData.Data) == 0 {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
	}
	clusterId := clustersData.Data[0].ClusterId

	// Delete Topic
	resp, err = kafkaRestClient.TopicApi.ClustersClusterIdTopicsTopicNameDelete(context.Background(), clusterId, topicName)
	if err != nil {
		return handleCommonKafkaRestClientErrors(url, kafkaRestClient, resp, err) // catches topic name not found (backend error)
	}
	pcmd.ErrPrintf(cmd, errors.DeletedTopicMsg, topicName) // topic successfully created
	return nil
}

func (topicCmd *topicCommand) updateTopicConfig(cmd *cobra.Command, args []string) error {
	// Parse Argument
	topicName := args[0]
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}

	// Get Cluster Id
	setServerURL(topicCmd.KafkaRestClient, url)
	kafkaRestClient := topicCmd.KafkaRestClient
	clustersData, resp, err := kafkaRestClient.ClusterApi.ClustersGet(context.Background())
	if err != nil {
		return handleCommonKafkaRestClientErrors(url, kafkaRestClient, resp, err) // handle URL error
	} else if clustersData.Data == nil || len(clustersData.Data) == 0 {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
	}
	clusterId := clustersData.Data[0].ClusterId

	// Update Config
	configStrings, err := cmd.Flags().GetStringSlice("config") // handle config parsing errors
	if err != nil {
		return err
	}
	configsMap, err := kafka.ToMap(configStrings)
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
	resp, err = kafkaRestClient.ConfigsApi.ClustersClusterIdTopicsTopicNameConfigsalterPost(context.Background(), clusterId, topicName,
		&kafkarestv3.ClustersClusterIdTopicsTopicNameConfigsalterPostOpts{
			AlterConfigBatchRequestData: optional.NewInterface(kafkarestv3.AlterConfigBatchRequestData{Data: configs}),
		})
	if err != nil {
		return handleCommonKafkaRestClientErrors(url, kafkaRestClient, resp, err) // handle config key/value invalid errors
	}
	// no errors (config update successful)
	pcmd.Printf(cmd, errors.UpdateTopicConfigMsg, topicName)
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
	return nil
}

type PartitionData struct {
	TopicName              string  `json:"topic" yaml:"topic"`
	PartitionId            int32   `json:"partition" yaml:"partition"`
	LeaderBrokerId         int32   `json:"leader" yaml:"leader"`
	ReplicaBrokerIds       []int32 `json:"replicas" yaml:"replicas"`
	InSyncReplicaBrokerIds []int32 `json:"isr" yaml:"isr"`
}

type TopicData struct {
	TopicName         string            `json:"topic_name" yaml:"topic_name"`
	PartitionCount    int               `json:"partition_count" yaml:"partition_count"`
	ReplicationFactor int               `json:"replication_factor" yaml:"replication_factor"`
	Partitions        []PartitionData   `json:"partitions" yaml:"partitions"`
	Configs           map[string]string `json:"config" yaml:"config"`
}

func (topicCmd *topicCommand) describeTopic(cmd *cobra.Command, args []string) error {
	// Parse Args
	topicName := args[0]
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	} else if output.IsValidFormat(format) == false { // catch format flag
		return output.NewInvalidOutputFormatFlagError(format)
	}

	// Get clusterId
	setServerURL(topicCmd.KafkaRestClient, url)
	client := topicCmd.KafkaRestClient
	clustersData, resp, err := client.ClusterApi.ClustersGet(context.Background())
	if err != nil {
		return handleCommonKafkaRestClientErrors(url, client, resp, err) // catch url incorrect error
	} else if clustersData.Data == nil || len(clustersData.Data) == 0 {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
	}
	clusterId := clustersData.Data[0].ClusterId

	// Get partitions
	topicData := &TopicData{}
	// TODO: partitions reassignment?
	partitionsResp, resp, err := client.PartitionApi.ClustersClusterIdTopicsTopicNamePartitionsGet(context.Background(), clusterId, topicName)
	if err != nil {
		return handleCommonKafkaRestClientErrors(url, client, resp, err) // catch topic not exist error
	} else if partitionsResp.Data == nil {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
	}
	topicData.TopicName = topicName
	topicData.PartitionCount = len(partitionsResp.Data)
	topicData.Partitions = make([]PartitionData, len(partitionsResp.Data))
	for i, partitionResp := range partitionsResp.Data {
		partitionId := partitionResp.PartitionId
		partitionData := PartitionData{
			TopicName:   topicName,
			PartitionId: partitionId,
		}

		// For each partition, get replicas
		replicasResp, resp, err := client.ReplicaApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicasGet(context.Background(), clusterId, topicName, partitionId)
		if err != nil {
			return handleCommonKafkaRestClientErrors(url, client, resp, err)
		} else if replicasResp.Data == nil {
			return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
		}
		partitionData.ReplicaBrokerIds = make([]int32, len(replicasResp.Data))
		partitionData.InSyncReplicaBrokerIds = make([]int32, 0, len(replicasResp.Data))
		for j, replicaResp := range replicasResp.Data {
			if replicaResp.IsLeader {
				partitionData.LeaderBrokerId = replicaResp.BrokerId
			}
			partitionData.ReplicaBrokerIds[j] = replicaResp.BrokerId
			if replicaResp.IsInSync {
				partitionData.InSyncReplicaBrokerIds = append(partitionData.InSyncReplicaBrokerIds, replicaResp.BrokerId)
			}
		}
		if i == 0 {
			topicData.ReplicationFactor = len(replicasResp.Data)
		}
		topicData.Partitions[i] = partitionData
	}

	// Get configs
	configsResp, resp, err := client.ConfigsApi.ClustersClusterIdTopicsTopicNameConfigsGet(context.Background(), clusterId, topicName)
	if err != nil {
		return handleCommonKafkaRestClientErrors(url, client, resp, err)
	} else if configsResp.Data == nil {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
	}
	topicData.Configs = make(map[string]string)
	for _, config := range configsResp.Data {
		topicData.Configs[config.Name] = *config.Value
	}

	// Print topic info
	if format == output.Human.String() { // human output
		// Output partitions info
		pcmd.Printf(cmd, "Topic: %s PartitionCount: %d ReplicationFactor: %d\n", topicData.TopicName, topicData.PartitionCount, topicData.ReplicationFactor)
		partitionsTableLabels := []string{"Topic", "Partition", "Leader", "Replicas", "ISR"}
		partitionsTableEntries := make([][]string, topicData.PartitionCount)
		for i, partition := range topicData.Partitions {
			partitionsTableEntries[i] = printer.ToRow(&partition, []string{"TopicName", "PartitionId", "LeaderBrokerId", "ReplicaBrokerIds", "InSyncReplicaBrokerIds"})
		}
		printer.RenderCollectionTable(partitionsTableEntries, partitionsTableLabels)
		// Output config info
		pcmd.Print(cmd, "\nConfiguration\n\n")
		configsTableLabels := []string{"Name", "Value"}
		configsTableEntries := make([][]string, len(topicData.Configs))
		i := 0
		for name, value := range topicData.Configs {
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
	} else { // machine output (json or yaml)
		err = output.StructuredOutput(format, topicData)
		if err != nil {
			return err
		}
	}
	return nil
}
