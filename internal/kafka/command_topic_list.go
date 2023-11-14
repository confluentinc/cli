package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type topicNameOut struct {
	Name string `human:"Name" serialized:"name"`
}

type topicOut struct {
	Name              string `human:"Name" serialized:"name"`
	ClusterId         string `human:"Cluster ID" serialized:"cluster_id"`
	Kind              string `human:"Kind" serialized:"kind"`
	IsInternal        bool   `human:"Is Internal" serialized:"is_internal"`
	ResourceName      string `human:"Resource Name" serialized:"resouce_name"`
	ReplicationFactor int32  `human:"Replication Factor" serialized:"replication_factor"`
	PartitionsCount   int32  `human:"Partitions Count" serialized:"partitions_count"`
}

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List Kafka topics.",
		Args:        cobra.NoArgs,
		RunE:        c.list,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.Flags().Bool("detailed", false, "List detailed topic information.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	topics, err := c.getTopics()
	if err != nil {
		return err
	}

	detailed, err := cmd.Flags().GetBool("detailed")
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, topic := range topics {
		if detailed {
			list.Add(&topicOut{
				Name:              topic.GetTopicName(),
				ClusterId:         topic.GetClusterId(),
				IsInternal:        topic.GetIsInternal(),
				Kind:              topic.GetKind(),
				ResourceName:      topic.Metadata.GetResourceName(),
				ReplicationFactor: topic.GetReplicationFactor(),
				PartitionsCount:   topic.GetPartitionsCount(),
			})
		} else {
			list.Add(&topicNameOut{Name: topic.GetTopicName()})
		}
	}
	return list.Print()
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
