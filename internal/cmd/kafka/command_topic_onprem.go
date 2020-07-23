package kafka

// confluent kafka topic <commands>
import (
	"context"
	"fmt"

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
				Desc: "List all topics at specified cluster (providing REST proxy endpoint).",
				Code: "confluent kafka topic list --url http://localhost:8082",
			},
		),
	}
	listCmd.Flags().String("url", "", "URL to REST Proxy Endpoint of Kafka Cluster")
	//	check(describeCmd.MarkFlagRequired("url")) // can set flag to being required
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	// same as topicCmd.CLICommand.Command.AddCommand(..)
	topicCmd.AddCommand(listCmd)
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

	// note for future, set up PreRunner CLI command and make client maker code in prerunner.go (see createMDSCLient)
	// Create Kafka Proxy Client
	config := kafkaproxy.NewConfiguration()
	config.BasePath = url + "/v3"
	proxyClient := kafkaproxy.NewAPIClient(config)
	// Get Cluster Id
	clusters, _, err := proxyClient.ClusterApi.ClustersGet(context.Background())
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
	outputWriter, err := output.NewListOutputWriter(cmd, []string{"ClusterId", "TopicName"}, []string{"Cluster Id", "Topic Name"}, []string{"cluster_id", "topic_name"})
	if err != nil {
		return err
	}
	for _, topicData := range topicDatas {
		outputWriter.AddElement(&topicData)
	}

	fmt.Printf("URL: %s\n", url)
	fmt.Printf("cluster id: %s\n", clusterId)
	return outputWriter.Out()
}
