package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type topicOut struct {
	Name string `human:"Name" serialized:"name"`
}

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List Kafka topics.",
		Args:        cobra.NoArgs,
		RunE:        c.list,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	if c.V2Client != nil {
		topics, err := c.getTopics()
		if err != nil {
			return err
		}

		list := output.NewList(cmd)
		for _, topic := range topics {
			list.Add(&topicOut{Name: topic.GetTopicName()})
		}
		return list.Print()
	}

	return c.getTopicsWithConfluentKafka(cmd)
}

func (c *command) getTopics() ([]kafkarestv3.TopicData, error) {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return nil, err
	}

	if err := c.provisioningClusterCheck(kafkaREST.GetClusterId()); err != nil {
		return nil, err
	}

	topics, err := kafkaREST.CloudClient.ListKafkaTopics()
	if err != nil {
		return nil, err
	}

	return topics.Data, nil
}

func (c *command) getTopicsWithConfluentKafka(cmd *cobra.Command) error {
	cluster, err := c.Context.GetKafkaClusterForCommand(nil)
	if err != nil {
		return err
	}

	if err := addApiKeyToCluster(cmd, cluster); err != nil {
		return err
	}

	producer, err := newProducer(cluster, c.clientID, "", nil)
	if err != nil {
		return err
	}

	adminClient, err := ckafka.NewAdminClientFromProducer(producer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientErrorMsg, err)
	}
	defer adminClient.Close()

	metadata, err := adminClient.GetMetadata(nil, true, int(adminClientTimeout.Milliseconds()))
	if err != nil {
		if err.Error() == ckafka.ErrTransport.String() {
			err = fmt.Errorf("API key may not be provisioned yet")
		}
		return fmt.Errorf("failed to obtain topics from client: %w", err)
	}

	list := output.NewList(cmd)
	for _, topic := range metadata.Topics {
		list.Add(&topicOut{Name: topic.Topic})
	}
	return list.Print()
}
