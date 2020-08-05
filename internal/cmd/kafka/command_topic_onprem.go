package kafka

// confluent kafka topic <commands>
import (
	"context"
	"net/http"
	purl "net/url"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
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
	// confluent kafka topic list [flags]
	// --url string   REST Proxy URL.
	// -o, --output string    Specify the output format as "human", "json" or "yaml". (default "human")
	// -h, --help
	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(topicCmd.listTopics),
		Short: "List Kafka topics.",
		Example: examples.BuildExampleString(
			examples.Example{
				// on-prem examples are ccloud examples + "at specified cluster (providing Kafka REST Proxy endpoint)."
				Text: "List all topics at specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "confluent kafka topic list --url http://localhost:8082",
			},
		),
	}
	listCmd.Flags().String("url", "", "Base URL to REST Proxy Endpoint of Kafka Cluster.")
	check(listCmd.MarkFlagRequired("url")) // TODO: unset url as required
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	// same as topicCmd.CLICommand.Command.AddCommand(..)
	topicCmd.AddCommand(listCmd)

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

	// Register create command
	// confluent kafka topic create <topic> [flags]
	// --cluster string      Kafka cluster ID. —> —url
	// --partitions uint32   Number of topic partitions. (default 6)
	// --config strings      A comma-separated list of topics configuration ('key=value') overrides for the topic being created.
	// --dry-run             Run the command without committing changes to Kafka. (Topic creation argument)
	// --if-not-exists       Exit gracefully if topic already exists. (CLI option)
	// createCmd := &cobra.Command{
	// 	Use:   "create <topic>",
	// 	Short: "Create a Kafka topic.",
	// 	Args:  cobra.ExactArgs(1),
	// 	RunE:  pcmd.NewCLIRunE(topicCmd.createTopic),
	// 	Example: examples.BuildExampleString(
	// 		examples.Example{
	// 			Text: "Create a topic named ``my_topic`` with default options at specified cluster (providing Kafka REST Proxy endpoint).",
	// 			Code: "confluent kafka topic create my_topic --url http://localhost:8082",
	// 		}),
	// }
	// createCmd.Flags().String("url", "", "Base URL to Kafka REST Proxy Endpoint of Kafka Cluster.")
	// check(listCmd.MarkFlagRequired("url")) // TODO: unset url as required
	// createCmd.Flags().Uint32("partitions", 6, "Number of topic partitions.")
	// createCmd.Flags().StringSlice("config", nil, "A comma-separated list of topic configuration ('key=value') overrides for the topic being created.")
	// createCmd.Flags().Bool("dry-run", false, "Run the command without committing changes to Kafka.")
	// createCmd.Flags().Bool("if-not-exists", false, "Exit gracefully if topic already exists.")
	// createCmd.Flags().SortFlags = false
	// topicCmd.AddCommand(createCmd)
}

func setServerURL(client *kafkarestv3.APIClient, url string) {
	client.ChangeBasePath(strings.Trim(url, "/") + "/v3")
}

func handleCommonKafkaRestClientErrors(url string, kafkaRestClient *kafkarestv3.APIClient, resp *http.Response, err error) error {
	// fmt.Printf("[INFO] http.Response:%v, TypeOf(err): %v, ValueOf(err): %v\n\n", resp, reflect.TypeOf(err), reflect.ValueOf(err))
	switch err.(type) {
	case *purl.Error: // Handle errors with request url
		if e, ok := err.(*purl.Error); ok {
			// TODO: Currently this error exposes implementation detail
			return errors.Errorf(errors.InvalidFlagValueWithWrappedErrorErrorMsg, url, "url", e.Err)
		}
	}
	return err
}

// Called when command matches registered words
// topicCommand is *this* current topicCommand
// cobra.Command contains the cobra.Command matched by CLI input
// args contains all the args after the first string
// testing: 1) argument parsing/validation, 2) argument -> return from proxy, 3) given objects -> formatting output
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
