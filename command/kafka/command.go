package kafka

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

var (
	listFields     = []string{"Name", "ServiceProvider", "Region", "Durability", "Status"}
	listLabels     = []string{"Name", "Provider", "Region", "Durability", "Status"}
	describeFields = []string{"Id", "Name", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Region", "Status", "Endpoint", "PricePerHour"}
	describeLabels = []string{"Id", "Name", "Ingress", "Egress", "Storage", "Provider", "Region", "Status", "Endpoint", "PricePerHour"}
)
type Command struct {
	*cobra.Command
	config *shared.Config
	kafka  Kafka
}

func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "kafka",
			Short: "Manage kafka clusters.",
		},
		config: config,
	}
	err := cmd.init()
	return cmd.Command, err
}

func (c *Command) init() error {
	path, err := exec.LookPath("confluent-kafka-plugin")
	if err != nil {
		return fmt.Errorf("skipping kafka: plugin isn't installed")
	}

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Plugins:          shared.PluginMap,
		Cmd:              exec.Command("sh", "-c", path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Managed:          true,
		Logger: hclog.New(&hclog.LoggerOptions{
			Output: hclog.DefaultOutput,
			Level:  hclog.Info,
			Name:   "plugin",
		}),
	})

	// Connect via RPC.
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("kafka")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Got a client now communicating over RPC.
	c.kafka = raw.(Kafka)

	// All commands require login first
	c.Command.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if err := c.config.CheckLogin(); err != nil {
			common.HandleError(err)
			os.Exit(0) // TODO: this should be 1 but that prints "exit status 1" to the console
		}
	}

	c.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create a Kafka cluster.",
		RunE:  c.create,
	})
	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List Kafka clusters.",
		RunE:  c.list,
	})
	c.AddCommand(&cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Kafka cluster.",
		RunE:  c.describe,
		Args:  cobra.ExactArgs(1),
	})
	c.AddCommand(&cobra.Command{
		Use:   "delete",
		Short: "Delete a Kafka cluster.",
		RunE:  c.delete,
	})
	c.AddCommand(&cobra.Command{
		Use:   "update",
		Short: "Update a Kafka cluster.",
		RunE:  c.update,
	})
	c.AddCommand(&cobra.Command{
		Use:   "auth",
		Short: "Auth a Kafka cluster.",
		RunE:  c.auth,
	})

	return nil
}

func (c *Command) create(cmd *cobra.Command, args []string) error {
	return shared.ErrNotImplemented
}

func (c *Command) list(cmd *cobra.Command, args []string) error {
	req := &schedv1.KafkaCluster{AccountId: c.config.Auth.Account.Id}
	clusters, err := c.kafka.List(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	var data [][]string
	for _, cluster := range clusters {
		data = append(data, common.ToRow(cluster, listFields))
	}
	common.RenderTable(data, listLabels)
	return nil
}

func (c *Command) describe(cmd *cobra.Command, args []string) error {
	req := &schedv1.KafkaCluster{AccountId: c.config.Auth.Account.Id, Name: args[0]}
	cluster, err := c.kafka.Describe(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	common.RenderDetail(cluster, describeFields, describeLabels)
	return nil
}

func (c *Command) update(cmd *cobra.Command, args []string) error {
	return shared.ErrNotImplemented
}

func (c *Command) delete(cmd *cobra.Command, args []string) error {
	return shared.ErrNotImplemented
}

func (c *Command) auth(cmd *cobra.Command, args []string) error {
	return shared.ErrNotImplemented
}
