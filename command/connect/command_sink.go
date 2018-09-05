package connect

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/codyaray/go-printer"
	"github.com/codyaray/go-printer/editor"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	opv1 "github.com/confluentinc/cc-structs/operator/v1"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	connectv1 "github.com/confluentinc/cli/shared/connect"
)

var (
	listFields       = []string{"Id", "Name", "Plugin", "ServiceProvider", "Region", "Status"}
	listLabels       = []string{"Id", "Name", "Kind", "Provider", "Region", "Status"}
	describeFields   = []string{"Id", "Name", "Plugin", "KafkaClusterId", "ServiceProvider", "Region", "Durability", "Status"}
	describeRenames  = map[string]string{"Plugin": "Kind", "KafkaClusterId": "Kafka", "ServiceProvider": "Provider"}
	validPluginTypes = []string{"s3"}
)

type sinkCommand struct {
	*cobra.Command
	config  *shared.Config
	connect Connect
}

// NewSink returns the Cobra sinkCommand for Connect Sink.
func NewSink(config *shared.Config, connect Connect) (*cobra.Command, error) {
	cmd := &sinkCommand{
		Command: &cobra.Command{
			Use:   "sink",
			Short: "Manage sink connectors.",
		},
		config:  config,
		connect: connect,
	}
	err := cmd.init()
	return cmd.Command, err
}

func (c *sinkCommand) init() error {
	createCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a connector.",
		RunE:  c.create,
		Args:  cobra.ExactArgs(1),
	}
	createCmd.Flags().StringP("type", "t", "", fmt.Sprintf(`Connector type to create; must be one of "%s"`, strings.Join(validPluginTypes, `", "`)))
	check(createCmd.MarkFlagRequired("type"))
	createCmd.Flags().StringP("config", "f", "", "Connector configuration file")
	check(createCmd.MarkFlagRequired("config"))
	defaultKafkaCluster := ""
	if ctx, _ := c.config.Context(); ctx != nil {
		defaultKafkaCluster = ctx.Kafka
	}
	createCmd.Flags().StringP("kafka-cluster", "k", defaultKafkaCluster, "Kafka Cluster ID")
	if defaultKafkaCluster == "" {
		check(createCmd.MarkFlagRequired("kafka-cluster"))
	}
	defaultKafkaUser := ""
	if c.config.Auth != nil && c.config.Auth.User != nil {
		defaultKafkaUser = c.config.Auth.User.Email
	}
	createCmd.Flags().StringP("kafka-user", "u", defaultKafkaUser, "Kafka User Email")
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

	editCmd := &cobra.Command{
		Use:   "edit ID",
		Short: "Edit a connector.",
		RunE:  c.edit,
		Args:  cobra.ExactArgs(1),
	}
	editCmd.Flags().StringP("output", "o", "yaml", "Output format")
	c.AddCommand(editCmd)

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

func (c *sinkCommand) list(cmd *cobra.Command, args []string) error {
	req := &schedv1.ConnectCluster{AccountId: c.config.Auth.Account.Id}
	clusters, err := c.connect.List(context.Background(), req)
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

func (c *sinkCommand) get(cmd *cobra.Command, args []string) error {
	outputFormat, err := cmd.Flags().GetString("output")
	if err != nil {
		return errors.Wrap(err, "error reading --output as string")
	}

	req := &schedv1.ConnectCluster{AccountId: c.config.Auth.Account.Id, Id: args[0]}
	cluster, err := c.connect.Describe(context.Background(), req)
	if err != nil {
		return common.HandleError(err, cmd)
	}
	if err = printer.Render(cluster, describeFields, describeRenames, nil, outputFormat); err != nil {
		return common.HandleError(err, cmd)
	}
	return nil
}

func (c *sinkCommand) create(cmd *cobra.Command, args []string) error {
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
	case "s3":
		return c.createS3Sink(kafkaClusterID, kafkaUserEmail, cmd, args)
	}
	return nil
}

func (c *sinkCommand) createS3Sink(kafkaClusterID, kafkaUserEmail string, cmd *cobra.Command, args []string) error {
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
		return common.HandleError(err, cmd)
	}
	fmt.Println("Created new connector:")
	printer.RenderTableOut(cluster.ConnectCluster, describeFields, describeRenames, os.Stdout)
	fmt.Println("\nS3/Sink Options:")
	fmt.Println(toConfig(cluster.Options))
	fmt.Println("\n\nCreate an S3 bucket policy with this user ARN:\n\t" + cluster.UserArn)
	return nil
}

func (c *sinkCommand) describe(cmd *cobra.Command, args []string) error {
	cluster, err := c.fetch(args[0])
	if err != nil {
		return common.HandleError(err, cmd)
	}
	switch cl := cluster.(type) {
	case *schedv1.ConnectS3SinkCluster:
		printer.RenderTableOut(cl, describeFields, describeRenames, os.Stdout)
		fmt.Println("\nS3 Sink Options:")
		printer.RenderTableOut(cl.Options, nil, nil, os.Stdout)
	default:
		return fmt.Errorf("unknown cluster type: %v", cl)
	}
	return nil
}

func (c *sinkCommand) edit(cmd *cobra.Command, args []string) error {
	outputFormat, err := cmd.Flags().GetString("output")
	if err != nil {
		return errors.Wrap(err, "error reading --output as string")
	}

	cluster, err := c.fetch(args[0])
	if err != nil {
		return common.HandleError(err, cmd)
	}
	var objType interface{}
	switch cl := cluster.(type) {
	case *schedv1.ConnectS3SinkCluster:
		objType = &schedv1.ConnectS3SinkCluster{}
	default:
		return fmt.Errorf("unknown cluster type: %v", cl)
	}

	formatter, err := printer.NewProtoFormatter(outputFormat)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	edit := editor.NewWithFormatter(formatter)
	updated, err := edit.Launch("editor-connect", cluster, objType)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	switch req := updated.(type) {
	case *schedv1.ConnectS3SinkCluster:
		cluster, err := c.connect.UpdateS3Sink(context.Background(), req)
		if err != nil {
			return common.HandleError(err, cmd)
		}
		fmt.Println("Updated connector:")
		printer.RenderTableOut(cluster, describeFields, describeRenames, os.Stdout)
	default:
		return fmt.Errorf("unknown edited object type: %v", updated)
	}
	return nil
}

func (c *sinkCommand) update(cmd *cobra.Command, args []string) error {
	cluster, err := c.fetch(args[0])
	if err != nil {
		return common.HandleError(err, cmd)
	}
	switch cl := cluster.(type) {
	case *schedv1.ConnectS3SinkCluster:
		options, err := getConfig(cmd)
		if err != nil {
			return err
		}
		req := &schedv1.ConnectS3SinkCluster{
			ConnectCluster: &schedv1.ConnectCluster{
				Id:        args[0],
				AccountId: c.config.Auth.Account.Id,
			},
			Options: options,
		}
		cluster, err := c.connect.UpdateS3Sink(context.Background(), req)
		if err != nil {
			return common.HandleError(err, cmd)
		}
		fmt.Println("Updated connector:")
		printer.RenderTableOut(cluster.ConnectCluster, describeFields, describeRenames, os.Stdout)
		fmt.Println("\nS3/Sink Options:")
		fmt.Println(toConfig(cluster.Options))
		fmt.Println("\n\nCreate an S3 bucket policy with this user ARN:\n\t" + cluster.UserArn)
	default:
		return fmt.Errorf("unknown cluster type: %v", cl)
	}
	return nil
}

func (c *sinkCommand) delete(cmd *cobra.Command, args []string) error {
	req := &schedv1.ConnectCluster{AccountId: c.config.Auth.Account.Id, Id: args[0]}
	err := c.connect.Delete(context.Background(), req)
	if err != nil {
		return common.HandleError(err, cmd)
	}
	fmt.Println("Your connect cluster has been deleted!")
	return nil
}

func (c *sinkCommand) auth(cmd *cobra.Command, args []string) error {
	return common.HandleError(shared.ErrNotImplemented, cmd)
}

//
// Helper methods
//

func (c *sinkCommand) fetch(id string) (interface{}, error) {
	req := &schedv1.ConnectCluster{AccountId: c.config.Auth.Account.Id, Id: id}
	cluster, err := c.connect.Describe(context.Background(), req)
	if err != nil {
		return nil, err
	}
	switch cluster.Plugin {
	case opv1.ConnectPlugin_S3_SINK:
		cl, err := c.connect.DescribeS3Sink(context.Background(), &schedv1.ConnectS3SinkCluster{
			ConnectCluster: &schedv1.ConnectCluster{Id: cluster.Id, AccountId: cluster.AccountId},
		})
		if err != nil {
			return nil, err
		}
		return cl, nil
	default:
		return nil, fmt.Errorf("unknown connect plugin type: %s", cluster.Plugin.String())
	}
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
			return fmt.Errorf(`invalid flag value: --%v "%v", value must be one of "%v"`, flag, pluginType, strings.Join(options, `", "`))
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
