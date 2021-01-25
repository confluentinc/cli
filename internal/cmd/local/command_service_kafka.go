package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/local"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	defaultBool        bool
	defaultInt         int
	defaultString      string
	defaultStringArray []string

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
		"property":              defaultStringArray,
		"skip-message-on-error": defaultBool,
		"timeout-ms":            defaultInt,
		"value-deserializer":    defaultString,
		"whitelist":             defaultString,
	}

	kafkaProduceFlagUsage = map[string]string{
		"bootstrap-server":           "The server(s) to connect to. The broker list string has the form HOST1:PORT1,HOST2:PORT2.",
		"batch-size":                 "Number of messages to send in a single batch if they are not being sent synchronously. (default 200)",
		"compression-codec":          "The compression codec: either \"none\", \"gzip\", \"snappy\", \"lz4\", or \"zstd\". If specified without value, the it defaults to \"gzip\".",
		"line-reader":                "The class name of the class to use for reading lines from stdin. By default each line is read as a separate message. (default \"kafka.tools.ConsoleProducer$LineMessageReader\")",
		"max-block-ms":               "The max time that the producer will block for during a send request (default 60000)",
		"max-memory-bytes":           "The total memory used by the producer to buffer records waiting to be sent to the server. (default 33554432)",
		"max-partition-memory-bytes": "The buffer size allocated for a partition. When records are received which are small than this size, the producer will attempt to optimistically group them together until this size is reached. (default 16384)",
		"message-send-max-retries":   "This property specifies the number of retries before the producer gives up and drops this message. Brokers can fail receiving a message for multiple reasons, and being unavailable transiently is just one of them. (default 3)",
		"metadata-expiry-ms":         "The amount of time in milliseconds before a forced metadata refresh. This will occur independent of any leadership changes. (default 300000)",
		"producer-property":          "A mechanism to pass user-defined properties in the form key=value to the producer.",
		"producer.config":            "Producer config properties file. Note that [producer-property] takes precedence over this config.",
		"property":                   "A mechanism to pass user-defined properties in the form key=value to the message reader. This allows custom configuration for a user-defined message reader. Default properties include:\n\tparse.key=true|false\n\tkey.separator=<key.separator>\n\tignore.error=true|false",
		"request-required-acks":      "The required ACKs of the producer requests (default 1)",
		"request-timeout-ms":         "The ACK timeout of the producer requests. Value must be positive (default 1500)",
		"retry-backoff-ms":           "Before each retry, the producer refreshes the metadata of relevant topics. Since leader election takes a bit of time, this property specifies the amount of time that the producer waits before refreshing the metadata. (default 100)",
		"socket-buffer-size":         "The size of the TCP RECV size. (default 102400)",
		"sync":                       "If set, message send requests to brokers arrive synchronously.",
		"timeout":                    "If set and the producer is running in asynchronous mode, this gives the maximum amount of time a message will queue awaiting sufficient batch size. The value is given in ms. (default 1000)",
	}
	kafkaProduceDefaultValues = map[string]interface{}{
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
		"property":                   defaultStringArray,
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
					Text: "Consume Avro data from the beginning of topic called ``mytopic1`` on a development Kafka cluster on localhost. Assumes Confluent Schema Registry is listening at ``http://localhost:8081``.",
					Code: "confluent local services kafka consume mytopic1 --value-format avro --from-beginning",
				},
				examples.Example{
					Text: "Consume newly arriving non-Avro data from a topic called ``mytopic2`` on a development Kafka cluster on localhost.",
					Code: "confluent local services kafka consume mytopic2",
				},
				examples.Example{
					Text: "Create a Confluent Cloud configuration file with connection details for the Confluent Cloud cluster using the format shown in this example, and save as ``/tmp/myconfig.properties``. You can specify the file location using ``--config <filename>``.",
					Code: "bootstrap.servers=<broker endpoint>\nsasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username=\"<api-key>\" password=\"<api-secret>\";\nbasic.auth.credentials.source=USER_INFO\nschema.registry.basic.auth.user.info=<username:password>\nschema.registry.url=<sr endpoint>",
				},
				examples.Example{
					Text: "Consume non-Avro data from the beginning of a topic named ``mytopic3`` in Confluent Cloud, using a user-specified Confluent Cloud configuration file at ``/tmp/myconfig.properties``.",
					Code: "confluent local services kafka consume mytopic3 --cloud --config /tmp/myconfig.properties --from-beginning",
				},
				examples.Example{
					Text: "Consume messages with keys and non-Avro values from the beginning of topic called ``mytopic4`` in Confluent Cloud, using a user-specified Confluent Cloud configuration file at ``/tmp/myconfig.properties``. See the sample Confluent Cloud configuration file above.",
					Code: "confluent local services kafka consume mytopic4 --cloud --config /tmp/myconfig.properties --from-beginning --property print.key=true",
				},
				examples.Example{
					Text: "Consume Avro data from a topic called ``mytopic5`` in Confluent Cloud. Assumes Confluent Schema Registry is listening at ``http://localhost:8081``.",
					Code: "confluent local services kafka consume mytopic5 --cloud --config /tmp/myconfig.properties --value-format avro \\\n--from-beginning --property schema.registry.url=http://localhost:8081",
				},
				examples.Example{
					Text: "Consume Avro data from a topic called ``mytopic6`` in Confluent Cloud. Assumes you are using Confluent Cloud Confluent Schema Registry.",
					Code: "confluent local services kafka consume mytopic6 --cloud --config /tmp/myconfig.properties --value-format avro \\\n--from-beginning --property schema.registry.url=https://<SR ENDPOINT> \\\n--property basic.auth.credentials.source=USER_INFO \\\n--property schema.registry.basic.auth.user.info=<SR API KEY>:<SR API SECRET>",
				},
			),
		}, prerunner)

	c.Command.RunE = cmd.NewCLIRunE(c.runKafkaConsumeCommand)
	c.initFlags("consume")

	return c.Command
}

func (c *Command) runKafkaConsumeCommand(command *cobra.Command, args []string) error {
	return c.runKafkaCommand(command, args, "consume", kafkaConsumeDefaultValues)
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
					Text: "Produce Avro data to a topic called ``mytopic1`` on a development Kafka cluster on localhost. Assumes Confluent Schema Registry is listening at ``http://localhost:8081``.",
					Code: "confluent local services kafka produce mytopic1 --value-format avro --property value.schema='{\"type\":\"record\",\"name\":\"myrecord\",\"fields\":[{\"name\":\"f1\",\"type\":\"string\"}]}'",
				},
				examples.Example{
					Text: "Produce non-Avro data to a topic called ``mytopic2`` on a development Kafka cluster on localhost:",
					Code: "confluent local produce mytopic2",
				},
				examples.Example{
					Text: "Create a customized Confluent Cloud configuration file with connection details for the Confluent Cloud cluster using the format shown in this example, and save as ``/tmp/myconfig.properties``. You can specify the file location using ``--config <filename>``.",
					Code: "bootstrap.servers=<broker endpoint>\nsasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username=\"<api-key>\" password=\"<api-secret>\";\nbasic.auth.credentials.source=USER_INFO\nschema.registry.basic.auth.user.info=<username:password>\nschema.registry.url=<sr endpoint>",
				},
				examples.Example{
					Text: "Produce non-Avro data to a topic called ``mytopic3`` in Confluent Cloud. Assumes topic has already been created.",
					Code: "confluent local services kafka produce mytopic3 --cloud --config /tmp/myconfig.properties",
				},
				examples.Example{
					Text: "Produce messages with keys and non-Avro values to a topic called ``mytopic4`` in Confluent Cloud, using a user-specified Confluent Cloud configuration file at ``/tmp/myconfig.properties``. Assumes topic has already been created.",
					Code: "confluent local services kafka produce mytopic4 --cloud --config /tmp/myconfig.properties --property parse.key=true --property key.separator=,",
				},
				examples.Example{
					Text: "Produce Avro data to a topic called ``mytopic5`` in Confluent Cloud. Assumes topic has already been created, and Confluent Schema Registry is listening at ``http://localhost:8081``.",
					Code: `confluent local services kafka produce mytopic5 --cloud --config /tmp/myconfig.properties --value-format avro --property \\\nvalue.schema='{"type":"record","name":"myrecord","fields":[{"name":"f1","type":"string"}]}' \\\n--property schema.registry.url=http://localhost:8081`,
				},
				examples.Example{
					Text: "Produce Avro data to a topic called ``mytopic6`` in Confluent Cloud. Assumes topic has already been created and you are using Confluent Cloud Confluent Schema Registry.",
					Code: `confluent local services kafka produce mytopic5 --cloud --config /tmp/myconfig.properties --value-format avro --property \\\nvalue.schema='{"type":"record","name":"myrecord","fields":[{"name":"f1","type":"string"}]}' \\\n--property schema.registry.url=https://<SR ENDPOINT> \\\n--property basic.auth.credentials.source=USER_INFO \\\n--property schema.registry.basic.auth.user.info=<SR API KEY>:<SR API SECRET>`,
				},
			),
		}, prerunner)

	c.Command.RunE = cmd.NewCLIRunE(c.runKafkaProduceCommand)
	c.initFlags("produce")

	return c.Command
}

func (c *Command) runKafkaProduceCommand(command *cobra.Command, args []string) error {
	return c.runKafkaCommand(command, args, "produce", kafkaProduceDefaultValues)
}

func (c *Command) initFlags(mode string) {
	// CLI Flags
	c.Flags().Bool("cloud", defaultBool, commonFlagUsage["cloud"])
	defaultConfig := fmt.Sprintf("%s/.ccloud/config", os.Getenv("HOME"))
	c.Flags().String("config", defaultConfig, commonFlagUsage["config"])
	c.Flags().String("value-format", defaultString, commonFlagUsage["value-format"])

	// Kafka Flags
	defaults := kafkaConsumeDefaultValues
	usage := kafkaConsumeFlagUsage
	if mode == "produce" {
		defaults = kafkaProduceDefaultValues
		usage = kafkaProduceFlagUsage
	}

	for flag, val := range defaults {
		switch val.(type) {
		case bool:
			c.Flags().Bool(flag, val.(bool), usage[flag])
		case int:
			c.Flags().Int(flag, val.(int), usage[flag])
		case string:
			c.Flags().String(flag, val.(string), usage[flag])
		case []string:
			c.Flags().StringArray(flag, val.([]string), usage[flag])
		}
	}
}

func (c *Command) runKafkaCommand(command *cobra.Command, args []string, mode string, kafkaFlagTypes map[string]interface{}) error {
	cloud, err := command.Flags().GetBool("cloud")
	if err != nil {
		return err
	}

	bootSet := command.Flags().Changed("bootstrap-server")

	// Only check if local Kafka is up if we are really connecting to a local Kafka
	if !(cloud || bootSet) {
		isUp, err := c.isRunning("kafka")
		if err != nil {
			return err
		}
		if !isUp {
			return c.printStatus(command, "kafka")
		}
	}

	format, err := command.Flags().GetString("value-format")
	if err != nil {
		return err
	}

	// "consume" -> "consumer"
	modeNoun := fmt.Sprintf("%sr", mode)

	scriptFile, err := c.ch.GetKafkaScript(format, modeNoun)
	if err != nil {
		return err
	}

	var cloudConfigFile string
	var cloudServer string

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
		cloudServer = config["bootstrap.servers"].(string)

		configFileFlag := fmt.Sprintf("%s.config", modeNoun)
		delete(kafkaFlagTypes, configFileFlag)
		delete(kafkaFlagTypes, "bootstrap-server")
	}

	kafkaArgs, err := local.CollectFlags(command.Flags(), kafkaFlagTypes)
	if err != nil {
		return err
	}

	kafkaArgs = append(kafkaArgs, "--topic", args[0])
	if cloud {
		configFileFlag := fmt.Sprintf("--%s.config", modeNoun)
		kafkaArgs = append(kafkaArgs, configFileFlag, cloudConfigFile)
		kafkaArgs = append(kafkaArgs, "--bootstrap-server", cloudServer)
	} else {
		if !utils.Contains(kafkaArgs, "--bootstrap-server") {
			defaultBootstrapServer := fmt.Sprintf("localhost:%d", services["kafka"].port)
			kafkaArgs = append(kafkaArgs, "--bootstrap-server", defaultBootstrapServer)
		}
	}

	kafkaCommand := exec.Command(scriptFile, kafkaArgs...)
	kafkaCommand.Stdout = os.Stdout
	kafkaCommand.Stderr = os.Stderr
	if mode == "produce" {
		kafkaCommand.Stdin = os.Stdin
		utils.Println(command, "Exit with Ctrl+D")
	}

	kafkaCommand.Env = []string{
		fmt.Sprintf("LOG_DIR=%s", os.TempDir()),
	}
	if mode == "consume" {
		kafkaCommand.Env = append(kafkaCommand.Env, "SCHEMA_REGISTRY_LOG4J_LOGGERS=\"INFO, stdout\"")
	}

	return kafkaCommand.Run()
}
