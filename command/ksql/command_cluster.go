package ksql

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	ksqlv1 "github.com/confluentinc/ccloudapis/ksql/v1"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/go-printer"
)

var (
	listFields      = []string{"Id", "Name", "OutputTopicPrefix", "KafkaClusterId", "Storage", "Endpoint", "Status"}
	listLabels      = []string{"Id", "Name", "Topic Prefix", "Kafka", "Storage", "Endpoint", "Status"}
	describeFields  = []string{"Id", "Name", "OutputTopicPrefix", "KafkaClusterId", "Storage", "Endpoint", "Status"}
	describeRenames = map[string]string{"KafkaClusterId": "Kafka", "OutputTopicPrefix": "Topic Prefix"}
)

type clusterCommand struct {
	*cobra.Command
	config *shared.Config
	client chttp.KSQL
}

// NewClusterCommand returns the Cobra clusterCommand for Ksql Cluster.
func NewClusterCommand(config *shared.Config, plugin common.GRPCPlugin) *cobra.Command {
	cmd := &clusterCommand{
		Command: &cobra.Command{
			Use:   "app",
			Short: "Manage KSQL apps",
		},
		config: config,
	}
	cmd.init(plugin)
	return cmd.Command
}

func (c *clusterCommand) init(plugin common.GRPCPlugin) {

	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := common.SetLoggingVerbosity(cmd, c.config.Logger); err != nil {
			return common.HandleError(err, cmd)
		}
		if err := c.config.CheckLogin(); err != nil {
			return common.HandleError(err, cmd)
		}
		// Lazy load plugin to avoid unnecessarily spawning child processes
		return plugin.Load(&c.client, c.config.Logger)
	}

	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List KSQL apps",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})

	createCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a KSQL app",
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
		Short: "Describe a ksql app",
		RunE:  c.describe,
		Args:  cobra.ExactArgs(1),
	})
	c.AddCommand(&cobra.Command{
		Use:   "delete ID",
		Short: "Delete a ksql app",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	})
}

func (c *clusterCommand) list(cmd *cobra.Command, args []string) error {
	req := &ksqlv1.KSQLCluster{AccountId: c.config.Auth.Account.Id}
	clusters, err := c.client.List(context.Background(), req)
	if err != nil {
		return common.HandleError(err, cmd)
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
		return common.HandleError(err, cmd)
	}
	storage, err := cmd.Flags().GetInt32("storage")
	if err != nil {
		return common.HandleError(err, cmd)
	}
	servers, err := cmd.Flags().GetInt32("servers")
	if err != nil {
		return common.HandleError(err, cmd)
	}
	config := &ksqlv1.KSQLClusterConfig{
		AccountId:      c.config.Auth.Account.Id,
		Name:           args[0],
		Servers:        servers,
		Storage:        storage,
		KafkaClusterId: kafkaClusterID,
	}
	cluster, err := c.client.Create(context.Background(), config)
	if err != nil {
		return common.HandleError(err, cmd)
	}
	return printer.RenderTableOut(cluster, describeFields, describeRenames, os.Stdout)
}

func (c *clusterCommand) describe(cmd *cobra.Command, args []string) error {
	req := &ksqlv1.KSQLCluster{AccountId: c.config.Auth.Account.Id, Id: args[0]}
	cluster, err := c.client.Describe(context.Background(), req)
	if err != nil {
		return common.HandleError(err, cmd)
	}
	return printer.RenderTableOut(cluster, describeFields, describeRenames, os.Stdout)
}

func (c *clusterCommand) delete(cmd *cobra.Command, args []string) error {
	req := &ksqlv1.KSQLCluster{AccountId: c.config.Auth.Account.Id, Id: args[0]}
	err := c.client.Delete(context.Background(), req)
	if err != nil {
		return common.HandleError(err, cmd)
	}
	fmt.Printf("The ksql app %s has been deleted.\n", args[0])
	return nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
