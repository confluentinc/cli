package connector_catalog

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	connectv1 "github.com/confluentinc/ccloudapis/connect/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

type catalogDisplay struct {
	PluginName string
	Type       string
}

var (
	catalogFields          = []string{"PluginName", "Type"}
	catalogStructureLabels = []string{"plugin_name", "type"}
)

// New returns the default command object for interacting with Connect.
func New(prerunner pcmd.PreRunner, config *v2.Config) *cobra.Command {
	cmd := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(&cobra.Command{
			Use:   "connector-catalog",
			Short: "Catalog of connectors and their configurations.",
		}, config, prerunner),
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

        {{.CLIName}} connector-catalog describe <connector-type>`, c.Config.CLIName),
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

        {{.CLIName}} connector-catalog list`, c.Config.CLIName),
		RunE: c.list,
		Args: cobra.NoArgs,
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connectorInfo, err := c.Client.Connect.GetPlugins(context.Background(), &connectv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.Id}, "")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	outputWriter, err := output.NewListOutputWriter(cmd, catalogFields, catalogFields, catalogStructureLabels)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	for _, conn := range connectorInfo {
		connector := &catalogDisplay{
			PluginName: conn.Class,
			Type:       conn.Type,
		}
		outputWriter.AddElement(connector)
	}
	return outputWriter.Out()
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if len(args) == 0 {
		return errors.HandleCommon(errors.ErrNoPluginName, cmd)
	}
	_, err = c.Client.Connect.Validate(context.Background(),
		&connectv1.ConnectorConfig{
			UserConfigs:    map[string]string{"connector.class": args[0]},
			AccountId:      c.EnvironmentId(),
			KafkaClusterId: kafkaCluster.Id,
			Plugin:         args[0]},
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
