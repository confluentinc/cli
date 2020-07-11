package kafka

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"

	"github.com/Shopify/sarama"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/go-printer"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type hasAPIKeyTopicCommand struct {
	*pcmd.HasAPIKeyCLICommand
	prerunner pcmd.PreRunner
	logger    *log.Logger
	clientID  string
}
type authenticatedTopicCommand struct {
	*pcmd.AuthenticatedCLICommand
	logger   *log.Logger
	clientID string
}

type partitionDescribeDisplay struct {
	Topic     string   `json:"topic" yaml:"topic"`
	Partition uint32   `json:"partition" yaml:"partition"`
	Leader    uint32   `json:"leader" yaml:"leader"`
	Replicas  []uint32 `json:"replicas" yaml:"replicas"`
	ISR       []uint32 `json:"isr" yaml:"isr"`
}

type structuredDescribeDisplay struct {
	TopicName         string                     `json:"topic_name" yaml:"topic_name"`
	PartitionCount    int                        `json:"partition_count" yaml:"partition_count"`
	ReplicationFactor int                        `json:"replication_factor" yaml:"replication_factor"`
	Partitions        []partitionDescribeDisplay `json:"partitions" yaml:"partitions"`
	Config            map[string]string          `json:"config" yaml:"config"`
}

// NewTopicCommand returns the Cobra command for Kafka topic.
func NewTopicCommand(isAPIKeyLogin bool, prerunner pcmd.PreRunner, logger *log.Logger, clientID string) *cobra.Command {
	command := &cobra.Command{
		Use:   "topic",
		Short: "Manage Kafka topics.",
	}
	hasAPIKeyCmd := &hasAPIKeyTopicCommand{
		HasAPIKeyCLICommand: pcmd.NewHasAPIKeyCLICommand(command, prerunner),
		prerunner:           prerunner,
		logger:              logger,
		clientID:            clientID,
	}
	hasAPIKeyCmd.init()
	if !isAPIKeyLogin {
		authenticatedCmd := &authenticatedTopicCommand{
			AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(command, prerunner),
			logger:                  logger,
			clientID:                clientID,
		}
		authenticatedCmd.init()
	}
	return command
}

func (h *hasAPIKeyTopicCommand) init() {
	cmd := &cobra.Command{
		Use:   "produce <topic>",
		Short: "Produce messages to a Kafka topic.",
		RunE:  h.produce,
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().String("delimiter", ":", "The key/value delimiter.")
	cmd.Flags().SortFlags = false
	h.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "consume <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  h.consume,
		Short: "Consume messages from a Kafka topic.",
		Example: examples.BuildExampleString(
			examples.Example{
				Desc: "Consume items from the ``my_topic`` topic and press ``Ctrl+C`` to exit.",
				Code: "ccloud kafka topic consume -b my_topic",
			},
		),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().String("group", fmt.Sprintf("confluent_cli_consumer_%s", uuid.New()), "Consumer group ID.")
	cmd.Flags().BoolP("from-beginning", "b", false, "Consume from beginning of the topic.")
	cmd.Flags().SortFlags = false
	h.AddCommand(cmd)
}

func (a *authenticatedTopicCommand) init() {
	cmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  a.list,
		Short: "List Kafka topics.",
		Example: examples.BuildExampleString(
			examples.Example{
				Desc: "List all topics.",
				Code: "ccloud kafka topic list",
			},
		),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false
	a.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "create <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  a.create,
		Short: "Create a Kafka topic.",
		Example: examples.BuildExampleString(
			examples.Example{
				Desc: "Create a topic named ``my_topic`` with default options.",
				Code: "ccloud kafka topic create my_topic",
			},
		),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().Uint32("partitions", 6, "Number of topic partitions.")
	cmd.Flags().StringSlice("config", nil, "A comma-separated list of topics. Configuration ('key=value') overrides for the topic being created.")
	cmd.Flags().Bool("dry-run", false, "Run the command without committing changes to Kafka.")
	cmd.Flags().Bool("if-not-exists", false, "Exit gracefully if topic already exists.")
	cmd.Flags().SortFlags = false
	a.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "describe <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  a.describe,
		Short: "Describe a Kafka topic.",
		Example: examples.BuildExampleString(
			examples.Example{
				Desc: "Describe the ``my_topic`` topic.",
				Code: "ccloud kafka topic describe my_topic",
			},
		),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false
	a.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "update <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  a.update,
		Short: "Update a Kafka topic.",
		Example: examples.BuildExampleString(
			examples.Example{
				Desc: "Modify the ``my_topic`` topic to have a retention period of 3 days (259200000 milliseconds).",
				Code: `ccloud kafka topic update my_topic --config="retention.ms=259200000"`,
			},
		),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().StringSlice("config", nil, "A comma-separated list of topics. Configuration ('key=value') overrides for the topic being created.")
	cmd.Flags().Bool("dry-run", false, "Execute request without committing changes to Kafka.")
	cmd.Flags().SortFlags = false
	a.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "delete <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  a.delete,
		Short: "Delete a Kafka topic.",
		Example: examples.BuildExampleString(
			examples.Example{
				Desc: "Delete the topics ``my_topic`` and ``my_topic_avro``. Use this command carefully as data loss can occur.",
				Code: "ccloud kafka topic delete my_topic\nccloud kafka topic delete my_topic_avro",
			},
		),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().SortFlags = false
	a.AddCommand(cmd)
}

func (a *authenticatedTopicCommand) list(cmd *cobra.Command, _ []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, a.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	resp, err := a.Client.Kafka.ListTopics(context.Background(), cluster)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, []string{"Name"}, []string{"Name"}, []string{"name"})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	for _, topic := range resp {
		outputWriter.AddElement(topic)
	}
	return outputWriter.Out()
}

func (a *authenticatedTopicCommand) create(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, a.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	topic := &schedv1.Topic{
		Spec: &schedv1.TopicSpecification{
			Configs: make(map[string]string)},
		Validate: false,
	}

	topic.Spec.Name = args[0]

	topic.Spec.NumPartitions, err = cmd.Flags().GetUint32("partitions")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	const defaultReplicationFactor = 3
	topic.Spec.ReplicationFactor = defaultReplicationFactor

	topic.Validate, err = cmd.Flags().GetBool("dry-run")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if topic.Spec.Configs, err = toMap(configs); err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if err := a.Client.Kafka.CreateTopic(context.Background(), cluster, topic); err != nil {
		if err.Error() == fmt.Sprintf("error creating topic %s: Topic '%s' already exists.", topic.Spec.Name, topic.Spec.Name) {
			ifNotExists, flagErr := cmd.Flags().GetBool("if-not-exists")
			if flagErr != nil {
				return errors.HandleCommon(flagErr, cmd)
			}
			if ifNotExists {
				return nil
			}
		}
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (a *authenticatedTopicCommand) describe(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, a.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	topic := &schedv1.TopicSpecification{Name: args[0]}
	resp, err := a.Client.Kafka.DescribeTopic(context.Background(), cluster, &schedv1.Topic{Spec: topic, Validate: false})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if outputOption == output.Human.String() {
		return printHumanDescribe(cmd, resp)
	} else {
		return printStructuredDescribe(resp, outputOption)
	}
}

func (a *authenticatedTopicCommand) update(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, a.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	topic := &schedv1.TopicSpecification{Name: args[0], Configs: make(map[string]string)}

	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if topic.Configs, err = toMap(configs); err != nil {
		return errors.HandleCommon(err, cmd)
	}

	validate, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = a.Client.Kafka.UpdateTopic(context.Background(), cluster, &schedv1.Topic{Spec: topic, Validate: validate})

	return errors.HandleCommon(err, cmd)
}

func (a *authenticatedTopicCommand) delete(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, a.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	topic := &schedv1.TopicSpecification{Name: args[0]}
	err = a.Client.Kafka.DeleteTopic(context.Background(), cluster, &schedv1.Topic{Spec: topic, Validate: false})

	return errors.HandleCommon(err, cmd)
}

func (h *hasAPIKeyTopicCommand) produce(cmd *cobra.Command, args []string) error {
	topic := args[0]
	cluster, err := h.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	delim, err := cmd.Flags().GetString("delimiter")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	pcmd.ErrPrintln(cmd, "Starting Kafka Producer. ^C or ^D to exit")

	InitSarama(h.logger)
	producer, err := NewSaramaProducer(cluster, h.clientID)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	// Line reader for producer input.
	scanner := bufio.NewScanner(os.Stdin)
	// CCloud Kafka messageMaxBytes:
	// https://github.com/confluentinc/cc-spec-kafka/blob/9f0af828d20e9339aeab6991f32d8355eb3f0776/plugins/kafka/kafka.go#L43.
	const maxScanTokenSize = 1024*1024*2 + 12
	scanner.Buffer(nil, maxScanTokenSize)
	input := make(chan string, 1)
	// Avoid blocking in for loop so ^C or ^D can exit immediately.
	var scanErr error
	scan := func() {
		hasNext := scanner.Scan()
		if !hasNext {
			// Actual error.
			if scanner.Err() != nil {
				scanErr = scanner.Err()
			}
			// Otherwise just EOF.
			close(input)
		} else {
			input <- scanner.Text()
		}
	}

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		close(input)
	}()
	// Prime reader
	scan()

	var key sarama.Encoder
	for data := range input {
		data = strings.TrimSpace(data)

		record := strings.SplitN(data, delim, 2)
		value := sarama.StringEncoder(record[len(record)-1])
		if len(record) == 2 {
			key = sarama.StringEncoder(record[0])
		}
		msg := &sarama.ProducerMessage{Topic: topic, Key: key, Value: value}
		_, offset, err := producer.SendMessage(msg)
		if err != nil {
			pcmd.ErrPrintf(cmd, "Failed to produce offset %d: %s\n", offset, err)
		}

		// Reset key prior to reuse
		key = nil
		go scan()
	}
	if scanErr != nil {
		return errors.HandleCommon(scanErr, cmd)
	}
	return errors.HandleCommon(producer.Close(), cmd)
}

func (h *hasAPIKeyTopicCommand) consume(cmd *cobra.Command, args []string) error {
	topic := args[0]
	beginning, err := cmd.Flags().GetBool("from-beginning")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	cluster, err := h.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	InitSarama(h.logger)
	consumer, err := NewSaramaConsumer(group, cluster, h.clientID, beginning)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		pcmd.ErrPrintln(cmd, "Stopping Consumer.")
		consumer.Close()
	}()

	go func() {
		for err := range consumer.Errors() {
			pcmd.ErrPrintln(cmd, "ERROR", err)
		}
	}()

	pcmd.ErrPrintln(cmd, "Starting Kafka Consumer. ^C to exit")

	err = consumer.Consume(context.Background(), []string{topic}, &GroupHandler{Out: cmd.OutOrStdout()})

	return errors.HandleCommon(err, cmd)
}

func toMap(configs []string) (map[string]string, error) {
	configMap := make(map[string]string)
	for _, cfg := range configs {
		pair := strings.SplitN(cfg, "=", 2)
		if len(pair) < 2 {
			return nil, fmt.Errorf("The configuration must be in the form of key=value")
		}
		configMap[pair[0]] = pair[1]
	}
	return configMap, nil
}

func printHumanDescribe(cmd *cobra.Command, resp *schedv1.TopicDescription) error {
	pcmd.Printf(cmd, "Topic: %s PartitionCount: %d ReplicationFactor: %d\n",
		resp.Name, len(resp.Partitions), len(resp.Partitions[0].Replicas))

	var partitions [][]string
	titleRow := []string{"Topic", "Partition", "Leader", "Replicas", "ISR"}
	for _, partition := range resp.Partitions {
		partitions = append(partitions, printer.ToRow(getPartitionDisplay(partition, resp.Name), titleRow))
	}

	printer.RenderCollectionTable(partitions, titleRow)

	var entries [][]string
	titleRow = []string{"Name", "Value"}
	for _, entry := range resp.Config {
		record := &struct {
			Name  string
			Value string
		}{
			entry.Name,
			entry.Value,
		}
		entries = append(entries, printer.ToRow(record, titleRow))
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i][0] < entries[j][0]
	})
	pcmd.Println(cmd, "\nConfiguration\n ")
	printer.RenderCollectionTable(entries, titleRow)
	return nil
}

func printStructuredDescribe(resp *schedv1.TopicDescription, format string) error {
	structuredDisplay := &structuredDescribeDisplay{Config: make(map[string]string)}
	structuredDisplay.TopicName = resp.Name
	structuredDisplay.PartitionCount = len(resp.Partitions)
	structuredDisplay.ReplicationFactor = len(resp.Partitions[0].Replicas)

	var partitionList []partitionDescribeDisplay
	for _, partition := range resp.Partitions {
		partitionList = append(partitionList, *getPartitionDisplay(partition, resp.Name))
	}
	structuredDisplay.Partitions = partitionList

	for _, entry := range resp.Config {
		structuredDisplay.Config[entry.Name] = entry.Value
	}
	return output.StructuredOutput(format, structuredDisplay)
}

func getPartitionDisplay(partition *schedv1.TopicPartitionInfo, topicName string) *partitionDescribeDisplay {
	var replicas []uint32
	for _, replica := range partition.Replicas {
		replicas = append(replicas, replica.Id)
	}

	var isr []uint32
	for _, replica := range partition.Isr {
		isr = append(isr, replica.Id)
	}

	return &partitionDescribeDisplay{
		Topic:     topicName,
		Partition: partition.Partition,
		Leader:    partition.Leader.Id,
		Replicas:  replicas,
		ISR:       isr,
	}
}
