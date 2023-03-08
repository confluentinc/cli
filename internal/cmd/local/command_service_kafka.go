package local

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/local"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/types"
)

var (
	defaultBool        bool
	defaultInt         int
	defaultString      string
	defaultStringSlice []string

	commonFlagUsage = map[string]string{
		"cloud":        "Consume from Confluent Cloud.",
		"config":       "Change the Confluent Cloud configuration file.",
		"value-format": "Format output data: avro, json, or protobuf.",
	}

	kafkaConsumeFlagUsage = map[string]string{
		"bootstrap-server":      "The server(s) to connect to. The broker list string has the form HOST1:PORT1,HOST2:PORT2.",
		"consumer-property":     "A mechanism to pass user-defined properties in the form key=value to the consumer.",
		"consumer.config":       "Consumer config properties file. Note that [consumer-property] takes precedence over this config.",
		"enable-systest-events": "Log lifecycle events of the consumer in addition to logging consumed messages. (This is specific for system tests.)",
		"formatter":             `The name of a class to use for formatting kafka messages for display. (default "kafka.tools.DefaultMessageFormatter")`,
		"from-beginning":        "If the consumer does not already have an established offset to consume from, start with the earliest message present in the log rather than the latest message.",
		"group":                 "The consumer group id of the consumer.",
		"isolation-level":       `Set to read_committed in order to filter out transactional messages which are not committed. Set to read_uncommitted to read all messages. (default "read_uncommitted")`,
		"key-deserializer":      "",
		"max-messages":          "The maximum number of messages to consume before exiting. If not set, consumption is continual.",
		"offset":                `The offset id to consume from (a non-negative number), or "earliest" which means from beginning, or "latest" which means from end. (default "latest")`,
		"partition":             `The partition to consume from. Consumption starts from the end of the partition unless "--offset" is specified.`,
		"property":              "The properties to initialize the message formatter. Default properties include:\n\tprint.timestamp=true|false\n\tprint.key=true|false\n\tprint.value=true|false\n\tkey.separator=<key.separator>\n\tline.separator=<line.separator>\n\tkey.deserializer=<key.deserializer>\n\tvalue.deserializer=<value.deserializer>\nUsers can also pass in customized properties for their formatter; more specifically, users can pass in properties keyed with \"key.deserializer.\" and \"value.deserializer.\" prefixes to configure their deserializers.",
		"skip-message-on-error": "If there is an error when processing a message, skip it instead of halting.",
		"timeout-ms":            "If specified, exit if no messages are available for consumption for the specified interval.",
		"value-deserializer":    "",
		"whitelist":             "Regular expression specifying whitelist of topics to include for consumption.",
	}
	kafkaConsumeDefaultValues = map[string]any{
		"bootstrap-server":      defaultString,
		"consumer-property":     defaultString,
		"consumer.config":       defaultString,
		"enable-systest-events": defaultBool,
		"formatter":             defaultString,
		"from-beginning":        defaultBool,
		"group":                 defaultString,
		"isolation-level":       defaultString,
		"key-deserializer":      defaultString,
		"max-messages":          defaultInt,
		"offset":                defaultString,
		"partition":             defaultInt,
		"property":              defaultStringSlice,
		"skip-message-on-error": defaultBool,
		"timeout-ms":            defaultInt,
		"value-deserializer":    defaultString,
		"whitelist":             defaultString,
	}

	kafkaProduceFlagUsage = map[string]string{
		"bootstrap-server":           "The server(s) to connect to. The broker list string has the form HOST1:PORT1,HOST2:PORT2.",
		"batch-size":                 "Number of messages to send in a single batch if they are not being sent synchronously. (default 200)",
		"compression-codec":          `The compression codec: either "none", "gzip", "snappy", "lz4", or "zstd". If specified without value, then it defaults to "gzip".`,
		"line-reader":                `The class name of the class to use for reading lines from stdin. By default each line is read as a separate message. (default "kafka.tools.ConsoleProducer$LineMessageReader")`,
		"max-block-ms":               "The max time that the producer will block for during a send request. (default 60000)",
		"max-memory-bytes":           "The total memory used by the producer to buffer records waiting to be sent to the server. (default 33554432)",
		"max-partition-memory-bytes": "The buffer size allocated for a partition. When records are received which are small than this size, the producer will attempt to optimistically group them together until this size is reached. (default 16384)",
		"message-send-max-retries":   "This property specifies the number of retries before the producer gives up and drops this message. Brokers can fail receiving a message for multiple reasons, and being unavailable transiently is just one of them. (default 3)",
		"metadata-expiry-ms":         "The amount of time in milliseconds before a forced metadata refresh. This will occur independent of any leadership changes. (default 300000)",
		"producer-property":          "A mechanism to pass user-defined properties in the form key=value to the producer.",
		"producer.config":            "Producer config properties file. Note that [producer-property] takes precedence over this config.",
		"property":                   "A mechanism to pass user-defined properties in the form key=value to the message reader. This allows custom configuration for a user-defined message reader. Default properties include:\n\tparse.key=true|false\n\tkey.separator=<key.separator>\n\tignore.error=true|false",
		"request-required-acks":      "The required ACKs of the producer requests. (default 1)",
		"request-timeout-ms":         "The ACK timeout of the producer requests. Value must be positive. (default 1500)",
		"retry-backoff-ms":           "Before each retry, the producer refreshes the metadata of relevant topics. Since leader election takes a bit of time, this property specifies the amount of time that the producer waits before refreshing the metadata. (default 100)",
		"socket-buffer-size":         "The size of the TCP RECV size. (default 102400)",
		"sync":                       "If set, message send requests to brokers arrive synchronously.",
		"timeout":                    "If set and the producer is running in asynchronous mode, this gives the maximum amount of time a message will queue awaiting sufficient batch size. The value is given in ms. (default 1000)",
	}
	kafkaProduceDefaultValues = map[string]any{
		"batch-size":                 defaultInt,
		"bootstrap-server":           defaultString,
		"compression-codec":          defaultString,
		"line-reader":                defaultString,
		"max-block-ms":               defaultInt,
		"max-memory-bytes":           defaultInt,
		"max-partition-memory-bytes": defaultInt,
		"message-send-max-retries":   defaultInt,
		"metadata-expiry-ms":         defaultInt,
		"producer-property":          defaultString,
		"producer.config":            defaultString,
		"property":                   defaultStringSlice,
		"request-required-acks":      defaultString,
		"request-timeout-ms":         defaultInt,
		"retry-backoff-ms":           defaultInt,
		"socket-buffer-size":         defaultInt,
		"sync":                       defaultBool,
		"timeout":                    defaultInt,
	}
)

func NewKafkaConsumeCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "consume <topic>",
			Short: "Consume from a Kafka topic.",
			Long:  "Consume data from topics. By default this command consumes binary data from the Apache Kafka® cluster on localhost.",
			Args:  cobra.ExactArgs(1),
			Example: examples.BuildExampleString(
				examples.Example{
					Text: "Consume Avro data from the beginning of topic called `mytopic1` on a development Kafka cluster on localhost. Assumes Confluent Schema Registry is listening at `http://localhost:8081`.",
					Code: "confluent local services kafka consume mytopic1 --value-format avro --from-beginning",
				},
				examples.Example{
					Text: "Consume newly arriving non-Avro data from a topic called `mytopic2` on a development Kafka cluster on localhost.",
					Code: "confluent local services kafka consume mytopic2",
				},
			),
		}, prerunner)

	c.Command.RunE = c.runKafkaConsumeCommand
	c.initFlags("consume")

	return c.Command
}

func (c *Command) runKafkaConsumeCommand(cmd *cobra.Command, args []string) error {
	return c.runKafkaCommand(cmd, args, "consume", kafkaConsumeDefaultValues)
}

func NewKafkaProduceCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "produce <topic>",
			Short: "Produce to a Kafka topic.",
			Long:  "Produce data to topics. By default this command produces non-Avro data to the Apache Kafka® cluster on localhost.",
			Args:  cobra.ExactArgs(1),
			Example: examples.BuildExampleString(
				examples.Example{
					Text: "Produce Avro data to a topic called `mytopic1` on a development Kafka cluster on localhost. Assumes Confluent Schema Registry is listening at `http://localhost:8081`.",
					Code: "confluent local services kafka produce mytopic1 --value-format avro --property value.schema='{\"type\":\"record\",\"name\":\"myrecord\",\"fields\":[{\"name\":\"f1\",\"type\":\"string\"}]}'",
				},
				examples.Example{
					Text: "Produce non-Avro data to a topic called `mytopic2` on a development Kafka cluster on localhost:",
					Code: "confluent local produce mytopic2",
				},
			),
		}, prerunner)

	c.Command.RunE = c.runKafkaProduceCommand
	c.initFlags("produce")

	return c.Command
}

func (c *Command) runKafkaProduceCommand(cmd *cobra.Command, args []string) error {
	return c.runKafkaCommand(cmd, args, "produce", kafkaProduceDefaultValues)
}

func (c *Command) initFlags(mode string) {
	// CLI Flags
	c.Flags().Bool("cloud", defaultBool, commonFlagUsage["cloud"])
	defaultConfig := fmt.Sprintf("%s/.confluent/config", os.Getenv("HOME"))
	c.Flags().String("config", defaultConfig, commonFlagUsage["config"])
	c.Flags().String("value-format", defaultString, commonFlagUsage["value-format"]+"\n") // "\n" separates the CLI flags from the Kafka flags

	// Kafka Flags
	defaults := kafkaConsumeDefaultValues
	usage := kafkaConsumeFlagUsage
	if mode == "produce" {
		defaults = kafkaProduceDefaultValues
		usage = kafkaProduceFlagUsage
	}

	flags := types.GetSortedKeys(defaults)

	for _, flag := range flags {
		switch val := defaults[flag].(type) {
		case bool:
			c.Flags().Bool(flag, val, usage[flag])
		case int:
			c.Flags().Int(flag, val, usage[flag])
		case string:
			c.Flags().String(flag, val, usage[flag])
		case []string:
			c.Flags().StringSlice(flag, val, usage[flag])
		}
	}
}

func (c *Command) runKafkaCommand(cmd *cobra.Command, args []string, mode string, kafkaFlagTypes map[string]any) error {
	cloud, err := cmd.Flags().GetBool("cloud")
	if err != nil {
		return err
	}

	bootSet := cmd.Flags().Changed("bootstrap-server")

	// Only check if local Kafka is up if we are really connecting to a local Kafka
	if !(cloud || bootSet) {
		isUp, err := c.isRunning("kafka")
		if err != nil {
			return err
		}
		if !isUp {
			return c.printStatus("kafka")
		}
	}

	valueFormat, err := cmd.Flags().GetString("value-format")
	if err != nil {
		return err
	}

	// "consume" -> "consumer"
	modeNoun := fmt.Sprintf("%sr", mode)

	scriptFile, err := c.ch.GetKafkaScript(valueFormat, modeNoun)
	if err != nil {
		return err
	}

	var config string
	var cloudServer string

	if cloud {
		config, err = cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		data, err := os.ReadFile(config)
		if err != nil {
			return err
		}

		config := local.ExtractConfig(data)
		cloudServer = config["bootstrap.servers"].(string)

		configFileFlag := fmt.Sprintf("%s.config", modeNoun)
		delete(kafkaFlagTypes, configFileFlag)
		delete(kafkaFlagTypes, "bootstrap-server")
	}

	kafkaArgs, err := local.CollectFlags(cmd.Flags(), kafkaFlagTypes)
	if err != nil {
		return err
	}

	kafkaArgs = append(kafkaArgs, "--topic", args[0])
	if cloud {
		configFileFlag := fmt.Sprintf("--%s.config", modeNoun)
		kafkaArgs = append(kafkaArgs, configFileFlag, config)
		kafkaArgs = append(kafkaArgs, "--bootstrap-server", cloudServer)
	} else {
		if !types.Contains(kafkaArgs, "--bootstrap-server") {
			defaultBootstrapServer := fmt.Sprintf("localhost:%d", services["kafka"].port)
			kafkaArgs = append(kafkaArgs, "--bootstrap-server", defaultBootstrapServer)
		}
	}

	kafkaCommand := exec.Command(scriptFile, kafkaArgs...)
	kafkaCommand.Stdout = os.Stdout
	kafkaCommand.Stderr = os.Stderr
	if mode == "produce" {
		kafkaCommand.Stdin = os.Stdin
		output.Println("Exit with Ctrl-D")
	}

	kafkaCommand.Env = []string{
		fmt.Sprintf("LOG_DIR=%s", os.TempDir()),
	}
	if mode == "consume" {
		kafkaCommand.Env = append(kafkaCommand.Env, "SCHEMA_REGISTRY_LOG4J_LOGGERS=\"INFO, stdout\"")
	}

	return kafkaCommand.Run()
}
