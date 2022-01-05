package kafka

import (
	"sort"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"
)

func (c *authenticatedTopicCommand) newDescribeCommandOnPrem() *cobra.Command {
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
	pcmd.AddOutputFlag(describeCmd)

	return describeCmd
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
