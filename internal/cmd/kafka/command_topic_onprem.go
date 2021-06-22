package kafka

// confluent kafka topic <commands>
import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/go-printer"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	kafka "github.com/confluentinc/cli/internal/pkg/kafka"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

// Info needed to complete kafka topic ...
type topicCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner pcmd.PreRunner
}

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

// Return the command to be registered to the kafka topic slot
func NewTopicCommandOnPrem(prerunner pcmd.PreRunner) *cobra.Command {
	topicCmd := &topicCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(
			&cobra.Command{
				Use:   "topic",
				Short: "Manage Kafka topics.",
			}, prerunner, OnPremTopicSubcommandFlags),
		prerunner: prerunner,
	}
	topicCmd.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(topicCmd.AuthenticatedCLICommand))
	topicCmd.init()
	return topicCmd.Command
}

// Register each of the verbs and expected args
func (topicCmd *topicCommand) init() {
	// Register list command
	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(topicCmd.listTopics),
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
	listCmd.Flags().SortFlags = false
	topicCmd.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create <topic>",
		Short: "Create a Kafka topic.",
		Args:  cobra.ExactArgs(1), // <topic>
		RunE:  pcmd.NewCLIRunE(topicCmd.createTopic),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a topic named ``my_topic`` with default options at specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "confluent kafka topic create my_topic --url http://localhost:8082",
			},
			examples.Example{
				Text: "Create a topic named ``my_topic_2`` with specified configuration parameters.",
				Code: "confluent kafka topic create my_topic_2 --url http://localhost:8082 --config cleanup.policy=compact,compression.type=gzip",
			}),
	}
	createCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags
	createCmd.Flags().Int32("partitions", 6, "Number of topic partitions.")
	createCmd.Flags().Int32("replication-factor", 3, "Number of replicas.")
	createCmd.Flags().StringSlice("config", nil, "A comma-separated list of topic configuration ('key=value') overrides for the topic being created.")
	createCmd.Flags().Bool("if-not-exists", false, "Exit gracefully if topic already exists.")
	createCmd.Flags().SortFlags = false
	topicCmd.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <topic>",
		Short: "Delete a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(topicCmd.deleteTopic),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete the topic ``my_topic`` at specified cluster (providing Kafka REST Proxy endpoint). Use this command carefully as data loss can occur.",
				Code: "confluent kafka topic delete my_topic --url http://localhost:8082",
			}),
	}
	deleteCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags
	deleteCmd.Flags().SortFlags = false
	topicCmd.AddCommand(deleteCmd)

	updateCmd := &cobra.Command{
		Use:   "update <topic>",
		Short: "Update a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(topicCmd.updateTopicConfig),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Modify the ``my_topic`` topic at specified cluster (providing Kafka REST Proxy endpoint) to have a retention period of 3 days (259200000 milliseconds).",
				Code: "confluent kafka topic update my_topic --url http://localhost:8082 --config=\"retention.ms=259200000\"",
			}),
	}
	updateCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags
	updateCmd.Flags().StringSlice("config", nil, "A comma-separated list of topics configuration ('key=value') overrides for the topic being created.")
	updateCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	updateCmd.Flags().SortFlags = false
	topicCmd.AddCommand(updateCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(topicCmd.describeTopic),
		Short: "Describe a Kafka topic.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the ``my_topic`` topic at specified cluster (providing Kafka REST Proxy endpoint).",
				Code: "confluent kafka topic describe my_topic --url http://localhost:8082",
			},
		),
	}
	describeCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet()) //includes url, ca-cert-path, client-cert-path, client-key-path, and no-auth flags
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	topicCmd.AddCommand(describeCmd)
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
func (topicCmd *topicCommand) listTopics(cmd *cobra.Command, args []string) error {
	restClient, restContext, err := initKafkaRest(topicCmd.AuthenticatedCLICommand, cmd)
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
func (topicCmd *topicCommand) createTopic(cmd *cobra.Command, args []string) error {
	// Parse arguments
	topicName := args[0]
	restClient, restContext, err := initKafkaRest(topicCmd.AuthenticatedCLICommand, cmd)
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

	topicConfigsMap, err := kafka.ToMap(topicConfigStrings)
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
func (topicCmd *topicCommand) deleteTopic(cmd *cobra.Command, args []string) error {
	// Parse arguments
	topicName := args[0]
	restClient, restContext, err := initKafkaRest(topicCmd.AuthenticatedCLICommand, cmd)
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
func (topicCmd *topicCommand) updateTopicConfig(cmd *cobra.Command, args []string) error {
	// Parse Argument
	topicName := args[0]
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	} else if !output.IsValidFormatString(format) { // catch format flag
		return output.NewInvalidOutputFormatFlagError(format)
	}
	restClient, restContext, err := initKafkaRest(topicCmd.AuthenticatedCLICommand, cmd)
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
	configsMap, err := kafka.ToMap(configStrings)
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
func (topicCmd *topicCommand) describeTopic(cmd *cobra.Command, args []string) error {
	// Parse Args
	topicName := args[0]
	format, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	} else if !output.IsValidFormatString(format) { // catch format flag
		return output.NewInvalidOutputFormatFlagError(format)
	}
	restClient, restContext, err := initKafkaRest(topicCmd.AuthenticatedCLICommand, cmd)
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
