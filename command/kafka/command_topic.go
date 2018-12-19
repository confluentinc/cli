package kafka

import (
	"fmt"
	"context"
	"strings"

	"github.com/spf13/cobra"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"

	"os"
)

type topicCommand struct {
	*cobra.Command
	config *shared.Config
	client chttp.Kafka
}

// NewTopicCommand returns the Cobra clusterCommand for Kafka Cluster.
func NewTopicCommand(config *shared.Config, plugin common.Provider) *cobra.Command {
	cmd := &topicCommand{
		Command: &cobra.Command{
			Use:   "topic",
			Short: "Manage client topics.",
		},
		config: config,
	}
	cmd.init(plugin)
	return cmd.Command
}

func (c *topicCommand) init(plugin common.Provider) {
	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := c.config.CheckLogin(); err != nil {
			return err
		}
		// Lazy load plugin to avoid unnecessarily spawning child processes
		return plugin(&c.client)
	}

	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List Kafka topics.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})

	cmd := &cobra.Command{
		Use:   "create TOPIC",
		Short: "Create a Kafka topic.",
		RunE:  c.create,
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().Uint32("partitions", 12, "Number of topic partitions.")
	cmd.Flags().Uint32("replication-factor", 3, "Replication factor.")
	cmd.Flags().StringSlice("config", nil, "A comma separated list of topic configuration (key=value) overrides for the topic being created.")
	cmd.Flags().Bool("dry-run", false, "Execute request without committing change to Kafka")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	c.AddCommand(&cobra.Command{
		Use:   "describe TOPIC",
		Short: "Describe a Kafka topic.",
		RunE:  c.describe,
		Args:  cobra.ExactArgs(1),
	})
	cmd = &cobra.Command{
		Use:   "update TOPIC",
		Short: "Update a Kafka topic.",
		RunE:  c.update,
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().StringSlice("config", nil, "A comma separated list of topic configuration (key=value) overrides for the topic being created.")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	c.AddCommand(&cobra.Command{
		Use:   "delete TOPIC",
		Short: "Delete a Kafka topic.",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	})
	// TODO: add consume/produce functionality
	//c.AddCommand(&cobra.Command{
	//	Use:   "produce TOPIC",
	//	Short: "Produce messages to a Kafka topic.",
	//	RunE:  c.produce,
	//	Args:  cobra.ExactArgs(1),
	//})
	//c.AddCommand(&cobra.Command{
	//	Use:   "consume TOPIC",
	//	Short: "Consume messages from a Kafka topic.",
	//	RunE:  c.consume,
	//	Args:  cobra.ExactArgs(1),
	//})
}

func (c *topicCommand) list(cmd *cobra.Command, args []string) error {

	cluster, err := c.currentCluster()
	if err != nil {
		return err
	}

	resp, err := c.client.ListTopics(context.Background(), cluster)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	jsonPrinter.PrintObj(resp, os.Stdout)
	return nil
}

func (c *topicCommand) create(cmd *cobra.Command, args []string) error {
	cluster, err := c.currentCluster()
	if err != nil {
		return err
	}

	topic := &kafkav1.Topic{
		Spec: &kafkav1.TopicSpecification{
			Configs: make(map[string]string)},
		Validate: false}

	topic.Spec.Name = args[0]

	topic.Spec.NumPartitions, err = cmd.Flags().GetUint32("partitions")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	topic.Spec.ReplicationFactor, err = cmd.Flags().GetUint32("replication-factor")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	topic.Validate, err = cmd.Flags().GetBool("dry-run")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	topic.Spec.Configs = toMap(configs)

	err = c.client.CreateTopic(context.Background(), cluster, topic)
	return common.HandleError(shared.KafkaError(err), cmd)
}

func (c *topicCommand) describe(cmd *cobra.Command, args []string) error {

	cluster, err := c.currentCluster()
	if err != nil {
		return err
	}

	topic := &kafkav1.TopicSpecification{Name: args[0]}

	resp, err := c.client.DescribeTopic(context.Background(), cluster, &kafkav1.Topic{Spec: topic, Validate: false})

	if err != nil {
		return common.HandleError(shared.KafkaError(err), cmd)
	}

	jsonPrinter.PrintObj(resp, os.Stdout)
	return nil
}

func (c *topicCommand) update(cmd *cobra.Command, args []string) error {

	cluster, err := c.currentCluster()
	if err != nil {
		return err
	}

	topic := &kafkav1.TopicSpecification{Name: args[0], Configs: make(map[string]string)}
	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	topic.Configs = toMap(configs)

	err = c.client.UpdateTopic(context.Background(), cluster, &kafkav1.Topic{Spec: topic, Validate: false})
	return common.HandleError(shared.KafkaError(err), cmd)
}

func (c *topicCommand) delete(cmd *cobra.Command, args []string) error {

	cluster, err := c.currentCluster()
	if err != nil {
		return err
	}

	topic := &kafkav1.TopicSpecification{Name: args[0]}
	err = c.client.DeleteTopic(context.Background(), cluster, &kafkav1.Topic{Spec: topic, Validate: false})

	return common.HandleError(shared.KafkaError(err), cmd)
}

func toMap(configs []string) map[string]string {
	configMap := make(map[string]string)
	for _, config := range configs {
		pair := strings.SplitN(config, "=", 2)
		configMap[pair[0]] = pair[1]
	}
	return configMap
}

func (c *topicCommand) currentCluster() (*kafkav1.Cluster, error) {
	ctx, err := c.config.Context()
	if ctx == nil {
		return nil, fmt.Errorf("no cluster context is currently selected")
	}

	conf, err := c.config.KafkaClusterConfig()
	if err != nil {
		return nil, err
	}

	return &kafkav1.Cluster{AccountId: c.config.Auth.Account.Id, Id: ctx.Kafka, ApiEndpoint: conf.APIEndpoint}, nil
}
