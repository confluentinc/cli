package unifiedstreammanager

import (
	"fmt"

	"github.com/spf13/cobra"

	usmv1 "github.com/confluentinc/ccloud-sdk-go-v2/usm/v1"

	"github.com/confluentinc/cli/v4/pkg/output"
)

const kafkaClusterNotFoundErrorMsg = "USM Kafka cluster corresponding to Confluent Platform Kafka cluster %s not found"

type connectOut struct {
	Id                              string `human:"ID" serialized:"id"`
	ConfluentPlatformConnectCluster string `human:"Confluent Platform Connect Cluster" serialized:"confluent_platform_connect_cluster"`
	USMKafkaClusterId               string `human:"USM Kafka Cluster ID" serialized:"usm_kafka_cluster_id"`
	ConfluentPlatformKafkaClusterId string `human:"Confluent Platform Kafka Cluster ID" serialized:"confluent_platform_kafka_cluster_id"`
	Cloud                           string `human:"Cloud" serialized:"cloud"`
	Region                          string `human:"Region" serialized:"region"`
	Environment                     string `human:"Environment" serialized:"environment"`
}

func (c *command) newConnectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Manage USM Connect clusters.",
	}

	cmd.AddCommand(c.newConnectDeregisterCommand())
	cmd.AddCommand(c.newConnectDescribeCommand())
	cmd.AddCommand(c.newConnectListCommand())
	cmd.AddCommand(c.newConnectRegisterCommand())

	return cmd
}

func (c *command) validConnectArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validConnectArgsMultiple(cmd, args)
}

func (c *command) validConnectArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteConnectClusters()
}

func (c *command) autocompleteConnectClusters() []string {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	clusters, err := c.V2Client.ListUsmConnectClusters(environmentId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters))
	for i, cluster := range clusters {
		suggestions[i] = fmt.Sprintf("%s\t%s", cluster.GetId(), cluster.GetConfluentPlatformConnectClusterId())
	}
	return suggestions
}

func printConnectTable(cmd *cobra.Command, connect usmv1.UsmV1ConnectCluster, usmKafkaClusterId string) error {
	out := &connectOut{
		Id:                              connect.GetId(),
		ConfluentPlatformConnectCluster: connect.GetConfluentPlatformConnectClusterId(),
		USMKafkaClusterId:               usmKafkaClusterId,
		ConfluentPlatformKafkaClusterId: connect.GetKafkaClusterId(),
		Cloud:                           connect.GetCloud(),
		Region:                          connect.GetRegion(),
		Environment:                     connect.Environment.GetId(),
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}

func (c *command) getOnPremToCloudKafkaIdMap(environment string) (map[string]string, error) {
	clusters, err := c.V2Client.ListUsmKafkaClusters(environment)
	if err != nil {
		return nil, err
	}

	idMap := make(map[string]string, len(clusters))
	for _, cluster := range clusters {
		idMap[cluster.GetConfluentPlatformKafkaClusterId()] = cluster.GetId()
	}

	return idMap, nil
}
