package local

import (
	"context"
	"fmt"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
)

type topicOut struct {
	Name string `human:"Name" serialized:"name"`
}

func (c *kafkaCommand) newTopicCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "Run Kafka topic related commands",
		Long:  `Run Kafka commands including produce/consume and list topics`,
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newUpdateCommand())
	cmd.AddCommand(c.newProduceCommand())
	cmd.AddCommand(c.newConsumeCommand())
	return cmd
}

func initKafkaRest(c *pcmd.CLICommand, cmd *cobra.Command) (*kafkarestv3.APIClient, string, error) {
	if c.Config.LocalPorts == nil {
		return nil, "", errors.NewErrorWithSuggestions(errors.FailedToReadPortsErrorMsg, errors.FailedToReadPortsSuggestions)
	}
	url := fmt.Sprintf(urlPrefix, c.Config.LocalPorts.RestPort)

	unsafeTrace, err := c.Flags().GetBool("unsafe-trace")
	if err != nil {
		return nil, "", err
	}
	kafkaREST := pcmd.KafkaREST{
		Context: context.Background(),
		Client:  pcmd.CreateKafkaRESTClient(url, unsafeTrace),
	}
	kafkaRestClient := kafkaREST.Client
	kafka.SetServerURL(cmd, kafkaRestClient, url)

	clusterListData, _, err := kafkaRestClient.ClusterV3Api.ClustersGet(kafkaREST.Context)
	if err != nil {
		return nil, "", err
	}

	if len(clusterListData.Data) < 1 {
		return nil, "", errors.New("failed to obtain cluster information")
	}
	return kafkaRestClient, clusterListData.Data[0].ClusterId, nil
}
