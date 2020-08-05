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
}

func setServerURL(client *kafkarestv3.APIClient, url string) {
	client.ChangeBasePath(strings.Trim(url, "/") + "/v3")
}

func handleCommonKafkaRestClientErrors(url string, kafkaRestClient *kafkarestv3.APIClient, resp *http.Response, err error) error {
	switch err.(type) {
	case *purl.Error: // Handle errors with request url
		if e, ok := err.(*purl.Error); ok {
			// TODO: Currently this error exposes implementation detail
			return errors.Errorf(errors.InvalidFlagValueWithWrappedErrorErrorMsg, url, "url", e.Err)
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
