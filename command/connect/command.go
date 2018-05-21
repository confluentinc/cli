package connect

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/ghodss/yaml"
	"github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	common "github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

var (
	listFields     = []string{"Name", "Plugin", "ServiceProvider", "Region", "Status"}
	listLabels     = []string{"Name", "Kind", "Provider", "Region", "Status"}
	describeFields = []string{"Name", "Plugin", "KafkaClusterId", "ServiceProvider", "Region", "Durability", "Status"}
	describeLabels = []string{"Name", "Kind", "Kafka", "Provider", "Region", "Durability", "Status"}
)

type Command struct {
	*cobra.Command
	config  *shared.Config
	connect Connect
}

func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "connect",
			Short: "Manage connectors.",
		},
		config: config,
	}
	err := cmd.init()
	return cmd.Command, err
}

func (c *Command) init() error {
	path, err := exec.LookPath("confluent-connect-plugin")
	if err != nil {
		return fmt.Errorf("skipping connect: plugin isn't installed")
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
	raw, err := rpcClient.Dispense("connect")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Got a client now communicating over RPC.
	c.connect = raw.(Connect)

	// All commands require login first
	c.Command.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if err := c.config.CheckLogin(); err != nil {
			common.HandleError(err)
			os.Exit(0) // TODO: this should be 1 but that prints "exit status 1" to the console
		}
	}

	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a connector.",
		RunE:  c.create,
		Args:  cobra.ExactArgs(1),
	}
	createCmd.Flags().String("config", "", "Connector configuration file")
	createCmd.MarkFlagRequired("config")
	createCmd.Flags().String("kafka-cluster", "", "Kafka Cluster Name")
	createCmd.MarkFlagRequired("kafka-cluster")
	c.AddCommand(createCmd)

	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List connectors.",
		RunE:  c.list,
	})

	c.AddCommand(&cobra.Command{
		Use:   "get <name>",
		Short: "Get a connector.",
		RunE:  c.get,
		Args:  cobra.ExactArgs(1),
	})

	c.AddCommand(&cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a connector.",
		RunE:  c.describe,
		Args:  cobra.ExactArgs(1),
	})

	c.AddCommand(&cobra.Command{
		Use:   "delete",
		Short: "Delete a connector.",
		RunE:  c.delete,
	})

	updateCmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a connector.",
		RunE:  c.update,
		Args:  cobra.ExactArgs(1),
	}
	updateCmd.Flags().String("config", "", "Connector configuration file")
	updateCmd.MarkFlagRequired("config")
	c.AddCommand(updateCmd)

	c.AddCommand(&cobra.Command{
		Use:   "auth",
		Short: "Auth a connector.",
		RunE:  c.auth,
	})

	return nil
}

func (c *Command) list(cmd *cobra.Command, args []string) error {
	req := &schedv1.ConnectCluster{AccountId: c.config.Auth.Account.Id}
	clusters, err := c.connect.List(context.Background(), req)
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

func (c *Command) get(cmd *cobra.Command, args []string) error {
	req := &schedv1.ConnectCluster{AccountId: c.config.Auth.Account.Id, Name: args[0]}
	cluster, err := c.connect.Describe(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	common.RenderDetail(cluster, describeFields, describeLabels)
	return nil
}

func (c *Command) create(cmd *cobra.Command, args []string) error {
	options, err := getConfig(cmd)

	kafkaClusterName, err := cmd.Flags().GetString("kafka-cluster")
	if err != nil {
		return errors.Wrap(err, "error reading --kafka-cluster as string")
	}

	// Create connect cluster config
	req := &schedv1.ConnectS3SinkClusterConfig{
		Name:           args[0],
		AccountId:      c.config.Auth.Account.Id,
		Options:        options,
		KafkaClusterId: kafkaClusterName, // NOTE: We store the name in the ID field and resolve the ID in the plugin.
	}

	cluster, err := c.connect.CreateS3Sink(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	fmt.Println("Created new connector:")
	common.RenderDetail(cluster.ConnectCluster, describeFields, describeLabels)
	fmt.Println("\nS3/Sink Options:")
	fmt.Println(toConfig(cluster.Options))
	fmt.Println("\n\nCreate an S3 bucket policy with this user ARN:\n\t" + cluster.UserArn)
	return nil
}

func (c *Command) describe(cmd *cobra.Command, args []string) error {
	req := &schedv1.ConnectCluster{AccountId: c.config.Auth.Account.Id, Name: args[0]}
	cluster, err := c.connect.Describe(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	switch cluster.Plugin {
	case schedv1.ConnectPlugin_S3_SINK:
		cl, err := c.connect.DescribeS3Sink(context.Background(), &schedv1.ConnectS3SinkCluster{
			ConnectCluster: &schedv1.ConnectCluster{Id: cluster.Id, AccountId: cluster.AccountId},
		})
		if err != nil {
			return common.HandleError(err)
		}
		common.RenderDetail(cl, describeFields, describeLabels)
		fmt.Println("\nS3 Sink Options:")
		common.RenderDetail(cl.Options, nil, nil)
	default:
		return fmt.Errorf("unknown connect plugin type: %s", cluster.Plugin.String())
	}
	return nil
}

func (c *Command) update(cmd *cobra.Command, args []string) error {
	options, err := getConfig(cmd)

	// Create updated connect s3-sink cluster
	req := &schedv1.ConnectS3SinkCluster{
		ConnectCluster: &schedv1.ConnectCluster{
			Name:      args[0],
			AccountId: c.config.Auth.Account.Id,
		},
		Options: options,
	}

	cluster, err := c.connect.UpdateS3Sink(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	fmt.Println("Updated connector:")
	common.RenderDetail(cluster.ConnectCluster, describeFields, describeLabels)
	fmt.Println("\nS3/Sink Options:")
	fmt.Println(toConfig(cluster.Options))
	fmt.Println("\n\nCreate an S3 bucket policy with this user ARN:\n\t" + cluster.UserArn)
	return nil
}

func (c *Command) delete(cmd *cobra.Command, args []string) error {
	req := &schedv1.ConnectCluster{AccountId: c.config.Auth.Account.Id, Name: args[0]}
	err := c.connect.Delete(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	fmt.Println("Your connect cluster has been deleted!")
	return nil
}

func (c *Command) auth(cmd *cobra.Command, args []string) error {
	return common.HandleError(shared.ErrNotImplemented)
}

func getConfig(cmd *cobra.Command) (*schedv1.ConnectS3SinkOptions, error) {
	// Set s3-sink connector options
	filename, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, errors.Wrap(err, "error reading --config as string")
	}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read config file %s", filename)
	}
	options := &schedv1.ConnectS3SinkOptions{}
	err = yaml.Unmarshal(yamlFile, options)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse config file %s", filename)
	}
	return options, nil
}

func toConfig(options *schedv1.ConnectS3SinkOptions) (string, error) {
	opts, err := yaml.Marshal(options)
	if err != nil {
		return "", errors.Wrapf(err, "unable to marshal options")
	}
	return string(opts), nil
}
