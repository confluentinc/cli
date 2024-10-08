package local

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v4/internal/kafka"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

func (c *command) newKafkaTopicCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topic",
		Short: "Run Kafka topic related commands.",
		Long:  `Run Kafka commands including produce/consume and list topics.`,
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(c.newKafkaTopicConfigurationCommand())
	cmd.AddCommand(c.newKafkaTopicConsumeCommand())
	cmd.AddCommand(c.newKafkaTopicCreateCommand())
	cmd.AddCommand(c.newKafkaTopicDeleteCommand())
	cmd.AddCommand(c.newKafkaTopicDescribeCommand())
	cmd.AddCommand(c.newKafkaTopicListCommand())
	cmd.AddCommand(c.newKafkaTopicProduceCommand())
	cmd.AddCommand(c.newKafkaTopicUpdateCommand())

	return cmd
}

func initKafkaRest(c *pcmd.CLICommand, cmd *cobra.Command) (*kafkarestv3.APIClient, string, error) {
	if c.Config.LocalPorts == nil {
		return nil, "", errors.NewErrorWithSuggestions(errors.FailedToReadPortsErrorMsg, errors.FailedToReadPortsSuggestions)
	}
	url := fmt.Sprintf(localhostPrefix, c.Config.LocalPorts.KafkaRestPort)

	unsafeTrace, err := c.Flags().GetBool("unsafe-trace")
	if err != nil {
		return nil, "", err
	}

	kafkaRest := pcmd.KafkaREST{
		Context: context.Background(),
		Client:  pcmd.CreateKafkaRESTClient(url, unsafeTrace),
	}
	kafkaRestClient := kafkaRest.Client
	kafka.SetServerURL(cmd, kafkaRestClient, url)

	clusterListData, _, err := kafkaRestClient.ClusterV3Api.ClustersGet(kafkaRest.Context)
	if err != nil {
		return nil, "", err
	}

	if len(clusterListData.Data) < 1 {
		return nil, "", fmt.Errorf("failed to obtain local cluster information")
	}

	return kafkaRestClient, clusterListData.Data[0].ClusterId, nil
}

func (c *command) getPlaintextBootstrapServers() string {
	portStrings := make([]string, 0, len(c.Config.LocalPorts.PlaintextPorts))
	for _, port := range c.Config.LocalPorts.PlaintextPorts {
		portStrings = append(portStrings, ":"+port)
	}
	return strings.Join(portStrings, ",")
}
