package connector_catalog

import (
	"context"
	"fmt"

	"github.com/c-bata/go-prompt"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	completableChildren []*cobra.Command
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
func New(cliName string, prerunner pcmd.PreRunner) *command {
	cmd := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(&cobra.Command{
			Use:   "connector-catalog",
			Short: "Catalog of connectors and their configurations.",
		}, prerunner),
	}
	cmd.init(cliName)
	return cmd
}

func (c *command) init(cliName string) {
	describeCmd := &cobra.Command{
		Use:   "describe <connector-type>",
		Short: "Describe a connector plugin type.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe required connector configuration parameters for a specific connector plugin.",
				Code: fmt.Sprintf("%s connector-catalog describe <plugin-name>", cliName),
			},
			examples.Example{
				Text: "With the ``--sample-file`` flag, create a sample connector configuration file.",
				Code: fmt.Sprintf("%s connector-catalog describe <plugin-name> --sample-file <filename>", cliName),
			},
		),
	}
	describeCmd.Flags().String("cluster", "", "Kafka cluster ID.")
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	c.AddCommand(describeCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List connector plugin types.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List connectors in the current or specified Kafka cluster context.",
				Code: fmt.Sprintf("%s connector-catalog list", cliName),
			},
		),
	}
	listCmd.Flags().String("cluster", "", "Kafka cluster ID.")
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)
	c.completableChildren = []*cobra.Command{describeCmd}
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	outputWriter, err := output.NewListOutputWriter(cmd, catalogFields, catalogFields, catalogStructureLabels)
	if err != nil {
		return err
	}
	catalog, err := c.getCatalog(cmd)
	if err != nil {
		return err
	}
	for _, conn := range catalog {
		outputWriter.AddElement(conn)
	}
	return outputWriter.Out()
}

func (c *command) getCatalog(cmd *cobra.Command) ([]*catalogDisplay, error) {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return nil, err
	}
	connectorInfo, err := c.Client.Connect.GetPlugins(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID}, "")
	if err != nil {
		return nil, err
	}
	var plugins []*catalogDisplay
	for _, conn := range connectorInfo {
		plugins = append(plugins, &catalogDisplay{
			PluginName: conn.Class,
			Type:       conn.Type,
		})
	}
	return plugins, nil
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.Errorf(errors.PluginNameNotPassedErrorMsg)
	}
	config := map[string]string{"connector.class": args[0]}

	reply, err := c.Client.Connect.Validate(context.Background(),
		&schedv1.ConnectorConfig{
			UserConfigs:    config,
			AccountId:      c.EnvironmentId(),
			KafkaClusterId: kafkaCluster.ID,
			Plugin:         args[0]})
	if reply != nil && err != nil {
		outputFormat, flagErr := cmd.Flags().GetString(output.FlagName)
		if flagErr != nil {
			return flagErr
		}
		if outputFormat == output.Human.String() {
			utils.Println(cmd, "Following are the required configs: \nconnector.class: "+args[0]+"\n"+err.Error())
		} else {

			for _, c := range reply.Configs {
				if len(c.Value.Errors) > 0 {
					config[c.Value.Name] = fmt.Sprintf("%s ", c.Value.Errors[0])
				}
			}
			return output.StructuredOutput(outputFormat, &config)
		}
		return nil
	}
	return errors.Errorf(errors.InvalidCloudErrorMsg)
}

func (c *command) Cmd() *cobra.Command {
	return c.Command
}

func (c *command) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	if !pcmd.CanCompleteCommand(c.Command) {
		return suggestions
	}
	catalog, err := c.getCatalog(c.Command)
	if err != nil {
		return suggestions
	}
	for _, conn := range catalog {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        conn.PluginName,
			Description: conn.Type,
		})
	}
	return suggestions
}

func (c *command) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}
