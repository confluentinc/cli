package kafka

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/Shopify/sarama"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/go-printer"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	serdes "github.com/confluentinc/cli/internal/pkg/serdes"

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
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(h.produce),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().String("delimiter", ":", "The key/value delimiter.")
	cmd.Flags().String("value-format", "string", "Format of message value as string, avro, protobuf, or jsonschema.")
	cmd.Flags().String("schema", "", "The path to the schema file.")
	cmd.Flags().Bool("parse-key", false, "Parse key from the message.")
	cmd.Flags().String("sr-endpoint", "", "Endpoint for Schema Registry cluster.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false
	h.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "consume <topic>",
		Short: "Consume messages from a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(h.consume),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Consume items from the ``my_topic`` topic and press ``Ctrl+C`` to exit.",
				Code: "ccloud kafka topic consume -b my_topic",
			},
		),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().String("group", fmt.Sprintf("confluent_cli_consumer_%s", uuid.New()), "Consumer group ID.")
	cmd.Flags().BoolP("from-beginning", "b", false, "Consume from beginning of the topic.")
	cmd.Flags().String("value-format", "string", "Format of message value as string, avro, protobuf, or jsonschema.")
	cmd.Flags().Bool("print-key", false, "Print key of the message.")
	cmd.Flags().String("delimiter", "\t", "The key/value delimiter.")
	cmd.Flags().String("sr-endpoint", "", "Endpoint for Schema Registry cluster.")
	cmd.Flags().SortFlags = false
	h.AddCommand(cmd)
}

func (a *authenticatedTopicCommand) init() {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka topics.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(a.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all topics.",
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
		Short: "Create a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(a.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a topic named ``my_topic`` with default options.",
				Code: "ccloud kafka topic create my_topic",
			},
		),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().Uint32("partitions", 6, "Number of topic partitions.")
	cmd.Flags().StringSlice("config", nil, "A comma-separated list of topics. Configuration ('key=value') overrides for the topic being created.")
	cmd.Flags().String("link", "", "The name of the cluster link the topic is associated with, if mirrored.")
	cmd.Flags().String("mirror-topic", "", "The name of the topic over the cluster link to mirror.")
	cmd.Flags().Bool("dry-run", false, "Run the command without committing changes to Kafka.")
	cmd.Flags().Bool("if-not-exists", false, "Exit gracefully if topic already exists.")
	cmd.Flags().SortFlags = false
	a.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "describe <topic>",
		Short: "Describe a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(a.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the ``my_topic`` topic.",
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
		Short: "Update a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(a.update),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Modify the ``my_topic`` topic to have a retention period of 3 days (259200000 milliseconds).",
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
		Use:   "mirror <action> <topic>",
		Short: "Perform a mirroring action on a Kafka topic.",
		Args:  cobra.ExactArgs(2),
		RunE:  pcmd.NewCLIRunE(a.mirror),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Stop the mirroring of topic ``my_topic``.",
				Code: "ccloud kafka topic mirror stop my_topic",
			},
		),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().Bool("dry-run", false, "Validate the request without applying changes to Kafka.")
	cmd.Flags().SortFlags = false
	a.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "delete <topic>",
		Short: "Delete a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(a.delete),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete the topics ``my_topic`` and ``my_topic_avro``. Use this command carefully as data loss can occur.",
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
		return err
	}
	resp, err := a.Client.Kafka.ListTopics(context.Background(), cluster)
	if err != nil {
		err = errors.CatchClusterNotReadyError(err, cluster.Id)
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, []string{"Name"}, []string{"Name"}, []string{"name"})
	if err != nil {
		return err
	}
	for _, topic := range resp {
		outputWriter.AddElement(topic)
	}
	return outputWriter.Out()
}

func (a *authenticatedTopicCommand) create(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, a.Context)
	if err != nil {
		return err
	}

	topic := &schedv1.Topic{
		Spec: &schedv1.TopicSpecification{
			Configs: make(map[string]string)},
		Validate: false,
	}

	topic.Spec.Name = args[0]

	topic.Spec.NumPartitions, err = cmd.Flags().GetUint32("partitions")
	if err != nil {
		return err
	}

	const defaultReplicationFactor = 3
	topic.Spec.ReplicationFactor = defaultReplicationFactor

	topic.Validate, err = cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}

	if topic.Spec.Configs, err = toMap(configs); err != nil {
		return err
	}

	linkName, err := cmd.Flags().GetString("link")
	if err != nil {
		return err
	}

	mirrorTopic, err := cmd.Flags().GetString("mirror-topic")
	if err != nil {
		return err
	}

	if len(linkName) > 0 || len(mirrorTopic) > 0 {
		topic.Spec.Mirror = &schedv1.TopicMirrorSpecification{LinkName: linkName, MirrorTopic: mirrorTopic}
	}

	if err := a.Client.Kafka.CreateTopic(context.Background(), cluster, topic); err != nil {
		ifNotExistsFlag, flagErr := cmd.Flags().GetBool("if-not-exists")
		if flagErr != nil {
			return flagErr
		}
		err = errors.CatchTopicExistsError(err, cluster.Id, topic.Spec.Name, ifNotExistsFlag)
		err = errors.CatchClusterNotReadyError(err, cluster.Id)
		return err
	}
	pcmd.ErrPrintf(cmd, errors.CreatedTopicMsg, topic.Spec.Name)
	return nil
}

func (a *authenticatedTopicCommand) describe(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, a.Context)
	if err != nil {
		return err
	}

	topic := &schedv1.TopicSpecification{Name: args[0]}
	resp, err := a.Client.Kafka.DescribeTopic(context.Background(), cluster, &schedv1.Topic{Spec: topic, Validate: false})
	if err != nil {
		return err
	}
	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
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
		return err
	}

	topic := &schedv1.TopicSpecification{Name: args[0], Configs: make(map[string]string)}

	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}

	configMap, err := toMap(configs)
	if err != nil {
		return err
	}
	topic.Configs = copyMap(configMap)

	validate, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}
	err = a.Client.Kafka.UpdateTopic(context.Background(), cluster, &schedv1.Topic{Spec: topic, Validate: validate})
	if err != nil {
		err = errors.CatchClusterNotReadyError(err, cluster.Id)
		return err
	}
	pcmd.Printf(cmd, errors.UpdateTopicConfigMsg, args[0])
	var entries [][]string
	titleRow := []string{"Name", "Value"}
	for name, value := range configMap {
		record := &struct {
			Name  string
			Value string
		}{
			name,
			value,
		}
		entries = append(entries, printer.ToRow(record, titleRow))
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i][0] < entries[j][0]
	})
	printer.RenderCollectionTable(entries, titleRow)
	return nil
}

func (a *authenticatedTopicCommand) mirror(cmd *cobra.Command, args []string) error {
	const stopAction = "stop"

	action := args[0]
	topic := args[1]

	cluster, err := pcmd.KafkaCluster(cmd, a.Context)
	if err != nil {
		return err
	}

	validate, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	op := &schedv1.AlterMirrorOp{}
	switch action {
	case stopAction:
		op.Type = &schedv1.AlterMirrorOp_StopTopicMirror_{
			StopTopicMirror: &schedv1.AlterMirrorOp_StopTopicMirror{
				Topic: &schedv1.Topic{Spec: &schedv1.TopicSpecification{Name: topic}, Validate: validate},
			},
		}
	default:
		return fmt.Errorf(errors.InvalidMirrorActionMsg, action)
	}

	result, err := a.Client.Kafka.AlterMirror(context.Background(), cluster, op)
	if err != nil {
		return err
	}

	switch action {
	case stopAction:
		result.GetStopTopicMirror()
		pcmd.Printf(cmd, errors.StoppedTopicMirrorMsg, topic)
	default:
		panic("unreachable")
	}

	return nil
}

func (a *authenticatedTopicCommand) delete(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, a.Context)
	if err != nil {
		return err
	}

	topic := &schedv1.TopicSpecification{Name: args[0]}
	err = a.Client.Kafka.DeleteTopic(context.Background(), cluster, &schedv1.Topic{Spec: topic, Validate: false})
	if err != nil {
		err = errors.CatchClusterNotReadyError(err, cluster.Id)
		return err
	}
	pcmd.ErrPrintf(cmd, errors.DeletedTopicMsg, args[0])
	return nil
}

func (h *hasAPIKeyTopicCommand) registerSchema(cmd *cobra.Command, subject string, valueFormat string, schemaPath string) ([]byte, error) {
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}
	var refs []srsdk.SchemaReference

	srClient, ctx, err := sr.GetApiClient(cmd, nil, h.Config, h.Version)
	if err != nil {
		if err.Error() == "ccloud" {
			return nil, &errors.SRNotAuthenticatedError{CLIName: err.Error()}
		} else {
			return nil, err
		}
	}

	response, _, err := srClient.DefaultApi.Register(ctx, subject, srsdk.RegisterSchemaRequest{Schema: string(schema), SchemaType: valueFormat, References: refs})
	if err != nil {
		return nil, err
	}

	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return nil, err
	}
	if outputFormat == output.Human.String() {
		pcmd.Println(cmd, errors.RegisteredSchemaMsg, response.Id)
	} else {
		err = output.StructuredOutput(outputFormat, &struct {
			Id int32 `json:"id" yaml:"id"`
		}{response.Id})
		if err != nil {
			return nil, err
		}
	}

	metaInfo := []byte{0x0}
	schemaIdBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(schemaIdBuffer, uint32(response.Id))
	metaInfo = append(metaInfo, schemaIdBuffer...)
	return metaInfo, nil
}

func (h *hasAPIKeyTopicCommand) produce(cmd *cobra.Command, args []string) error {
	topic := args[0]
	cluster, err := h.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}

	delim, err := cmd.Flags().GetString("delimiter")
	if err != nil {
		return err
	}

	valueFormat, err := cmd.Flags().GetString("value-format")
	if err != nil {
		return err
	}

	schemaPath, err := cmd.Flags().GetString("schema")
	if err != nil {
		return err
	}

	parseKey, err := cmd.Flags().GetBool("parse-key")
	if err != nil {
		return err
	}

	subject := topic + "-value"
	serializationProvider, err := serdes.GetSerializationProvider(valueFormat)
	if err != nil {
		return err
	}
	err = serializationProvider.LoadSchema(schemaPath)
	if err != nil {
		return err
	}

	// Meta info contains magic byte and schema ID (4 bytes).
	// For plain string encoding, meta info is empty.
	metaInfo := []byte{}

	// Registering schema when specified, and fill metaInfo array.
	if valueFormat != "string" && len(schemaPath) > 0 {
		info, err := h.registerSchema(cmd, subject, serializationProvider.GetSchemaName(), schemaPath)
		if err != nil {
			return err
		}
		metaInfo = info
	}

	pcmd.ErrPrintln(cmd, errors.StartingProducerMsg)

	InitSarama(h.logger)
	producer, err := NewSaramaProducer(cluster, h.clientID)
	if err != nil {
		err = errors.CatchClusterUnreachableError(err, cluster.ID, cluster.APIKey)
		return err
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
	go scan()

	var key sarama.Encoder
	for data := range input {
		if len(data) == 0 {
			go scan()
			continue
		}
		var valueString string
		if parseKey {
			record := strings.SplitN(data, delim, 2)
			valueString = strings.TrimSpace(record[len(record)-1])

			if len(record) == 2 {
				key = sarama.StringEncoder(strings.TrimSpace(record[0]))
			} else {
				return errors.New(errors.MissingKeyErrorMsg)
			}
		} else {
			valueString = strings.TrimSpace(data)
		}
		encodedMessage, err := serdes.Serialize(serializationProvider, valueString)
		if err != nil {
			return err
		}
		encoded := append(metaInfo, encodedMessage...)
		value := sarama.StringEncoder(string(encoded))

		msg := &sarama.ProducerMessage{Topic: topic, Key: key, Value: value}
		_, offset, err := producer.SendMessage(msg)
		if err != nil {
			isTopicNotExistError, err := errors.CatchTopicNotExistError(err, topic, cluster.ID)
			if isTopicNotExistError {
				scanErr = err
				close(input)
				break
			}
			pcmd.ErrPrintf(cmd, errors.FailedToProduceErrorMsg, offset, err)
		}

		// Reset key prior to reuse
		key = nil
		go scan()
	}
	if scanErr != nil {
		return scanErr
	}
	return producer.Close()
}

func (h *hasAPIKeyTopicCommand) consume(cmd *cobra.Command, args []string) error {
	topic := args[0]
	beginning, err := cmd.Flags().GetBool("from-beginning")
	if err != nil {
		return err
	}

	valueFormat, err := cmd.Flags().GetString("value-format")
	if err != nil {
		return err
	}

	cluster, err := h.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}
	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	printKey, err := cmd.Flags().GetBool("print-key")
	if err != nil {
		return err
	}

	delimiter, err := cmd.Flags().GetString("delimiter")
	if err != nil {
		return err
	}

	var srClient *srsdk.APIClient
	var ctx context.Context
	if valueFormat != "string" {

		// Only initialize client and context when schema is specified.
		srClient, ctx, err = sr.GetApiClient(cmd, nil, h.Config, h.Version)
		if err != nil {
			if err.Error() == "ccloud" {
				return &errors.SRNotAuthenticatedError{CLIName: err.Error()}
			} else {
				return err
			}
		}
	} else {
		srClient, ctx = nil, nil
	}

	InitSarama(h.logger)
	consumer, err := NewSaramaConsumer(group, cluster, h.clientID, beginning)
	if err != nil {
		err = errors.CatchClusterUnreachableError(err, cluster.ID, cluster.APIKey)
		return err
	}

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		pcmd.ErrPrintln(cmd, errors.StoppingConsumer)
		consumer.Close()
	}()

	go func() {
		for err := range consumer.Errors() {
			pcmd.ErrPrintln(cmd, "ERROR", err)
		}
	}()

	pcmd.ErrPrintln(cmd, errors.StartingConsumerMsg)

	dir := filepath.Join(os.TempDir(), "ccloud-schema")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}
	}

	groupHandler := &GroupHandler{
		SrClient:   srClient,
		Ctx:        ctx,
		Format:     valueFormat,
		Out:        cmd.OutOrStdout(),
		Properties: ConsumerProperties{PrintKey: printKey, Delimiter: delimiter, SchemaPath: dir},
	}
	err = consumer.Consume(context.Background(), []string{topic}, groupHandler)
	_, err = errors.CatchTopicNotExistError(err, topic, cluster.ID)
	if err != nil {
		return err
	}
	err = os.RemoveAll(dir)
	return err
}

func toMap(configs []string) (map[string]string, error) {
	configMap := make(map[string]string)
	for _, cfg := range configs {
		pair := strings.SplitN(cfg, "=", 2)
		if len(pair) < 2 {
			return nil, fmt.Errorf(errors.ConfigurationFormErrorMsg)
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
