package connector_catalog

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"
	connectv1 "github.com/confluentinc/ccloudapis/connect/v1"
	"github.com/confluentinc/go-printer"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

type command struct {
	*cobra.Command
	config *config.Config
	client ccloud.Connect
	ch     *pcmd.ConfigHelper
}

type catalogDisplay struct {
	PluginName string
	Type       string
}

var (
	catalogFields = []string{"PluginName", "Type"}
)

// New returns the default command object for interacting with Connect.
func New(prerunner pcmd.PreRunner, config *config.Config, client ccloud.Connect, ch *pcmd.ConfigHelper) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "connector-catalog",
			Short:             "Catalog of connectors and their configurations.",
			PersistentPreRunE: prerunner.Authenticated(),
		},
		config: config,
		client: client,
		ch:     ch,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	cmd := &cobra.Command{
		Use:   "describe <connector-type>",
		Short: "Describe a connector plugin type.",
		Example: FormatDescription(`
Describe required connector configuration parameters for a specific connector plugin.

::

        {{.CLIName}} connector-catalog describe <connector-type>`, c.config.CLIName),
		RunE: c.describe,
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "list",
		Short: "List connector plugin types.",
		Example: FormatDescription(`
List connectors in the current or specified Kafka cluster context.

::

        {{.CLIName}} connector-catalog list`, c.config.CLIName),
		RunE: c.list,
		Args: cobra.NoArgs,
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.GetKafkaCluster(cmd, c.ch)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connectorInfo, err := c.client.GetPlugins(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id}, "")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	var data [][]string
	for _, conn := range connectorInfo {
		connector := &catalogDisplay{
			PluginName: conn.Class,
			Type:       conn.Type,
		}
		data = append(data, printer.ToRow(connector, catalogFields))
	}
	printer.RenderCollectionTable(data, catalogFields)
	return nil
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.GetKafkaCluster(cmd, c.ch)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if len(args) == 0 {
		return errors.HandleCommon(errors.ErrNoPluginName, cmd)
	}
	_, err = c.client.Validate(context.Background(),
		&connectv1.ConnectorConfig{
			UserConfigs: map[string]string{"connector.class": args[0]},
			AccountId: c.config.Auth.Account.Id,
			KafkaClusterId: kafkaCluster.Id,
			Plugin: args[0]},
			false)

	if err != nil {
		pcmd.Println(cmd, "Following are the required configs: \nconnector.class \n"+err.Error())
		return nil
	}

	return errors.HandleCommon(errors.ErrInvalidCloud, cmd)
}

func FormatDescription(description string, cliName string) string {
	return strings.ReplaceAll(description, "{{.CLIName}}", cliName)
}
