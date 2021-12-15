package connect

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
	"github.com/confluentinc/cli/internal/pkg/version"
)

type pluginCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	completableChildren []*cobra.Command
}

type pluginDisplay struct {
	PluginName string
	Type       string
}

var (
	pluginFields          = []string{"PluginName", "Type"}
	pluginHumanFields     = []string{"Plugin Name", "Type"}
	pluginStructureLabels = []string{"plugin_name", "type"}
)

// New returns the default command object for interacting with Connect.
func NewPluginCommand(prerunner pcmd.PreRunner) *cobra.Command {
	c := &pluginCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(&cobra.Command{
			Use:         "plugin",
			Short:       "Show plugins and their configurations.",
			Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		}, prerunner, SubcommandFlags),
	}
	c.init()
	return c.Command
}

func (c *pluginCommand) init() {
	describeCmd := &cobra.Command{
		Use:   "describe <plugin-type>",
		Short: "Describe a connector plugin type.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe required connector configuration parameters for a specific connector plugin.",
				Code: fmt.Sprintf("%s connect plugin describe <plugin-name>", version.CLIName),
			},
		),
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	c.AddCommand(describeCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List connector plugin types.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List connectors in the current or specified Kafka cluster context.",
				Code: fmt.Sprintf("%s connect plugin list", version.CLIName),
			},
		),
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	c.AddCommand(listCmd)
	c.completableChildren = []*cobra.Command{describeCmd}
}

func (c *pluginCommand) list(cmd *cobra.Command, _ []string) error {
	outputWriter, err := output.NewListOutputWriter(cmd, pluginFields, pluginHumanFields, pluginStructureLabels)
	if err != nil {
		return err
	}
	plugins, err := c.getPlugins(cmd)
	if err != nil {
		return err
	}
	for _, conn := range plugins {
		outputWriter.AddElement(conn)
	}
	outputWriter.StableSort()
	return outputWriter.Out()
}

func (c *pluginCommand) getPlugins(cmd *cobra.Command) ([]*pluginDisplay, error) {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return nil, err
	}
	connectorInfo, err := c.Client.Connect.GetPlugins(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID}, "")
	if err != nil {
		return nil, err
	}
	var plugins []*pluginDisplay
	for _, conn := range connectorInfo {
		plugins = append(plugins, &pluginDisplay{
			PluginName: conn.Class,
			Type:       conn.Type,
		})
	}
	return plugins, nil
}

func (c *pluginCommand) describe(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
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

func (c *pluginCommand) Cmd() *cobra.Command {
	return c.Command
}

func (c *pluginCommand) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	plugins, err := c.getPlugins(c.Command)
	if err != nil {
		return suggestions
	}
	for _, conn := range plugins {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        conn.PluginName,
			Description: conn.Type,
		})
	}
	return suggestions
}

func (c *pluginCommand) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}
