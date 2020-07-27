package kafka

// confluent kafka topic <commands>
import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"

	kafkaproxy "github.com/confluentinc/kafka-rest-proxy-sdk-go/kafkaproxyv3-6.0.x"
)

// Info needed to complete kafka topic ...
type topicCommand struct {
	*pcmd.CLICommand
	prerunner pcmd.PreRunner
}

// Return the command to be registered to the kafka topic slot
func NewTopicCommandOnPrem(prerunner pcmd.PreRunner) *cobra.Command {
	topicCmd := &topicCommand{
		// Create CLICommand, use prerunner to set up PersistentPreRunE for Anonymous
		CLICommand: pcmd.NewAnonymousCLICommand(
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
		RunE:  topicCmd.listTopics,
		Short: "List Kafka topics.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all topics at specified cluster (providing REST proxy endpoint).",
				Code: "confluent kafka topic list --url http://localhost:8082",
			},
		),
	}
	listCmd.Flags().String("url", "", "Base URL to REST Proxy Endpoint of Kafka Cluster.")
	check(listCmd.MarkFlagRequired("url")) // can set flag to being required
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	// same as topicCmd.CLICommand.Command.AddCommand(..)
	topicCmd.AddCommand(listCmd)

	// Register topic describe command
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
	// describeCmd.Flags().String("url", "", "URL to REST Proxy Endpoint of Kafka Cluster.")
	// describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	// describeCmd.Flags().SortFlags = false
	// topicCmd.AddCommand(describeCmd)
}

// TODO: Should I refactor this into a Prerunner?
// I currently have it outside because the url is passed in by user
func createProxyClient(url string) *kafkaproxy.APIClient {
	config := kafkaproxy.NewConfiguration()
	config.BasePath = strings.Trim(url, "/") + "/v3" // TODO: Not sure if I want this trimming
	return kafkaproxy.NewAPIClient(config)
}

// Called when command matches registered words
// topicCommand is *this* current topicCommand
// cobra.Command contains the cobra.Command matched by CLI input
// args contains all the args after the first string
func (topicCmd *topicCommand) listTopics(cmd *cobra.Command, args []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil { // require the flag
		return err
	}

	proxyClient := createProxyClient(url)
	// Get Cluster Id
	clusters, _, err := proxyClient.ClusterApi.ClustersGet(context.Background())
	// TODO: Should I take the HTTP response and parse it create a more helpful error message? Similar to cmd/cluster/command_list.go:listCmd.list()
	if err != nil {
		return err
	}
	clusterId := clusters.Data[0].ClusterId

	// Get Topics
	topicGetResp, _, err := proxyClient.TopicApi.ClustersClusterIdTopicsGet(context.Background(), clusterId)
	if err != nil {
		return err
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
