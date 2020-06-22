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

const (
	defaultBool   = false
	defaultInt    = 0
	defaultString = ""
)

var (
	commonFlagUsage = map[string]string{
		"bootstrap-server": "The server(s) to connect to. The broker list string has the form HOST1:PORT1,HOST2:PORT2.",
		"cloud":            "Consume from Confluent Cloud.",
		"config":           "Change the ccloud configuration file.",
		"value-format":     "Format output data: avro, json, or protobuf.",
	}

	kafkaConsumeFlagUsage = map[string]string{
		"consumer-property":     "A mechanism to pass user-defined properties in the form key=value to the consumer.",
		"consumer.config":       "Consumer config properties file. Note that [consumer-property] takes precedence over this config.",
		"enable-systest-events": "Log lifecycle events of the consumer in addition to logging consumed messages. (This is specific for system tests.)",
		"formatter":             "The name of a class to use for formatting kafka messages for display. (default \"kafka.tools.DefaultMessageFormatter\")",
		"from-beginning":        "If the consumer does not already have an established offset to consume from, start with the earliest message present in the log rather than the latest message.",
		"group":                 "The consumer group id of the consumer.",
		"isolation-level":       "Set to read_committed in order to filter out transactional messages which are not committed. Set to read_uncommitted to read all messages. (default \"read_uncommitted\")",
		"key-deserializer":      "",
		"max-messages":          "The maximum number of messages to consume before exiting. If not set, consumption is continual.",
		"offset":                "The offset id to consume from (a non-negative number), or \"earliest\" which means from beginning, or \"latest\" which means from end (default \"latest\")",
		"partition":             "The partition to consume from. Consumption starts from the end of the partition unless \"--offset\" is specified.",
		"property":              "The properties to initialize the message formatter. Default properties include:\n\tprint.timestamp=true|false\n\tprint.key=true|false\n\tprint.value=true|false\n\tkey.separator=<key.separator>\n\tline.separator=<line.separator>\n\tkey.deserializer=<key.deserializer>\n\tvalue.deserializer=<value.deserializer>\nUsers can also pass in customized properties for their formatter; more specifically, users can pass in properties keyed with \"key.deserializer.\" and \"value.deserializer.\" prefixes to configure their deserializers.",
		"skip-message-on-error": "If there is an error when processing a message, skip it instead of halting.",
		"timeout-ms":            "If specified, exit if no messages are available for consumption for the specified interval.",
		"value-deserializer":    "",
		"whitelist":             "Regular expression specifying whitelist of topics to include for consumption.",
	}
	kafkaConsumeDefaultValues = map[string]interface{}{
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

	kafkaProduceFlagUsage = map[string]string{
		"batch-size":                 "Number of messages to send in a single batch if they are not being sent synchronously. (default 200)",
		"compression-codec":          "The compression codec: either \"none\", \"gzip\", \"snappy\", \"lz4\", or \"zstd\". If specified without value, the it defaults to \"gzip\".",
		"line-reader":                "The class name of the class to use for reading lines from stdin. By default each line is read as a separate message. (default \"kafka.tools.ConsoleProducer$LineMessageReader\")",
		"max-block-ms":               "The max time that the producer will block for during a send request (default 60000)",
		"max-memory-bytes":           "The total memory used by the producer to buffer records waiting to be sent to the server. (default 33554432)",
		"max-partition-memory-bytes": "The buffer size allocated for a partition. When records are received which are small than this size, the producer will attempt to optimistically group them together until this size is reached. (default 16384)",
		"message-send-max-retries":   "Brokers can fail receiving a message for multiple reasons, and being unavailable transiently is just one of them. This property specifies the number of retries before the producer gives up and drops this message. (default 3)",
		"metadata-expiry-ms":         "The period of time in milliseconds after which we force a refresh of metadata even if we haven't seen any leadership changes. (default 300000)",
		"producer-property":          "A mechanism to pass user-defined properties in the form key=value to the producer.",
		"producer.config":            "Producer config properties file. Note that [producer-property] takes precedence over this config.",
		"property":                   "A mechanism to pass user-defined properties in the form key=value to the message reader. This allows custom configuration for a user-defined message reader. Default properties include:\n\tparse.key=true|false\n\tkey.separator=<key.separator>\n\tignore.error=true|false",
		"request-required-acks":      "The required acks of the producer requests (default 1)",
		"request-timeout-ms":         "The ack timeout of the producer requests. Value must be positive (default 1500)",
		"retry-backoff-ms":           "Before each retry, the producer refreshes the metadata of relevant topics. Since leader election takes a bit of time, this property specifies the amount of time that the producer waits before refreshing the metadata. (default 100)",
		"socket-buffer-size":         "The size of the TCP RECV size. (default 102400)",
		"sync":                       "If set, message send requests to brokers arrive synchronously.",
		"timeout":                    "If set and the producer is running in asynchronous mode, this gives the maximum amount of time a message will queue awaiting sufficient batch size. The value is given in ms. (default 1000)",
	}
	kafkaProduceDefaultValues = map[string]interface{}{
		"batch-size":                 defaultInt,
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
	kafkaConsumeCommand.Flags().Bool("cloud", defaultBool, commonFlagUsage["cloud"])
	defaultConfig := fmt.Sprintf("%s/.ccloud/config", os.Getenv("HOME"))
	kafkaConsumeCommand.Flags().String("config", defaultConfig, commonFlagUsage["config"])
	kafkaConsumeCommand.Flags().String("value-format", defaultString, commonFlagUsage["value-format"])

	// Kafka Flags
	defaultBootstrapServer := fmt.Sprintf("localhost:%d", services["kafka"].port)
	kafkaConsumeCommand.Flags().String("bootstrap-server", defaultBootstrapServer, commonFlagUsage["bootstrap-server"])
	for flag, val := range kafkaConsumeDefaultValues {
		switch val.(type) {
		case bool:
			kafkaConsumeCommand.Flags().Bool(flag, val.(bool), kafkaConsumeFlagUsage[flag])
		case int:
			kafkaConsumeCommand.Flags().Int(flag, val.(int), kafkaConsumeFlagUsage[flag])
		case string:
			kafkaConsumeCommand.Flags().String(flag, val.(string), kafkaConsumeFlagUsage[flag])
		}
	}

	return kafkaConsumeCommand.Command
}

func runKafkaConsumeCommand(command *cobra.Command, args []string) error {
	return runKafkaCommand(command, args, "consume", kafkaConsumeDefaultValues)
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
	kafkaProduceCommand.Flags().Bool("cloud", defaultBool, commonFlagUsage["cloud"])
	defaultConfig := fmt.Sprintf("%s/.ccloud/config", os.Getenv("HOME"))
	kafkaProduceCommand.Flags().String("config", defaultConfig, commonFlagUsage["config"])
	kafkaProduceCommand.Flags().String("value-format", defaultString, commonFlagUsage["value-format"])

	// Kafka Flags
	defaultBootstrapServer := fmt.Sprintf("localhost:%d", services["kafka"].port)
	kafkaProduceCommand.Flags().String("bootstrap-server", defaultBootstrapServer, commonFlagUsage["bootstrap-server"])
	for flag, val := range kafkaProduceDefaultValues {
		switch val.(type) {
		case bool:
			kafkaProduceCommand.Flags().Bool(flag, val.(bool), kafkaProduceFlagUsage[flag])
		case int:
			kafkaProduceCommand.Flags().Int(flag, val.(int), kafkaProduceFlagUsage[flag])
		case string:
			kafkaProduceCommand.Flags().String(flag, val.(string), kafkaProduceFlagUsage[flag])
		}
	}

	return kafkaProduceCommand.Command
}

func runKafkaProduceCommand(command *cobra.Command, args []string) error {
	return runKafkaCommand(command, args, "produce", kafkaProduceDefaultValues)
}

func runKafkaCommand(command *cobra.Command, args []string, mode string, kafkaFlagTypes map[string]interface{}) error {
	format, err := command.Flags().GetString("value-format")
	if err != nil {
		return err
	}

	// "consume" -> "consumer"
	// "produce" -> "producer"
	modeNoun := fmt.Sprintf("%sr", mode)

	ch := local.NewConfluentHomeManager()

	scriptFile, err := ch.GetKafkaScript(format, modeNoun)
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

		config := local.ExtractConfig(data)
		cloudServer = config["bootstrap.servers"]
	}

	if cloud {
		configFileFlag := fmt.Sprintf("%s.config", modeNoun)
		delete(kafkaFlagTypes, configFileFlag)
		delete(kafkaFlagTypes, "bootstrap-server")
	}

	kafkaArgs, err := collectFlags(command.Flags(), kafkaFlagTypes)
	if err != nil {
		return err
	}

	kafkaArgs = append(kafkaArgs, "--topic", args[0])
	if cloud {
		configFileFlag := fmt.Sprintf("--%s.config", modeNoun)
		kafkaArgs = append(kafkaArgs, configFileFlag, cloudConfigFile)
		kafkaArgs = append(kafkaArgs, "--bootstrap-server", cloudServer)
	}

	kafkaCommand := exec.Command(scriptFile, kafkaArgs...)
	kafkaCommand.Stdout = os.Stdout
	kafkaCommand.Stderr = os.Stderr
	if mode == "produce" {
		kafkaCommand.Stdin = os.Stdin
		fmt.Println("Exit with Ctrl+D")
	}

	return kafkaCommand.Run()
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
