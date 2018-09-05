package kafka

import (
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

type topicCommand struct {
	*cobra.Command
	config *shared.Config
	kafka  Kafka
}

// NewTopicCommand returns the Cobra clusterCommand for Kafka Cluster.
func NewTopicCommand(config *shared.Config, kafka Kafka) *cobra.Command {
	cmd := &topicCommand{
		Command: &cobra.Command{
			Use:   "topic",
			Short: "Manage kafka topics.",
		},
		config: config,
		kafka:  kafka,
	}
	cmd.init()
	return cmd.Command
}

func (c *topicCommand) init() {
	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List Kafka topics.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})

	createCmd := &cobra.Command{
		Use:   "create TOPIC",
		Short: "Create a Kafka topic.",
		RunE:  c.create,
		Args:  cobra.ExactArgs(1),
	}
	createCmd.Flags().Int32("partitions", 12, "Number of topic partitions.")
	createCmd.Flags().Int16("replication-factor", 3, "Replication factor.")
	createCmd.Flags().StringSlice("config", nil, "A comma separated list of topic configuration (key=value) overrides for the topic being created.")
	c.AddCommand(createCmd)

	c.AddCommand(&cobra.Command{
		Use:   "describe TOPIC",
		Short: "Describe a Kafka topic.",
		RunE:  c.describe,
		Args:  cobra.ExactArgs(1),
	})
	c.AddCommand(&cobra.Command{
		Use:   "update TOPIC",
		Short: "Update a Kafka topic.",
		RunE:  c.update,
		Args:  cobra.ExactArgs(1),
	})
	c.AddCommand(&cobra.Command{
		Use:   "delete TOPIC",
		Short: "Delete a Kafka topic.",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	})
	c.AddCommand(&cobra.Command{
		Use:   "produce TOPIC",
		Short: "Produce messages to a Kafka topic.",
		RunE:  c.produce,
		Args:  cobra.ExactArgs(1),
	})
	c.AddCommand(&cobra.Command{
		Use:   "consume TOPIC",
		Short: "Consume messages from a Kafka topic.",
		RunE:  c.consume,
		Args:  cobra.ExactArgs(1),
	})
}

func (c *topicCommand) list(cmd *cobra.Command, args []string) error {
	client, err := NewSaramaKafkaForConfig(c.config)
	if err != nil {
		return common.HandleError(shared.KafkaError(err), cmd)
	}
	topics, err := client.Topics()
	if err != nil {
		return common.HandleError(shared.KafkaError(err), cmd)
	}
	for _, topic := range topics {
		fmt.Println(topic)
	}
	return nil
}

func (c *topicCommand) create(cmd *cobra.Command, args []string) error {
	partitions, err := cmd.Flags().GetInt32("partitions")
	if err != nil {
		return common.HandleError(err, cmd)
	}
	replicationFactor, err := cmd.Flags().GetInt16("replication-factor")
	if err != nil {
		return common.HandleError(err, cmd)
	}
	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return common.HandleError(err, cmd)
	}
	client, err := NewSaramaAdminForConfig(c.config)
	if err != nil {
		return common.HandleError(shared.KafkaError(err), cmd)
	}
	entries := map[string]*string{}
	for _, config := range configs {
		pair := strings.SplitN(config, "=", 2)
		entries[pair[0]] = &pair[1]
	}
	config := &sarama.TopicDetail{
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
		ConfigEntries:     entries,
	}
	err = client.CreateTopic(args[0], config, false)
	return common.HandleError(shared.KafkaError(err), cmd)
}

func (c *topicCommand) describe(cmd *cobra.Command, args []string) error {
	return shared.ErrNotImplemented
}

func (c *topicCommand) update(cmd *cobra.Command, args []string) error {
	return shared.ErrNotImplemented
}

func (c *topicCommand) delete(cmd *cobra.Command, args []string) error {
	client, err := NewSaramaAdminForConfig(c.config)
	if err != nil {
		return common.HandleError(shared.KafkaError(err), cmd)
	}
	err = client.DeleteTopic(args[0])
	return common.HandleError(shared.KafkaError(err), cmd)
}

func (c *topicCommand) produce(cmd *cobra.Command, args []string) error {
	return shared.ErrNotImplemented
}

func (c *topicCommand) consume(cmd *cobra.Command, args []string) error {
	return shared.ErrNotImplemented
}
