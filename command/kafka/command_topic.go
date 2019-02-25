package kafka

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/codyaray/go-printer"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

type topicCommand struct {
	*cobra.Command
	config *shared.Config
	client chttp.Kafka
}

// NewTopicCommand returns the Cobra clusterCommand for Kafka Cluster.
func NewTopicCommand(config *shared.Config, plugin common.GRPCPlugin) *cobra.Command {
	cmd := &topicCommand{
		Command: &cobra.Command{
			Use:   "topic",
			Short: "Manage Kafka topics.",
		},
		config: config,
	}
	cmd.init(plugin)
	return cmd.Command
}

func (c *topicCommand) init(plugin common.GRPCPlugin) {
	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := c.config.CheckLogin(); err != nil {
			return common.HandleError(err, cmd)
		}
		// Lazy load plugin to avoid unnecessarily spawning child processes
		return plugin.Load(&c.client)
	}

	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List Kafka topics.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})

	cmd := &cobra.Command{
		Use:   "create TOPIC",
		Short: "Load a Kafka topic.",
		RunE:  c.create,
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().Uint32("partitions", 6, "Number of topic partitions.")
	cmd.Flags().Uint32("replication-factor", 3, "Replication factor.")
	cmd.Flags().StringSlice("config", nil, "A comma separated list of topic configuration (key=value) overrides for the topic being created.")
	cmd.Flags().Bool("dry-run", false, "Execute request without committing changes to Kafka")
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
	cmd.Flags().Bool("dry-run", false, "Execute request without committing changes to Kafka")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	c.AddCommand(&cobra.Command{
		Use:   "delete TOPIC",
		Short: "Delete a Kafka topic.",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	})

	cmd = &cobra.Command{
		Use:   "produce TOPIC",
		Short: "Produce messages to a Kafka topic.",
		RunE:  c.produce,
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().String("delimiter", ":", "Key/Value delimiter")
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "consume TOPIC",
		Short: "Consume messages from a Kafka topic.",
		RunE:  c.consume,
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().String("group", fmt.Sprintf("confluent_cli_consumer%s", uuid.New()), "Consumer group id")
	c.AddCommand(cmd)

}

func (c *topicCommand) list(cmd *cobra.Command, args []string) error {
	cluster, err := common.Cluster(c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	resp, err := c.client.ListTopics(context.Background(), cluster)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	var topics [][]string
	for _, topic := range resp {
		topics = append(topics, printer.ToRow(topic, []string{"Name"}))
	}

	printer.RenderCollectionTable(topics, []string{"Name"})

	return nil
}

func (c *topicCommand) create(cmd *cobra.Command, args []string) error {
	cluster, err := common.Cluster(c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	topic := &kafkav1.Topic{
		Spec: &kafkav1.TopicSpecification{
			Configs: make(map[string]string)},
		Validate: false,
	}

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

	if topic.Spec.Configs, err = toMap(configs); err != nil {
		return common.HandleError(err, cmd)
	}

	err = c.client.CreateTopic(context.Background(), cluster, topic)

	return common.HandleError(err, cmd)
}

func (c *topicCommand) describe(cmd *cobra.Command, args []string) error {
	cluster, err := common.Cluster(c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	topic := &kafkav1.TopicSpecification{Name: args[0]}

	resp, err := c.client.DescribeTopic(context.Background(), cluster, &kafkav1.Topic{Spec: topic, Validate: false})
	if err != nil {
		return common.HandleError(err, cmd)
	}

	fmt.Printf("Topic: %s PartitionCount: %d ReplicationFactor: %d\n",
		resp.Name, len(resp.Partitions), len(resp.Partitions[0].Replicas))

	var partitions [][]string
	for _, partition := range resp.Partitions {
		var replicas []uint32
		for _, replica := range partition.Replicas {
			replicas = append(replicas, replica.Id)
		}

		var isr []uint32
		for _, replica := range partition.Isr {
			isr = append(isr, replica.Id)
		}

		record := &struct {
			Topic     string
			Partition uint32
			Leader    uint32
			Replicas  []uint32
			ISR       []uint32
		}{
			resp.Name,
			partition.Partition,
			partition.Leader.Id,
			replicas,
			isr,
		}
		partitions = append(partitions, printer.ToRow(record, []string{"Topic", "Partition", "Leader", "Replicas", "ISR"}))
	}

	printer.RenderCollectionTable(partitions, []string{"Topic", "Partition", "Leader", "Replicas", "ISR"})

	return nil
}

func (c *topicCommand) update(cmd *cobra.Command, args []string) error {
	cluster, err := common.Cluster(c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	topic := &kafkav1.TopicSpecification{Name: args[0], Configs: make(map[string]string)}

	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	if topic.Configs, err = toMap(configs); err != nil {
		return common.HandleError(err, cmd)
	}

	validate, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	err = c.client.UpdateTopic(context.Background(), cluster, &kafkav1.Topic{Spec: topic, Validate: validate})

	return common.HandleError(err, cmd)
}

func (c *topicCommand) delete(cmd *cobra.Command, args []string) error {
	cluster, err := common.Cluster(c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	topic := &kafkav1.TopicSpecification{Name: args[0]}
	err = c.client.DeleteTopic(context.Background(), cluster, &kafkav1.Topic{Spec: topic, Validate: false})

	return common.HandleError(err, cmd)
}

func (c *topicCommand) produce(cmd *cobra.Command, args []string) error {
	topic := args[0]

	delim, err := cmd.Flags().GetString("delimiter")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	fmt.Println("Starting Kafka Producer. ^C to exit")

	producer, err := NewSaramaProducer(c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	// Line reader for producer input
	scanner := bufio.NewScanner(os.Stdin)
	input := make(chan string, 1)

	// Avoid blocking in for loop so ^C can exit immediately.
	scan := func() {
		scanner.Scan()
		input <- scanner.Text()
	}
	// Prime reader
	scan()

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		close(input)
	}()

	key := ""
	for data := range input {
		data = strings.TrimSpace(data)
		if data == "" {
			continue
		}

		record := strings.SplitN(data, delim, 2)

		value := record[len(record)-1]
		if len(record) == 2 {
			key = record[0]
		}

		msg := &sarama.ProducerMessage{Topic: topic, Key: sarama.StringEncoder(key), Value: sarama.StringEncoder(value)}

		_, offset, err := producer.SendMessage(msg)
		if err != nil {
			fmt.Printf("Failed to produce offset %d: %s\n", offset, err)
		}

		// Reset key prior to reuse
		key = ""
		go scan()
	}

	return common.HandleError(producer.Close(), cmd)
}

func (c *topicCommand) consume(cmd *cobra.Command, args []string) error {
	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	consumer, err := NewSaramaConsumer(group, c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		fmt.Println("Stopping Consumer.")
		consumer.Close()
	}()

	go func() {
		for err := range consumer.Errors() {
			fmt.Println("ERROR", err)
		}
	}()

	fmt.Println("Starting Kafka Consumer. ^C to exit")

	err = consumer.Consume(context.Background(), []string{args[0]}, &GroupHandler{})

	return common.HandleError(err, cmd)
}

func toMap(configs []string) (map[string]string, error) {
	configMap := make(map[string]string)
	for _, config := range configs {
		pair := strings.SplitN(config, "=", 2)
		if len(pair) < 2 {
			return nil, fmt.Errorf("configuration must be in the form of key=value")
		}
		configMap[pair[0]] = pair[1]
	}
	return configMap, nil
}
