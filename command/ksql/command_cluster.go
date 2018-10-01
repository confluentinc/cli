package ksql

import (
	"context"
	"fmt"
	"os"

	"github.com/codyaray/go-printer"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

var (
	listFields      = []string{"Id", "Name", "KafkaClusterId", "Storage", "Servers", "Region", "Status"}
	listLabels      = []string{"Id", "Name", "Kafka", "Storage", "Servers", "Region", "Status"}
	describeFields  = []string{"Id", "Name", "KafkaClusterId", "Storage", "Servers", "Region", "Status"}
	describeRenames = map[string]string{"KafkaClusterId": "Kafka"}
)

type clusterCommand struct {
	*cobra.Command
	config *shared.Config
	ksql   Ksql
}

// NewClusterCommand returns the Cobra clusterCommand for Ksql Cluster.
func NewClusterCommand(config *shared.Config, ksql Ksql) *cobra.Command {
	cmd := &clusterCommand{
		Command: &cobra.Command{
			Use:   "app",
			Short: "Manage ksql apps.",
		},
		config: config,
		ksql:   ksql,
	}
	cmd.init()
	return cmd.Command
}

func (c *clusterCommand) init() {
	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List ksql apps.",
		RunE:  c.list,
	})

	createCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a ksql app.",
		RunE:  c.create,
		Args:  cobra.ExactArgs(1),
	}
	createCmd.Flags().String("kafka-cluster", "", "Kafka Cluster ID")
	check(createCmd.MarkFlagRequired("kafka-cluster"))
	createCmd.Flags().Int32("storage", 50, "total usable data storage in GB")
	check(createCmd.MarkFlagRequired("storage"))
	createCmd.Flags().Int32("servers", 1, "number of servers in the cluster")

	c.AddCommand(createCmd)

	c.AddCommand(&cobra.Command{
		Use:   "describe ID",
		Short: "Describe a ksql app.",
		RunE:  c.describe,
		Args:  cobra.ExactArgs(1),
	})
	c.AddCommand(&cobra.Command{
		Use:   "delete ID",
		Short: "Delete a ksql app.",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	})
}

func (c *clusterCommand) list(cmd *cobra.Command, args []string) error {
	req := &schedv1.KSQLCluster{AccountId: c.config.Auth.Account.Id}
	clusters, err := c.ksql.List(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	var data [][]string
	for _, cluster := range clusters {
		data = append(data, printer.ToRow(cluster, listFields))
	}
	printer.RenderCollectionTable(data, listLabels)
	return nil
}

func (c *clusterCommand) create(cmd *cobra.Command, args []string) error {
	kafkaClusterID, err := cmd.Flags().GetString("kafka-cluster")
	if err != nil {
		return common.HandleError(err)
	}
	storage, err := cmd.Flags().GetInt32("storage")
	if err != nil {
		return common.HandleError(err)
	}
	servers, err := cmd.Flags().GetInt32("servers")
	if err != nil {
		return common.HandleError(err)
	}
	config := &schedv1.KSQLClusterConfig{
		AccountId:      c.config.Auth.Account.Id,
		Name:           args[0],
		Servers:        servers,
		Storage:        storage,
		KafkaClusterId: kafkaClusterID,
	}
	cluster, err := c.ksql.Create(context.Background(), config)
	if err != nil {
		return common.HandleError(err)
	}
	printer.RenderTableOut(cluster, describeFields, describeRenames, os.Stdout)
	return nil
}

func (c *clusterCommand) describe(cmd *cobra.Command, args []string) error {
	req := &schedv1.KSQLCluster{AccountId: c.config.Auth.Account.Id, Id: args[0]}
	cluster, err := c.ksql.Describe(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	printer.RenderTableOut(cluster, describeFields, describeRenames, os.Stdout)
	return nil
}

func (c *clusterCommand) delete(cmd *cobra.Command, args []string) error {
	req := &schedv1.KSQLCluster{AccountId: c.config.Auth.Account.Id, Id: args[0]}
	err := c.ksql.Delete(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	fmt.Printf("The ksql app %s has been deleted.\n", args[0])
	return nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
