package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/local"
)

var (
	defaultBool   bool
	defaultInt    int
	defaultString string
)

func NewKafkaConsumeCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	kafkaConsumeCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "consume [topic]",
			Short: "Consume from a kafka topic.",
			Args:  cobra.ExactArgs(1),
			RunE:  runKafkaConsumeCommand,
		},
		cfg, prerunner)

	// CLI Flags
	kafkaConsumeCommand.Flags().Bool("cloud", defaultBool, "Consume from Confluent Cloud.")
	defaultConfig := fmt.Sprintf("%s/.ccloud/config", os.Getenv("HOME")) // TODO: Supposed to be config.json?
	kafkaConsumeCommand.Flags().String("config", defaultConfig, "Change the ccloud configuration file.")
	kafkaConsumeCommand.Flags().String("value-format", defaultString, "Format output data: avro, json, or protobuf.")

	// Kafka Flags
	defaultBootstrapServer := fmt.Sprintf("localhost:%d", services["kafka"].port)
	kafkaConsumeCommand.Flags().String("bootstrap-server", defaultBootstrapServer, "The server(s) to connect to. The broker list string has the form HOST1:PORT1,HOST2:PORT2.")
	kafkaConsumeCommand.Flags().String("consumer-property", defaultString, "A mechanism to pass user-defined properties in the form key=value to the consumer.")
	kafkaConsumeCommand.Flags().String("consumer.config", defaultString, "Consumer config properties file. Note that [consumer-property] takes precedence over this config.")
	kafkaConsumeCommand.Flags().Bool("enable-systest-events", defaultBool, "Log lifecycle events of the consumer in addition to logging consumed messages. (This is specific for system tests.)")
	kafkaConsumeCommand.Flags().String("formatter", defaultString, "The name of a class to use for formatting kafka messages for display. (default: kafka.tools.DefaultMessageFormatter)")
	kafkaConsumeCommand.Flags().Bool("from-beginning", defaultBool, "If the consumer does not already have an established offset to consume from, start with the earliest message present in the log rather than the latest message.")
	kafkaConsumeCommand.Flags().String("group", defaultString, "The consumer group id of the consumer.")
	kafkaConsumeCommand.Flags().String("isolation-level", defaultString, "Set to read_committed in order to filter out transactional messages which are not committed. Set to read_uncommitted to read all messages. (default: read_uncommitted)")
	kafkaConsumeCommand.Flags().String("key-deserializer", defaultString, "")
	kafkaConsumeCommand.Flags().Int("max-messages", defaultInt, "The maximum number of messages to consume before exiting. If not set, consumption is continual.")
	kafkaConsumeCommand.Flags().String("offset", defaultString, "The offset id to consume from (a non-negative number), or 'earliest' which means from beginning, or 'latest' which means from end (default: latest)")
	kafkaConsumeCommand.Flags().Int("partition", defaultInt, "The partition to consume from. Consumption starts from the end of the partition unless '--offset' is specified.")
	kafkaConsumeCommand.Flags().String("property", defaultString, "The properties to initialize the message formatter. Default properties include:\n\tprint.timestamp=true|false\n\tprint.key=true|false\n\tprint.value=true|false\n\tkey.separator=<key.separator>\n\tline.separator=<line.separator>\n\tkey.deserializer=<key.deserializer>\n\tvalue.deserializer=<value.deserializer>\nUsers can also pass in customized properties for their formatter; more specifically, users can pass in properties keyed with 'key.deserializer.' and 'value.deserializer.' prefixes to configure their deserializers.")
	kafkaConsumeCommand.Flags().Bool("skip-message-on-error", defaultBool, "If there is an error when processing a message, skip it instead of halting.")
	kafkaConsumeCommand.Flags().Int("timeout-ms", defaultInt, "If specified, exit if no messages are available for consumption for the specified interval.")
	kafkaConsumeCommand.Flags().String("value-deserializer", defaultString, "")
	kafkaConsumeCommand.Flags().String("whitelist", defaultString, "Regular expression specifying whitelist of topics to include for consumption.")

	return kafkaConsumeCommand.Command
}

func runKafkaConsumeCommand(command *cobra.Command, args []string) error {
	format, err := command.Flags().GetString("value-format")
	if err != nil {
		return err
	}

	ch := local.NewConfluentHomeManager()
	scriptFile, err := ch.GetKafkaScriptFile(format, "consumer")
	if err != nil {
		return err
	}

	var cloudConfigFile string
	var cloudServer string

	cloud, err := command.Flags().GetBool("cloud")
	if err != nil {
		return err
	}
	if cloud {
		cloudConfigFile, err = command.Flags().GetString("config")
		if err != nil {
			return err
		}
		if err := command.Flags().Set("producer.config", cloudConfigFile); err != nil {
			return err
		}

		data, err := ioutil.ReadFile(cloudConfigFile)
		if err != nil {
			return err
		}

		config := extractConfig(data)
		cloudServer = config["bootstrap.servers"]
	}

	kafkaFlagTypes := map[string]interface{}{
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
		"property":              defaultString,
		"skip-message-on-error": defaultBool,
		"timeout-ms":            defaultInt,
		"value-deserializer":    defaultString,
		"whitelist":             defaultString,
	}

	if cloud {
		delete(kafkaFlagTypes, "consumer.config")
		delete(kafkaFlagTypes, "bootstrap-server")
	}

	kafkaArgs, err := collectFlags(command.Flags(), kafkaFlagTypes)
	if err != nil {
		return err
	}

	kafkaArgs = append(kafkaArgs, "--topic", args[0])
	if cloud {
		kafkaArgs = append(kafkaArgs, "--consumer.config", cloudConfigFile)
		kafkaArgs = append(kafkaArgs, "--bootstrap-server", cloudServer)
	}

	consumer := exec.Command(scriptFile, kafkaArgs...)
	consumer.Stdout = os.Stdout
	consumer.Stderr = os.Stderr

	return consumer.Run()
}

func NewKafkaProduceCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	kafkaProduceCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "produce [topic]",
			Short: "Produce to a kafka topic.",
			Args:  cobra.ExactArgs(1),
			RunE:  runKafkaProduceCommand,
		},
		cfg, prerunner)

	// CLI Flags
	kafkaProduceCommand.Flags().Bool("cloud", defaultBool, "Consume from Confluent Cloud.")
	defaultConfig := fmt.Sprintf("%s/.ccloud/config", os.Getenv("HOME")) // TODO: Supposed to be config.json?
	kafkaProduceCommand.Flags().String("config", defaultConfig, "Change the ccloud configuration file.")
	kafkaProduceCommand.Flags().String("value-format", defaultString, "Format output data: avro, json, or protobuf.")

	// Kafka Flags
	kafkaProduceCommand.Flags().Int("batch-size", defaultInt, "Number of messages to send in a single batch if they are not being sent synchronously. (default: 200)")
	defaultBootstrapServer := fmt.Sprintf("localhost:%d", services["kafka"].port)
	kafkaProduceCommand.Flags().String("bootstrap-server", defaultBootstrapServer, "The server(s) to connect to. The broker list string has the form HOST1:PORT1,HOST2:PORT2.")
	kafkaProduceCommand.Flags().String("compression-codec", defaultString, "The compression codec: either 'none', 'gzip', 'snappy', 'lz4', or 'zstd'. If specified without value, the it defaults to 'gzip'.")
	kafkaProduceCommand.Flags().String("line-reader", defaultString, "The class name of the class to use for reading lines from stdin. By default each line is read as a separate message. (default: kafka.tools.ConsoleProducer$LineMessageReader)")
	kafkaProduceCommand.Flags().Int("max-block-ms", defaultInt, "The max time that the producer will block for during a send request (default: 60000)")
	kafkaProduceCommand.Flags().Int("max-memory-bytes", defaultInt, "The total memory used by the producer to buffer records waiting to be sent to the server. (default: 33554432)")
	kafkaProduceCommand.Flags().Int("max-partition-memory-bytes", defaultInt, "The buffer size allocated for a partition. When records are received which are small than this size, the producer will attempt to optimistically group them together until this size is reached. (default: 16384)")
	kafkaProduceCommand.Flags().Int("message-send-max-retries", defaultInt, "Brokers can fail receiving a message for multiple reasons, and being unavailable transiently is just one of them. This property specifies the number of retries before the producer gives up and drops this message. (default: 3)")
	kafkaProduceCommand.Flags().Int("metadata-expiry-ms", defaultInt, "The period of time in milliseconds after which we force a refresh of metadata even if we haven't seen any leadership changes. (default: 300000)")
	kafkaProduceCommand.Flags().String("producer-property", defaultString, "A mechanism to pass user-defined properties in the form key=value to the producer.")
	kafkaProduceCommand.Flags().String("producer.config", defaultString, "Producer config properties file. Note that [producer-property] takes precedence over this config.")
	kafkaProduceCommand.Flags().String("property", defaultString, "A mechanism to pass user-defined properties in the form key=value to the message reader. This allows custom configuration for a user-defined message reader. Default properties include:\n\tparse.key=true|false\n\tkey.separator=<key.separator>\n\tignore.error=true|false")
	kafkaProduceCommand.Flags().String("request-required-acks", defaultString, "The required acks of the producer requests (default: 1)")
	kafkaProduceCommand.Flags().Int("request-timeout-ms", defaultInt, "The ack timeout of the producer requests. Value must be positive (default: 1500)")
	kafkaProduceCommand.Flags().Int("retry-backoff-ms", defaultInt, "Before each retry, the producer refreshes the metadata of relevant topics. Since leader election takes a bit of time, this property specifies the amount of time that the producer waits before refreshing the metadata. (default: 100)")
	kafkaProduceCommand.Flags().Int("socket-buffer-size", defaultInt, "The size of the TCP RECV size. (default 102400)")
	kafkaProduceCommand.Flags().Bool("sync", defaultBool, "If set, message send requests to brokers arrive synchronously.")
	kafkaProduceCommand.Flags().Int("timeout", defaultInt, "If set and the producer is running in asynchronous mode, this gives the maximum amount of time a message will queue awaiting sufficient batch size. The value is given in ms. (default: 1000)")

	return kafkaProduceCommand.Command
}

func runKafkaProduceCommand(command *cobra.Command, args []string) error {
	format, err := command.Flags().GetString("value-format")
	if err != nil {
		return err
	}

	ch := local.NewConfluentHomeManager()
	scriptFile, err := ch.GetKafkaScriptFile(format, "producer")
	if err != nil {
		return err
	}

	var cloudConfigFile string
	var cloudServer string

	cloud, err := command.Flags().GetBool("cloud")
	if err != nil {
		return err
	}
	if cloud {
		cloudConfigFile, err = command.Flags().GetString("config")
		if err != nil {
			return err
		}

		data, err := ioutil.ReadFile(cloudConfigFile)
		if err != nil {
			return err
		}

		config := extractConfig(data)
		cloudServer = config["bootstrap.servers"]
	}

	kafkaFlagTypes := map[string]interface{}{
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
		"property":                   defaultString,
		"request-required-acks":      defaultString,
		"request-timeout-ms":         defaultInt,
		"retry-backoff-ms":           defaultInt,
		"socket-buffer-size":         defaultInt,
		"sync":                       defaultBool,
		"timeout":                    defaultInt,
	}

	if cloud {
		delete(kafkaFlagTypes, "consumer.config")
		delete(kafkaFlagTypes, "bootstrap-server")
	}

	kafkaArgs, err := collectFlags(command.Flags(), kafkaFlagTypes)
	if err != nil {
		return err
	}

	kafkaArgs = append(kafkaArgs, "--topic", args[0])
	if cloud {
		kafkaArgs = append(kafkaArgs, "--producer.config", cloudConfigFile)
		kafkaArgs = append(kafkaArgs, "--bootstrap-server", cloudServer)
	}

	producer := exec.Command(scriptFile, kafkaArgs...)
	producer.Stdin = os.Stdin
	producer.Stdout = os.Stdout
	producer.Stderr = os.Stderr

	fmt.Println("Exit with Ctrl+D")
	return producer.Run()
}

func collectFlags(flags *pflag.FlagSet, flagTypes map[string]interface{}) ([]string, error) {
	var args []string

	for key, typeDefault := range flagTypes {
		var val interface{}
		var err error

		switch typeDefault.(type) {
		case bool:
			val, err = flags.GetBool(key)
		case string:
			val, err = flags.GetString(key)
		case int:
			val, err = flags.GetInt(key)
		}

		if err != nil {
			return []string{}, err
		}
		if val == typeDefault {
			continue
		}

		flag := fmt.Sprintf("--%s", key)

		switch typeDefault.(type) {
		case bool:
			args = append(args, flag)
		case string:
			args = append(args, flag, val.(string))
		case int:
			args = append(args, flag, strconv.Itoa(val.(int)))
		}
	}

	return args, nil
}
