package kafka

// confluent kafka topic <commands>
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	purl "net/url"
	"reflect"
	"strings"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafka"
	"github.com/confluentinc/cli/internal/pkg/output"
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
	check(listCmd.MarkFlagRequired("url"))                                  // TODO: unset url as required
	createCmd.Flags().Int32("partitions", 6, "Number of topic partitions.") // TODO: change back to uint32
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

	// Register describe command
	// describeCmd := &cobra.Command{
	// 	Use:   "describe <topic>",
	// 	Args:  cobra.ExactArgs(1),
	// 	RunE:  topicCmd.describeTopic,
	// 	Short: "Describe a Kafka topic.",
	// 	Example: examples.BuildExampleString(
	// 		examples.Example{
	// 			Desc: "Describe the ``my_topic`` topic at specified cluser.",
	// 			Code: "ccloud kafka topic describe my_topic --url http://localht:8082",
	// 		},
	// 	),
	// }
	// describeCmd.Flags().String("url", "", "URL to Kafka REST Proxy Endpoint of Kafka Cluster.")
	// describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	// describeCmd.Flags().SortFlags = false
	// topicCmd.AddCommand(describeCmd)
}

func setServerURL(client *kafkarestv3.APIClient, url string) {
	client.ChangeBasePath(strings.Trim(url, "/") + "/v3")
}

type kafkaRestV3Error struct {
	Code    int    `json:"error_code"`
	Message string `json:"message"`
}

func handleCommonKafkaRestClientErrors(url string, kafkaRestClient *kafkarestv3.APIClient, resp *http.Response, err error) error {
	fmt.Printf("[DEBUG] http.Response:%v, TypeOf(err): %v, ValueOf(err): %v\n\n", resp, reflect.TypeOf(err), reflect.ValueOf(err))
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
	// cobra requires 1 argument, no need to check for error
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
		return handleCommonKafkaRestClientErrors(url, kafkaRestClient, resp, err)
	} else if clusters.Data == nil || len(clusters.Data) == 0 { // TODO: should I check for this? If so update list topics to do the same, what error should I return?
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions) // TODO: refactor into error_messages.go
	}
	clusterId := clusters.Data[0].ClusterId

	// Parse remaining arguments
	numPartitions, err := cmd.Flags().GetInt32("partitions") // TODO: change back to uint32
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
	_, err = cmd.Flags().GetBool("if-not-exists") //TODO: noErrIfExists
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
		if resp.StatusCode == http.StatusBadRequest {
			return errors.NewErrorWithSuggestions("error during topic creation",
				"The error can be one of: topic name already exists, invalid replication factor or invalid configuration keys or values")
			// TODO: Currently Kafka REST Proxy returns 400 for all errors during topic creation (including dup topic, invalid configs...)
			// 		 temporarily returning all as one error. Implement noErrIfExists when able to distinguish errors.
		}
		return handleCommonKafkaRestClientErrors(url, kafkaRestClient, resp, err)
	}
	// no error if topic is created successfully.
	pcmd.ErrPrintf(cmd, errors.CreatedTopicMsg, topicName) // TODO: why print to StdErr
	// fmt.Printf("[DEBUG] topicData: %v\n", topicData)
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

// func (topicCmd *topicCommand) describeTopic(cmd *cobra.Command, args []string) error {
// 	url, err := cmd.Flags().GetString("url")
// 	if err != nil {
// 		return err
// 	}
// 	if len(args) < 1 {
// 		return errors.New("missing topic name")
// 	}
// 	topicName := args[0]

// 	proxyClient := createProxyClient(url)
// 	clusterGetResp, _, err := proxyClient.ClusterApi.ClustersGet(context.Background())
// 	if err != nil {
// 		return err
// 	}
// 	clusterId := clusterGetResp.Data[0].ClusterId

// 	partitionsGetResp, _, err := proxyClient.PartitionApi.ClustersClusterIdTopicsTopicNamePartitionsGet(context.Background(), clusterId, topicName)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("TODO: Topic: %s, PartitionCount: %s ReplicationFactor: %s")

// 	partitionsOutputWriter, err := output.NewListOutputWriter(cmd,
// 		[]string{"TopicName", "PartitionId"},
// 		[]string{"Topic", "Partition"},
// 		[]string{"topic", "partition"})
// 	if err != nil {
// 		return err
// 	}

// 	for _, partition := range partitionsGetResp.Data {
// 		partitionsOutputWriter.AddElement(&partition)
// 	}
// 	partitionsOutputWriter.Out()
// 	fmt.Println("TODO: Leader, Replicas, and ISR columns\n")

// 	fmt.Println("Configuration\n")
// 	configsGetResp, _, err := proxyClient.ConfigsApi.ClustersClusterIdTopicsTopicNameConfigsGet(context.Background(), clusterId, topicName)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Printf("TODO: Configurations %v\n", configsGetResp)

// 	return nil
// }

// func (topicCmd *topicCommand) createTopic(cmd *cobra.Command, args []string) error {
// 	url, err := cmd.Flags().GetString("url")
// 	if err != nil {
// 		return err
// 	}

// 	setServerURL(topicCmd.ProxyClient, url)
// 	proxyClient := topicCmd.ProxyClient

// 	clusters, _, err := proxyClient.ClusterApi.ClustersGet(context.Background())
// 	if err != nil {
// 		return err
// 	}
// 	clusterId := clusters.Data[0].ClusterId

// 	// Create topic
// 	topicName := args[0]
// 	topicCreateRequestData := &kafkaproxy.CreateTopicRequestData{
// 		TopicName: topicName,
// 	}
// 	cmd.Flags().GetStringSlice("config")

// 	proxyClient.TopicApi.ClustersClusterIdTopicsPost(context.Background(), clusterId,
// 		&kafkaproxy.ClustersClusterIdTopicsPostOpts{CreateTopicRequestData: optional.NewInterface(topicCreateRequestData)})

// 	return nil
// }
