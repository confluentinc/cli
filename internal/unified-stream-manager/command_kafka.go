package unifiedstreammanager

import (
	"fmt"

	"github.com/spf13/cobra"

	usmv1 "github.com/confluentinc/ccloud-sdk-go-v2/usm/v1"

	"github.com/confluentinc/cli/v4/pkg/output"
)

type kafkaOut struct {
	Id                              string `human:"ID" serialized:"id"`
	Name                            string `human:"Name" serialized:"name"`
	ConfluentPlatformKafkaClusterId string `human:"Confluent Platform Kafka Cluster ID" serialized:"confluent_platform_kafka_cluster_id"`
	Cloud                           string `human:"Cloud" serialized:"cloud"`
	Region                          string `human:"Region" serialized:"region"`
	Environment                     string `human:"Environment" serialized:"environment"`
}

func (c *command) newKafkaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kafka",
		Short: "Manage USM Kafka clusters.",
	}

	cmd.AddCommand(c.newKafkaDeregisterCommand())
	cmd.AddCommand(c.newKafkaDescribeCommand())
	cmd.AddCommand(c.newKafkaListCommand())
	cmd.AddCommand(c.newKafkaRegisterCommand())

	return cmd
}

func (c *command) validKafkaArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validKafkaArgsMultiple(cmd, args)
}

func (c *command) validKafkaArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteKafkaClusters()
}

func (c *command) autocompleteKafkaClusters() []string {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	clusters, err := c.V2Client.ListUsmKafkaClusters(environmentId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters))
	for i, cluster := range clusters {
		suggestions[i] = fmt.Sprintf("%s\t%s", cluster.GetId(), cluster.GetDisplayName())
	}
	return suggestions
}

func printKafkaTable(cmd *cobra.Command, kafka usmv1.UsmV1KafkaCluster) error {
	out := &kafkaOut{
		Id:                              kafka.GetId(),
		Name:                            kafka.GetDisplayName(),
		ConfluentPlatformKafkaClusterId: kafka.GetConfluentPlatformKafkaClusterId(),
		Cloud:                           kafka.GetCloud(),
		Region:                          kafka.GetRegion(),
		Environment:                     kafka.Environment.GetId(),
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
