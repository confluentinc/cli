package kafka

// confluent kafka topic <commands>
import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/antihax/optional"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/serdes"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type PartitionData struct {
	TopicName              string  `json:"topic" yaml:"topic"`
	PartitionId            int32   `json:"partition" yaml:"partition"`
	LeaderBrokerId         int32   `json:"leader" yaml:"leader"`
	ReplicaBrokerIds       []int32 `json:"replicas" yaml:"replicas"`
	InSyncReplicaBrokerIds []int32 `json:"isr" yaml:"isr"`
}

type TopicData struct {
	TopicName         string            `json:"topic_name" yaml:"topic_name"`
	PartitionCount    int               `json:"partition_count" yaml:"partition_count"`
	ReplicationFactor int               `json:"replication_factor" yaml:"replication_factor"`
	Partitions        []PartitionData   `json:"partitions" yaml:"partitions"`
	Configs           map[string]string `json:"config" yaml:"config"`
}

type SchemaObject struct {
	Namespace  string                `json:"namespace"`
	SchemaType string                `json:"type"`
	SchemaName string                `json:"name"`
	Fields     [](map[string]string) `json:"fields"`
}

type SchemaRequest struct {
	Schema     string `json:"schema"`
	SchemaType string `json:"schemaType"`
}

// Register each of the verbs and expected args
func (c *authenticatedTopicCommand) onPremInit() {
	// Register list command
	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.onPremList),
		Short: "List Kafka topics.",
		Example: examples.BuildExampleString(
			examples.Example{
				// on-prem examples are ccloud examples + "of a specified cluster (providing Kafka REST Proxy endpoint)."
				Text: "List all topics of a specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "confluent kafka topic list --url http://localhost:8082",
			},
		),
	}
	listCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create <topic>",
		Short: "Create a Kafka topic.",
		Args:  cobra.ExactArgs(1), // <topic>
		RunE:  pcmd.NewCLIRunE(c.onPremCreate),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a topic named `my_topic` with default options at specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "confluent kafka topic create my_topic --url http://localhost:8082",
			},
			examples.Example{
				Text: "Create a topic named `my_topic_2` with specified configuration parameters.",
				Code: "confluent kafka topic create my_topic_2 --url http://localhost:8082 --config cleanup.policy=compact,compression.type=gzip",
			}),
	}
	createCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags
	createCmd.Flags().Int32("partitions", 6, "Number of topic partitions.")
	createCmd.Flags().Int32("replication-factor", 3, "Number of replicas.")
	createCmd.Flags().StringSlice("config", nil, "A comma-separated list of topic configuration ('key=value') overrides for the topic being created.")
	createCmd.Flags().Bool("if-not-exists", false, "Exit gracefully if topic already exists.")
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <topic>",
		Short: "Delete a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.onPremDelete),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete the topic `my_topic` at specified cluster (providing Kafka REST Proxy endpoint). Use this command carefully as data loss can occur.",
				Code: "confluent kafka topic delete my_topic --url http://localhost:8082",
			}),
	}
	deleteCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags
	c.AddCommand(deleteCmd)

	updateCmd := &cobra.Command{
		Use:   "update <topic>",
		Short: "Update a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.onPremUpdate),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Modify the `my_topic` topic at specified cluster (providing Kafka REST Proxy endpoint) to have a retention period of 3 days (259200000 milliseconds).",
				Code: "confluent kafka topic update my_topic --url http://localhost:8082 --config=\"retention.ms=259200000\"",
			}),
	}
	updateCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags
	updateCmd.Flags().StringSlice("config", nil, "A comma-separated list of topics configuration ('key=value') overrides for the topic being created.")
	updateCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	c.AddCommand(updateCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.onPremDescribe),
		Short: "Describe a Kafka topic.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the `my_topic` topic at specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "confluent kafka topic describe my_topic --url http://localhost:8082",
			},
		),
	}
	describeCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	c.AddCommand(describeCmd)

	produceCmd := &cobra.Command{
		Use:   "produce <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.onPremProduce),
		Short: "Produce messages to a Kafka topic.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Produce message to topic `my_topic` with SASL_SSL protocol (providing username and password).",
				Code: "confluent kafka topic produce my_topic --url https://localhost:8092/kafka --ca-cert-path ca.crt --protocol SASL_SSL --bootstrap \":19091\" --username user --password secret",
			},
		),
	}
	produceCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	produceCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	produceCmd.Flags().AddFlagSet(pcmd.OnPremAuthenticationSet()) // includes bootstrap, protocol, ssl and sasl credentials
	produceCmd.Flags().String("schema", "", "The path to the local schema file.")
	produceCmd.Flags().String("refs", "", "The path to the references file.")
	produceCmd.Flags().String("tester", "", "The path to the local schema file.")
	produceCmd.Flags().Bool("parse-key", false, "Parse key from the message.")
	produceCmd.Flags().String("delimiter", ":", "The key/value delimiter.")
	produceCmd.Flags().String("value-format", "string", "Format of message value as string, avro, protobuf, or jsonschema.")
	produceCmd.Flags().String("sr-endpoint", "", "url to schema registry cluster.")
	produceCmd.Flags().String("sr-username", "", "Username for connecting to schema registry cluster.")
	produceCmd.Flags().String("sr-password", "", "Password for connecting to schema registry cluster.")
	produceCmd.Flags().SortFlags = false
	c.AddCommand(produceCmd)

	consumeCmd := &cobra.Command{
		Use:   "consume <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.onPremConsume),
		Short: "Consume messages from a Kafka topic.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Consume message from topic `my_topic` with SSL protocol and SSL verification enabled (providing certificate and private key).",
				Code: "confluent kafka topic consume my_topic --url https://localhost:8092/kafka --ca-cert-path ca.crt --protocol SSL --bootstrap \":19091\" --ssl-verification --ca-location ca-cert --cert-location client.pem --key-location client.key",
			},
		),
	}
	consumeCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	consumeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	consumeCmd.Flags().AddFlagSet(pcmd.OnPremAuthenticationSet()) // includes bootstrap, protocol, ssl and sasl credentials
	consumeCmd.Flags().String("group", fmt.Sprintf("confluent_cli_consumer_%s", uuid.New()), "Consumer group ID.")
	consumeCmd.Flags().BoolP("from-beginning", "b", false, "Consume from beginning of the topic.")
	consumeCmd.Flags().String("value-format", "string", "Format of message value as string, avro, protobuf, or jsonschema.")
	consumeCmd.Flags().Bool("print-key", false, "Print key of the message.")
	consumeCmd.Flags().String("delimiter", "\t", "The key/value delimiter.")
	consumeCmd.Flags().String("sr-endpoint", "", "url to schema registry cluster.")
	consumeCmd.Flags().String("sr-username", "", "Username for connecting to schema registry cluster.")
	consumeCmd.Flags().String("sr-password", "", "Password for connecting to schema registry cluster.")

	consumeCmd.Flags().SortFlags = false
	c.AddCommand(consumeCmd)
}

//List Kafka topics.
//
//Usage:
//confluent kafka topic list [flags]
//
//Flags:
//--url string                Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.
//--ca-cert-path string       Path to a PEM-encoded CA to verify the Confluent REST Proxy.
//--client-cert-path string   Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.
//--client-key-path string    Path to client private key, include for mTLS authentication.
//--no-auth                   Include if requests should be made without authentication headers, and user will not be prompted for credentials.
//-o, --output string         Specify the output format as "human", "json", or "yaml". (default "human")
//--context string            CLI Context name.
func (c *authenticatedTopicCommand) onPremList(cmd *cobra.Command, _ []string) error {
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	// Get Topics
	topicGetResp, resp, err := restClient.TopicApi.ClustersClusterIdTopicsGet(restContext, clusterId)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	topicDatas := topicGetResp.Data

	// Create and populate output writer
	outputWriter, err := output.NewListOutputWriter(cmd, []string{"TopicName"}, []string{"Name"}, []string{"name"})
	if err != nil {
		return err
	}
	for _, topicData := range topicDatas {
		outputWriter.AddElement(&topicData)
	}

	return outputWriter.Out()
}

//Create a Kafka topic.
//
//Usage:
//confluent kafka topic create <topic> [flags]
//
//Flags:
//--url string                 Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.
//--ca-cert-path string        Path to a PEM-encoded CA to verify the Confluent REST Proxy.
//--client-cert-path string    Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.
//--client-key-path string     Path to client private key, include for mTLS authentication.
//--no-auth                    Include if requests should be made without authentication headers, and user will not be prompted for credentials.
//--partitions int32           Number of topic partitions. (default 6)
//--replication-factor int32   Number of replicas. (default 3)
//--config strings             A comma-separated list of topic configuration ('key=value') overrides for the topic being created.
//--if-not-exists              Exit gracefully if topic already exists.
//--context string             CLI Context name.
func (c *authenticatedTopicCommand) onPremCreate(cmd *cobra.Command, args []string) error {
	// Parse arguments
	topicName := args[0]
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	// Parse remaining arguments
	numPartitions, err := cmd.Flags().GetInt32("partitions")
	if err != nil {
		return err
	}
	replicationFactor, err := cmd.Flags().GetInt32("replication-factor")
	if err != nil {
		return err
	}
	topicConfigStrings, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}
	ifNotExists, err := cmd.Flags().GetBool("if-not-exists")
	if err != nil {
		return err
	}

	topicConfigsMap, err := utils.ToMap(topicConfigStrings)
	if err != nil {
		return err
	}
	topicConfigs := make([]kafkarestv3.CreateTopicRequestDataConfigs, len(topicConfigsMap))
	i := 0
	for k, v := range topicConfigsMap {
		v2 := v // create a local copy to use pointer
		topicConfigs[i] = kafkarestv3.CreateTopicRequestDataConfigs{
			Name:  k,
			Value: &v2,
		}
		i++
	}
	// Create new topic
	_, resp, err := restClient.TopicApi.ClustersClusterIdTopicsPost(restContext, clusterId, &kafkarestv3.ClustersClusterIdTopicsPostOpts{
		CreateTopicRequestData: optional.NewInterface(kafkarestv3.CreateTopicRequestData{
			TopicName:         topicName,
			PartitionsCount:   numPartitions,
			ReplicationFactor: replicationFactor,
			Configs:           topicConfigs,
		}),
	})
	if err != nil {
		// catch topic exists error
		if openAPIError, ok := err.(kafkarestv3.GenericOpenAPIError); ok {
			var decodedError kafkaRestV3Error
			err2 := json.Unmarshal(openAPIError.Body(), &decodedError)
			if err2 != nil {
				return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
			}
			if decodedError.Message == fmt.Sprintf("Topic '%s' already exists.", topicName) {
				if !ifNotExists {
					return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.TopicExistsOnPremErrorMsg, topicName), errors.TopicExistsOnPremSuggestions)
				} // ignore error if ifNotExists flag is set
				return nil
			}
		}
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	// no error if topic is created successfully.
	utils.Printf(cmd, errors.CreatedTopicMsg, topicName)
	return nil
}

//Delete a Kafka topic.
//
//Usage:
//confluent kafka topic delete <topic> [flags]
//
//Flags:
//--url string                Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.
//--ca-cert-path string       Path to a PEM-encoded CA to verify the Confluent REST Proxy.
//--client-cert-path string   Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.
//--client-key-path string    Path to client private key, include for mTLS authentication.
//--no-auth                   Include if requests should be made without authentication headers, and user will not be prompted for credentials.
//--context string            CLI Context name.
func (c *authenticatedTopicCommand) onPremDelete(cmd *cobra.Command, args []string) error {
	// Parse arguments
	topicName := args[0]
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	// Delete Topic
	resp, err := restClient.TopicApi.ClustersClusterIdTopicsTopicNameDelete(restContext, clusterId, topicName)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	utils.Printf(cmd, errors.DeletedTopicMsg, topicName) // topic successfully deleted
	return nil
}

//Update a Kafka topic.
//
//Usage:
//confluent kafka topic update <topic> [flags]
//
//Flags:
//--url string                Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.
//--ca-cert-path string       Path to a PEM-encoded CA to verify the Confluent REST Proxy.
//--client-cert-path string   Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.
//--client-key-path string    Path to client private key, include for mTLS authentication.
//--no-auth                   Include if requests should be made without authentication headers, and user will not be prompted for credentials.
//--config strings            A comma-separated list of topics configuration ('key=value') overrides for the topic being created.
//--context string            CLI Context name.
func (c *authenticatedTopicCommand) onPremUpdate(cmd *cobra.Command, args []string) error {
	// Parse Argument
	topicName := args[0]
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	} else if !output.IsValidOutputString(format) { // catch format flag
		return output.NewInvalidOutputFormatFlagError(format)
	}
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	// Update Config
	configStrings, err := cmd.Flags().GetStringSlice("config") // handle config parsing errors
	if err != nil {
		return err
	}
	configsMap, err := utils.ToMap(configStrings)
	if err != nil {
		return err
	}
	configs := make([]kafkarestv3.AlterConfigBatchRequestDataData, len(configsMap))
	i := 0
	for k, v := range configsMap {
		v2 := v
		configs[i] = kafkarestv3.AlterConfigBatchRequestDataData{
			Name:      k,
			Value:     &v2,
			Operation: nil,
		}
		i++
	}
	resp, err := restClient.ConfigsApi.ClustersClusterIdTopicsTopicNameConfigsalterPost(restContext, clusterId, topicName,
		&kafkarestv3.ClustersClusterIdTopicsTopicNameConfigsalterPostOpts{
			AlterConfigBatchRequestData: optional.NewInterface(kafkarestv3.AlterConfigBatchRequestData{Data: configs}),
		})
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	if format == output.Human.String() {
		// no errors (config update successful)
		utils.Printf(cmd, errors.UpdateTopicConfigMsg, topicName)
		// Print Updated Configs
		tableLabels := []string{"Name", "Value"}
		tableEntries := make([][]string, len(configs))
		for i, config := range configs {
			tableEntries[i] = printer.ToRow(
				&struct {
					Name  string
					Value string
				}{Name: config.Name, Value: *config.Value}, []string{"Name", "Value"})
		}
		sort.Slice(tableEntries, func(i int, j int) bool {
			return tableEntries[i][0] < tableEntries[j][0]
		})
		printer.RenderCollectionTable(tableEntries, tableLabels)
	} else { //json or yaml
		sort.Slice(configs, func(i int, j int) bool {
			return configs[i].Name < configs[j].Name
		})
		err = output.StructuredOutput(format, configs)
		if err != nil {
			return err
		}
	}
	return nil
}

//Describe a Kafka topic.
//
//Usage:
//confluent kafka topic describe <topic> [flags]
//
//Flags:
//--url string                Base URL of REST Proxy Endpoint of Kafka Cluster (include /kafka for embedded Rest Proxy). Must set flag or CONFLUENT_REST_URL.
//--ca-cert-path string       Path to a PEM-encoded CA to verify the Confluent REST Proxy.
//--client-cert-path string   Path to client cert to be verified by Confluent REST Proxy, include for mTLS authentication.
//--client-key-path string    Path to client private key, include for mTLS authentication.
//--no-auth                   Include if requests should be made without authentication headers, and user will not be prompted for credentials.
//-o, --output string         Specify the output format as "human", "json", or "yaml". (default "human")
//--context string            CLI Context name.
func (c *authenticatedTopicCommand) onPremDescribe(cmd *cobra.Command, args []string) error {
	// Parse Args
	topicName := args[0]
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	} else if !output.IsValidOutputString(format) { // catch format flag
		return output.NewInvalidOutputFormatFlagError(format)
	}
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	// Get partitions
	topicData := &TopicData{}
	partitionsResp, resp, err := restClient.PartitionApi.ClustersClusterIdTopicsTopicNamePartitionsGet(restContext, clusterId, topicName)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	} else if partitionsResp.Data == nil {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
	}
	topicData.TopicName = topicName
	topicData.PartitionCount = len(partitionsResp.Data)
	topicData.Partitions = make([]PartitionData, len(partitionsResp.Data))
	for i, partitionResp := range partitionsResp.Data {
		partitionId := partitionResp.PartitionId
		partitionData := PartitionData{
			TopicName:   topicName,
			PartitionId: partitionId,
		}

		// For each partition, get replicas
		replicasResp, resp, err := restClient.ReplicaApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicasGet(restContext, clusterId, topicName, partitionId)
		if err != nil {
			return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
		} else if replicasResp.Data == nil {
			return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
		}
		partitionData.ReplicaBrokerIds = make([]int32, len(replicasResp.Data))
		partitionData.InSyncReplicaBrokerIds = make([]int32, 0, len(replicasResp.Data))
		for j, replicaResp := range replicasResp.Data {
			if replicaResp.IsLeader {
				partitionData.LeaderBrokerId = replicaResp.BrokerId
			}
			partitionData.ReplicaBrokerIds[j] = replicaResp.BrokerId
			if replicaResp.IsInSync {
				partitionData.InSyncReplicaBrokerIds = append(partitionData.InSyncReplicaBrokerIds, replicaResp.BrokerId)
			}
		}
		if i == 0 {
			topicData.ReplicationFactor = len(replicasResp.Data)
		}
		topicData.Partitions[i] = partitionData
	}

	// Get configs
	configsResp, resp, err := restClient.ConfigsApi.ClustersClusterIdTopicsTopicNameConfigsGet(restContext, clusterId, topicName)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	} else if configsResp.Data == nil {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
	}
	topicData.Configs = make(map[string]string)
	for _, config := range configsResp.Data {
		if config.Value != nil {
			topicData.Configs[config.Name] = *config.Value
		} else {
			topicData.Configs[config.Name] = ""
		}
	}
	// Print topic info
	if format == output.Human.String() { // human output
		// Output partitions info
		utils.Printf(cmd, "Topic: %s\nPartitionCount: %d\nReplicationFactor: %d\n\n", topicData.TopicName, topicData.PartitionCount, topicData.ReplicationFactor)
		partitionsTableLabels := []string{"Topic", "Partition", "Leader", "Replicas", "ISR"}
		partitionsTableEntries := make([][]string, topicData.PartitionCount)
		for i, partition := range topicData.Partitions {
			partitionsTableEntries[i] = printer.ToRow(&partition, []string{"TopicName", "PartitionId", "LeaderBrokerId", "ReplicaBrokerIds", "InSyncReplicaBrokerIds"})
		}
		printer.RenderCollectionTable(partitionsTableEntries, partitionsTableLabels)
		// Output config info
		utils.Print(cmd, "\nConfiguration\n\n")
		configsTableLabels := []string{"Name", "Value"}
		configsTableEntries := make([][]string, len(topicData.Configs))
		i := 0
		for name, value := range topicData.Configs {
			configsTableEntries[i] = printer.ToRow(&struct {
				name  string
				value string
			}{name: name, value: value}, []string{"name", "value"})
			i++
		}
		sort.Slice(configsTableEntries, func(i int, j int) bool {
			return configsTableEntries[i][0] < configsTableEntries[j][0]
		})
		printer.RenderCollectionTable(configsTableEntries, configsTableLabels)
	} else { // machine output (json or yaml)
		err = output.StructuredOutput(format, topicData)
		if err != nil {
			return err
		}
	}
	return nil
}

func getClusterIdForRestRequests(client *kafkarestv3.APIClient, ctx context.Context) (string, error) {
	clusters, resp, err := client.ClusterApi.ClustersGet(ctx)
	if err != nil {
		return "", kafkaRestError(client.GetConfig().BasePath, err, resp)
	}
	if clusters.Data == nil || len(clusters.Data) == 0 {
		return "", errors.NewErrorWithSuggestions(errors.NoClustersFoundErrorMsg, errors.NoClustersFoundSuggestions)
	}
	clusterId := clusters.Data[0].ClusterId
	return clusterId, nil
}

func (c *authenticatedTopicCommand) onPremProduce(cmd *cobra.Command, args []string) error {
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	topicName := args[0]

	bootstrap, err := cmd.Flags().GetString("bootstrap")
	if err != nil {
		return err
	}

	protocol, err := cmd.Flags().GetString("protocol")
	if err != nil {
		return err
	}
	enableSSLVerification, err := cmd.Flags().GetBool("ssl-verification")
	if err != nil {
		return err
	}
	caLocation, err := cmd.Flags().GetString("ca-location")
	if err != nil {
		return err
	}

	configMap := GetOnPremProducerCommonConfig(c.clientID, bootstrap, enableSSLVerification, caLocation)
	switch protocol {
	case "SSL":
		certLocation, err := cmd.Flags().GetString("cert-location")
		if err != nil {
			return err
		}
		keyLocation, err := cmd.Flags().GetString("key-location")
		if err != nil {
			return err
		}
		keyPassword, err := cmd.Flags().GetString("key-password")
		if err != nil {
			return err
		}
		configMap, err = SetSSLConfig(configMap, certLocation, keyLocation, keyPassword)
		if err != nil {
			return err
		}
	case "SASL_SSL":
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			return err
		}
		password, err := cmd.Flags().GetString("password")
		if err != nil {
			return err
		}
		configMap, err = SetSASLConfig(configMap, username, password)
		if err != nil {
			return err
		}
	}

	producer, err := ckafka.NewProducer(configMap)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateProducerMsg, err)
	}
	defer producer.Close()
	c.logger.Tracef("Create producer succeeded")

	adminClient, err := ckafka.NewAdminClientFromProducer(producer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientMsg, err)
	}
	defer adminClient.Close()

	err = c.validateTopic(adminClient, topicName, clusterId)
	if err != nil {
		return err
	}

	parseKey, err := cmd.Flags().GetBool("parse-key")
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

	subject := topicName + "-value"
	serializationProvider, err := serdes.GetSerializationProvider(valueFormat)
	if err != nil {
		return err
	}
	// Meta info contains magic byte and schema ID (4 bytes).
	metaInfo, _, err := c.registerSchema(cmd, valueFormat, schemaPath, subject, serializationProvider.GetSchemaName(), nil)
	if err != nil {
		return err
	}
	err = serializationProvider.LoadSchema(schemaPath, nil)
	if err != nil {
		return err
	}

	utils.ErrPrintln(cmd, errors.StartingProducerMsg)

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

	deliveryChan := make(chan ckafka.Event)
	for data := range input {
		if len(data) == 0 {
			go scan()
			continue
		}

		key, value, err := getMsgKeyAndValueOnPrem(metaInfo, data, delim, parseKey, serializationProvider)
		if err != nil {
			return err
		}

		msg := &ckafka.Message{
			TopicPartition: ckafka.TopicPartition{Topic: &topicName, Partition: ckafka.PartitionAny},
			Key:            []byte(key),
			Value:          []byte(value),
		}
		err = producer.Produce(msg, deliveryChan)
		if err != nil {
			utils.ErrPrintf(cmd, errors.FailedToProduceErrorMsg, msg.TopicPartition.Offset, err)
		}

		e := <-deliveryChan                // read a ckafka event from the channel
		m := e.(*ckafka.Message)           // extract the message from the event
		if m.TopicPartition.Error != nil { // catch all other errors
			isProduceToCompactedTopicError, err := errors.CatchProduceToCompactedTopicError(err, topicName)
			if isProduceToCompactedTopicError {
				scanErr = err
				close(input)
				break
			}
			utils.ErrPrintf(cmd, errors.FailedToProduceErrorMsg, m.TopicPartition.Offset, m.TopicPartition.Error)
		}
		go scan()
	}
	close(deliveryChan)
	return scanErr
}

func (c *authenticatedTopicCommand) onPremConsume(cmd *cobra.Command, args []string) error {
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	topicName := args[0]

	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	beginning, err := cmd.Flags().GetBool("from-beginning")
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

	valueFormat, err := cmd.Flags().GetString("value-format")
	if err != nil {
		return err
	}

	bootstrap, err := cmd.Flags().GetString("bootstrap")
	if err != nil {
		return err
	}

	protocol, err := cmd.Flags().GetString("protocol")
	if err != nil {
		return err
	}

	enableSSLVerification, err := cmd.Flags().GetBool("ssl-verification")
	if err != nil {
		return err
	}
	caLocation, err := cmd.Flags().GetString("ca-location")
	if err != nil {
		return err
	}

	configMap, err := GetOnPremConsumerCommonConfig(c.clientID, bootstrap, group, beginning, enableSSLVerification, caLocation)
	if err != nil {
		return err
	}
	switch protocol {
	case "SSL":
		certLocation, err := cmd.Flags().GetString("cert-location")
		if err != nil {
			return err
		}
		keyLocation, err := cmd.Flags().GetString("key-location")
		if err != nil {
			return err
		}
		keyPassword, err := cmd.Flags().GetString("key-password")
		if err != nil {
			return err
		}
		configMap, err = SetSSLConfig(configMap, certLocation, keyLocation, keyPassword)
		if err != nil {
			return err
		}
	case "SASL_SSL":
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			return err
		}
		password, err := cmd.Flags().GetString("password")
		if err != nil {
			return err
		}
		configMap, err = SetSASLConfig(configMap, username, password)
		if err != nil {
			return err
		}
	}

	consumer, err := ckafka.NewConsumer(configMap)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateConsumerMsg, err)
	}
	c.logger.Tracef("Create consumer succeeded")

	adminClient, err := ckafka.NewAdminClientFromConsumer(consumer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientMsg, err)
	}
	defer adminClient.Close()

	err = c.validateTopic(adminClient, topicName, clusterId)
	if err != nil {
		return err
	}

	utils.ErrPrintln(cmd, errors.StartingConsumerMsg)

	dir := filepath.Join(os.TempDir(), "ccloud-schema")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}
	}

	err = consumer.Subscribe(topicName, nil)
	if err != nil {
		return err
	}

	groupHandler := &GroupHandler{
		Format:     valueFormat,
		Out:        cmd.OutOrStdout(),
		Properties: ConsumerProperties{PrintKey: printKey, Delimiter: delimiter, SchemaPath: dir, Cloud: false},
	}

	// start consuming messages
	err = RunConsumer(cmd, consumer, groupHandler)
	return err
}

// validate that a topic exists before attempting to produce/consume messages
func (c *authenticatedTopicCommand) validateTopic(adminClient *ckafka.AdminClient, topic, clusterId string) error {
	timeout := 10 * time.Second
	metadata, err := adminClient.GetMetadata(nil, true, int(timeout.Milliseconds()))
	if err != nil {
		return fmt.Errorf("failed to obtain topics from client: %v", err)
	}

	var foundTopic bool
	for _, t := range metadata.Topics {
		c.logger.Tracef("validateTopic: found topic " + t.Topic)
		if topic == t.Topic {
			foundTopic = true // no break so that we see all topics from the above printout
		}
	}
	if !foundTopic {
		c.logger.Tracef("validateTopic failed due to topic not being found in the client's topic list")
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.TopicDoesNotExistOrMissingACLsErrorMsg, topic), fmt.Sprintf(errors.TopicDoesNotExistOrMissingACLsSuggestions, clusterId, clusterId, clusterId))
	}

	c.logger.Tracef("validateTopic succeeded")
	return nil
}

func getMsgKeyAndValueOnPrem(metaInfo []byte, data, delim string, parseKey bool, serializationProvider serdes.SerializationProvider) (string, string, error) {
	var key, valueString string
	if parseKey {
		record := strings.SplitN(data, delim, 2)
		valueString = strings.TrimSpace(record[len(record)-1])

		if len(record) == 2 {
			key = strings.TrimSpace(record[0])
		} else {
			return "", "", errors.New(errors.MissingKeyErrorMsg)
		}
	} else {
		valueString = strings.TrimSpace(data)
	}
	encodedMessage, err := serdes.Serialize(serializationProvider, valueString)
	if err != nil {
		return "", "", err
	}
	encoded := append(metaInfo, encodedMessage...)
	value := string(encoded)
	return key, value, nil
}

func (c *authenticatedTopicCommand) registerSchema(cmd *cobra.Command, valueFormat, schemaPath, subject, schemaType string, refs []srsdk.SchemaReference) ([]byte, map[string]string, error) {
	// For plain string encoding, meta info is empty.
	// Registering schema when specified, and fill metaInfo array.
	metaInfo := []byte{}
	referencePathMap := map[string]string{}
	if valueFormat != "string" && len(schemaPath) > 0 {
		srUsername, err := cmd.Flags().GetString("sr-username")
		if err != nil {
			return metaInfo, nil, err
		}
		srPassword, err := cmd.Flags().GetString("sr-password")
		if err != nil {
			return metaInfo, nil, err
		}
		info, err := c.registerSchemaWithUserInfo(cmd, subject, schemaType, schemaPath, srUsername, srPassword)
		if err != nil {
			return metaInfo, nil, err
		}
		metaInfo = info
		referencePathMap, err = c.storeSchemaReferencesOnPrem(cmd, refs)
		if err != nil {
			return metaInfo, nil, err
		}
	}
	return metaInfo, referencePathMap, nil
}

func (c *authenticatedTopicCommand) storeSchemaReferencesOnPrem(cmd *cobra.Command, refs []srsdk.SchemaReference) (map[string]string, error) {
	dir := filepath.Join(os.TempDir(), "ccloud-schema")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return nil, err
		}
	}

	referencePathMap := map[string]string{}
	for _, ref := range refs {
		tempStorePath := filepath.Join(dir, ref.Name)
		if !fileExists(tempStorePath) {
			err := getAndWriteSchemaByVersion(cmd, tempStorePath, ref.Subject, strconv.Itoa(int(ref.Version)))
			if err != nil {
				return nil, err
			}
		}
		referencePathMap[ref.Name] = tempStorePath
	}

	return referencePathMap, nil
}

func getAndWriteSchemaByVersion(cmd *cobra.Command, tempStorePath, subject, version string) error {
	// retrieve schema by subject and version, write to temp path
	schema := SchemaObject{}

	srEndpoint, err := cmd.Flags().GetString("sr-endpoint")
	if err != nil {
		return err
	}
	requestUrl := srEndpoint + "/subjects/" + subject + "/versions/" + version

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth("superUser", "superUser")

	caCertPath, err := cmd.Flags().GetString("ca-cert-path")
	if err != nil {
		return err
	}
	client := GetCAClient(caCertPath)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&schema)
	if err != nil {
		return err
	}

	file, _ := json.MarshalIndent(schema, "", " ")
	err = ioutil.WriteFile(tempStorePath, file, 0644)
	return err
}

func (c *authenticatedTopicCommand) registerSchemaWithUserInfo(cmd *cobra.Command, subject, valueFormat, schemaPath, srUsername, srPassword string) ([]byte, error) {
	srEndpoint, err := cmd.Flags().GetString("sr-endpoint")
	if err != nil {
		return nil, err
	}
	requestUrl := srEndpoint + "/subjects/" + subject + "/versions"

	// load schema into memory
	schemaBytes, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}
	schema := SchemaObject{}
	json.Unmarshal([]byte(schemaBytes), &schema)

	// convert marshalled json object into json request string
	requestString, err := ConvertSchemaToRequestString(schema, valueFormat)
	if err != nil {
		return nil, err
	}
	requestReader := strings.NewReader(requestString)

	req, err := http.NewRequest("POST", requestUrl, requestReader)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(srUsername, srPassword)
	req.Header.Set("Content-Type", "application/vnd.schemaregistry.v1+json")

	caCertPath, err := cmd.Flags().GetString("ca-cert-path")
	if err != nil {
		return nil, err
	}
	client := GetCAClient(caCertPath)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Println("response:", resp)
	schemaId, err := extractSchemaId(resp)
	if err != nil {
		return nil, err
	}

	metaInfo := []byte{0x0}
	schemaIdBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(schemaIdBuffer, schemaId)
	metaInfo = append(metaInfo, schemaIdBuffer...)
	return metaInfo, nil
}

func extractSchemaId(response *http.Response) (uint32, error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	responseBody := buf.String() // {"id":9}
	schemaId, err := strconv.ParseInt(responseBody[6:len(responseBody)-1], 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(schemaId), nil
}
