package connect

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/codyaray/go-printer"
	"github.com/ghodss/yaml"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	connectv1 "github.com/confluentinc/cli/shared/connect"
)

var (
	listFields       = []string{"Id", "Name", "Plugin", "ServiceProvider", "Region", "Status"}
	listLabels       = []string{"Id", "Name", "Kind", "Provider", "Region", "Status"}
	describeFields   = []string{"Id", "Name", "Plugin", "KafkaClusterId", "ServiceProvider", "Region", "Durability", "Status"}
	describeRenames  = map[string]string{"Plugin": "Kind", "KafkaClusterId": "Kafka", "ServiceProvider": "Provider"}
	validPluginTypes = []string{"s3-sink"}
)

type command struct {
	*cobra.Command
	config  *shared.Config
	connect Connect
}

// New returns the Cobra command for Connect.
func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "connect",
			Short: "Manage connectors.",
		},
		config: config,
	}
	err := cmd.init()
	return cmd.Command, err
}

func (c *command) init() error {
	path, err := exec.LookPath("confluent-connect-plugin")
	if err != nil {
		return fmt.Errorf("skipping connect: plugin isn't installed")
	}

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Plugins:          shared.PluginMap,
		Cmd:              exec.Command("sh", "-c", path), // nolint: gas
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
			_ = common.HandleError(err)
			os.Exit(1)
		}
	}

	createCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a connector.",
		RunE:  c.create,
	}
	createCmd.Flags().StringP("type", "t", "", fmt.Sprintf(`Connector type to create; must be one of "%s"`, strings.Join(validPluginTypes, `", "`)))
	check(createCmd.MarkFlagRequired("type"))
	createCmd.Flags().StringP("config", "f", "", "Connector configuration file")
	check(createCmd.MarkFlagRequired("config"))
	createCmd.Flags().StringP("kafka-cluster", "k", "", "Kafka Cluster Name")
	check(createCmd.MarkFlagRequired("kafka-cluster"))
	createCmd.Flags().StringP("kafka-user", "u", "", "Kafka User Email")
	check(createCmd.MarkFlagRequired("kafka-user"))
	c.AddCommand(createCmd)

	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List connectors.",
		RunE:  c.list,
	})

	getCmd := &cobra.Command{
		Use:   "get ID",
		Short: "Get a connector.",
		RunE:  c.get,
		Args:  cobra.ExactArgs(1),
	}
	getCmd.Flags().StringP("output", "o", "", "Output format")
	c.AddCommand(getCmd)

	c.AddCommand(&cobra.Command{
		Use:   "describe ID",
		Short: "Describe a connector.",
		RunE:  c.describe,
		Args:  cobra.ExactArgs(1),
	})

	updateCmd := &cobra.Command{
		Use:   "update ID",
		Short: "Update a connector.",
		RunE:  c.update,
		Args:  cobra.ExactArgs(1),
	}
	updateCmd.Flags().String("config", "", "Connector configuration file")
	check(updateCmd.MarkFlagRequired("config"))
	c.AddCommand(updateCmd)

	c.AddCommand(&cobra.Command{
		Use:   "delete ID",
		Short: "Delete a connector.",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	})

	c.AddCommand(&cobra.Command{
		Use:   "auth",
		Short: "Auth a connector.",
		RunE:  c.auth,
	})

	return nil
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	req := &schedv1.ConnectCluster{AccountId: c.config.Auth.Account.Id}
	clusters, err := c.connect.List(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	var data [][]string
	for _, cluster := range clusters {
		data = append(data, printer.ToRow(cluster, listFields))
	}
	printer.RenderTable(data, listLabels)
	return nil
}

func (c *command) get(cmd *cobra.Command, args []string) error {
	outputFormat, err := cmd.Flags().GetString("output")
	if err != nil {
		return errors.Wrap(err, "error reading --output as string")
	}

	req := &schedv1.ConnectCluster{AccountId: c.config.Auth.Account.Id, Id: args[0]}
	cluster, err := c.connect.Describe(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	if err = printer.Render(cluster, describeFields, describeRenames, nil, outputFormat); err != nil {
		return common.HandleError(err)
	}
	return nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	if err := Enum("type", validPluginTypes...)(cmd, args); err != nil {
		return err
	}

	pluginType, err := cmd.Flags().GetString("type")
	if err != nil {
		return errors.Wrap(err, "error reading --type as string")
	}

	kafkaClusterID, err := cmd.Flags().GetString("kafka-cluster")
	if err != nil {
		return errors.Wrap(err, "error reading --kafka-cluster as string")
	}

	kafkaUserEmail, err := cmd.Flags().GetString("kafka-user")
	if err != nil {
		return errors.Wrap(err, "error reading --kafka-user as string")
	}

	switch pluginType {
	case "s3-sink":
		return c.createS3Sink(kafkaClusterID, kafkaUserEmail, cmd, args)
	}
	return nil
}

func (c *command) createS3Sink(kafkaClusterID, kafkaUserEmail string, cmd *cobra.Command, args []string) error {
	options, err := getConfig(cmd)
	if err != nil {
		return err
	}

	// Create connect cluster config
	req := &connectv1.ConnectS3SinkClusterConfig{
		Name:           args[0],
		AccountId:      c.config.Auth.Account.Id,
		Options:        options,
		KafkaClusterId: kafkaClusterID,
		KafkaUserEmail: kafkaUserEmail,
	}

	cluster, err := c.connect.CreateS3Sink(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	fmt.Println("Created new connector:")
	printer.RenderDetail(cluster.ConnectCluster, describeFields, describeRenames)
	fmt.Println("\nS3/Sink Options:")
	fmt.Println(toConfig(cluster.Options))
	fmt.Println("\n\nCreate an S3 bucket policy with this user ARN:\n\t" + cluster.UserArn)
	return nil
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	req := &schedv1.ConnectCluster{AccountId: c.config.Auth.Account.Id, Id: args[0]}
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
		printer.RenderDetail(cl, describeFields, describeRenames)
		fmt.Println("\nS3 Sink Options:")
		printer.RenderDetail(cl.Options, nil, nil)
	default:
		return fmt.Errorf("unknown connect plugin type: %s", cluster.Plugin.String())
	}
	return nil
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	options, err := getConfig(cmd)
	if err != nil {
		return err
	}

	// Create updated connect s3-sink cluster
	req := &schedv1.ConnectS3SinkCluster{
		ConnectCluster: &schedv1.ConnectCluster{
			Id:        args[0],
			AccountId: c.config.Auth.Account.Id,
		},
		Options: options,
	}

	cluster, err := c.connect.UpdateS3Sink(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	fmt.Println("Updated connector:")
	printer.RenderDetail(cluster.ConnectCluster, describeFields, describeRenames)
	fmt.Println("\nS3/Sink Options:")
	fmt.Println(toConfig(cluster.Options))
	fmt.Println("\n\nCreate an S3 bucket policy with this user ARN:\n\t" + cluster.UserArn)
	return nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	req := &schedv1.ConnectCluster{AccountId: c.config.Auth.Account.Id, Id: args[0]}
	err := c.connect.Delete(context.Background(), req)
	if err != nil {
		return common.HandleError(err)
	}
	fmt.Println("Your connect cluster has been deleted!")
	return nil
}

func (c *command) auth(cmd *cobra.Command, args []string) error {
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

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Enum is a cobra Args validator that ensures a flag value is one of a set of options.
func Enum(flag string, options ...string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		pluginType, err := cmd.Flags().GetString(flag)
		if err != nil {
			return err
		}
		if !contains(options, pluginType) {
			return fmt.Errorf(`invalid flag '--%v "%v"', value must be one of "%v"`, flag, pluginType, strings.Join(options, `", "`))
		}
		return nil
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
