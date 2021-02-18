package connector

import (
	"context"
	"fmt"
	"os"

	"github.com/c-bata/go-prompt"

	"github.com/confluentinc/cli/internal/pkg/examples"

	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	opv1 "github.com/confluentinc/cc-structs/operator/v1"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
	completableChildren []*cobra.Command
	analyticsClient     analytics.Client
}

type connectorDescribeDisplay struct {
	Name   string `json:"name" yaml:"name"`
	ID     string `json:"id" yaml:"id"`
	Status string `json:"status" yaml:"status"`
	Type   string `json:"type" yaml:"type"`
	Trace  string `json:"trace,omitempty" yaml:"trace,omitempty"`
}

type taskDescribeDisplay struct {
	TaskId int32  `json:"task_id" yaml:"task_id"`
	State  string `json:"state" yaml:"state"`
}
type configDescribeDisplay struct {
	Config string `json:"config" yaml:"config"`
	Value  string `json:"value" yaml:"value"`
}

type structuredDescribeDisplay struct {
	Connector *connectorDescribeDisplay `json:"connector" yaml:"connector"`
	Tasks     []taskDescribeDisplay     `json:"tasks" yaml:"task"`
	Configs   []configDescribeDisplay   `json:"configs" yaml:"configs"`
}

var (
	describeRenames      = map[string]string{}
	listFields           = []string{"ID", "Name", "Status", "Type", "Trace"}
	listStructuredLabels = []string{"id", "name", "status", "type", "trace"}
)

// New returns the default command object for interacting with Connect.
func New(cliName string, prerunner pcmd.PreRunner, analyticsClient analytics.Client) *command {
	cmd := &command{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(
			&cobra.Command{
				Use:   "connector",
				Short: "Manage Kafka Connect.",
			}, prerunner, SubcommandFlags),
		analyticsClient: analyticsClient,
	}
	cmd.init(cliName)
	return cmd
}

func (c *command) init(cliName string) {
	describeCmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a connector.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe connector and task level details of a connector in the current or specified Kafka cluster context.",
				Code: fmt.Sprintf("%s connector describe <id>\n%s connector describe <id> --cluster <cluster-id>", cliName, cliName),
			},
		),
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	c.AddCommand(describeCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List connectors.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List connectors in the current or specified Kafka cluster context.",
				Code: fmt.Sprintf("%s connector list\n%s connector list --cluster <cluster-id>", cliName, cliName),
			},
		),
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a connector.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a connector in the current or specified Kafka cluster context.",
				Code: fmt.Sprintf("%s connector create --config <file>\n%s connector create --cluster <cluster-id> --config <file>", cliName, cliName),
			},
		),
	}
	createCmd.Flags().String("config", "", "JSON connector config file.")
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	panicOnError(createCmd.MarkFlagRequired("config"))
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a connector.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.delete),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete a connector in the current or specified Kafka cluster context.",
				Code: fmt.Sprintf("%s connector delete --config <file>\n%s connector delete --cluster <cluster-id> --config <file>", cliName, cliName),
			},
		),
	}
	c.AddCommand(deleteCmd)

	updateCmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a connector configuration.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.update),
	}
	updateCmd.Flags().String("config", "", "JSON connector config file.")
	panicOnError(updateCmd.MarkFlagRequired("config"))
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	pauseCmd := &cobra.Command{
		Use:   "pause <id>",
		Short: "Pause a connector.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.pause),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Pause a connector in the current or specified Kafka cluster context.",
				Code: fmt.Sprintf("%s connector pause --config <file>\n%s connector pause --cluster <cluster-id> --config <file>", cliName, cliName),
			},
		),
	}
	c.AddCommand(pauseCmd)

	resumeCmd := &cobra.Command{
		Use:   "resume <id>",
		Short: "Resume a connector.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.resume),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Resume a connector in the current or specified Kafka cluster context.",
				Code: fmt.Sprintf("%s connector resume --config <file>\n%s connector resume --cluster <cluster-id> --config <file>", cliName, cliName),
			},
		),
	}
	c.AddCommand(resumeCmd)
	c.completableChildren = []*cobra.Command{deleteCmd, describeCmd, pauseCmd, resumeCmd, updateCmd}
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}
	connectors, err := c.Client.Connect.ListWithExpansions(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID}, "status,info,id")
	if err != nil {
		return err
	}
	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listFields, listStructuredLabels)
	if err != nil {
		return err
	}
	for name, connector := range connectors {
		connector := &connectorDescribeDisplay{
			Name:   name,
			ID:     connector.Id.Id,
			Status: connector.Status.Connector.State,
			Type:   connector.Info.Type,
			Trace:  connector.Status.Connector.Trace,
		}
		outputWriter.AddElement(connector)
	}
	return outputWriter.Out()
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}
	connector, err := c.Client.Connect.GetExpansionById(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID, Id: args[0]})
	if err != nil {
		return err
	}

	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	if outputOption == output.Human.String() {
		return printHumanDescribe(cmd, connector)
	} else {
		return printStructuredDescribe(connector, outputOption)
	}
}

func (c *command) create(cmd *cobra.Command, _ []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}
	userConfigs, err := getConfig(cmd)
	if err != nil {
		return err
	}
	connector, err := c.Client.Connect.Create(context.Background(), &schedv1.ConnectorConfig{UserConfigs: *userConfigs, AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID, Name: (*userConfigs)["name"], Plugin: (*userConfigs)["connector.class"]})
	if err != nil {
		return err
	}
	// Resolve Connector ID from Name of created connector
	connectorExpansion, err := c.Client.Connect.GetExpansionByName(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID, Name: connector.Name})
	if err != nil {
		return err
	}
	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	trace := connectorExpansion.Status.Connector.Trace
	if outputFormat == output.Human.String() {
		utils.Printf(cmd, errors.CreatedConnectorMsg, connector.Name, connectorExpansion.Id.Id)
		if trace != "" {
			utils.Printf(cmd, "Error Trace: %s\n", trace)
		}
	} else {
		return output.StructuredOutput(outputFormat, &struct {
			ConnectorName string `json:"name" yaml:"name"`
			Id            string `json:"id" yaml:"id"`
			Trace         string `json:"error_trace,omitempty" yaml:"error_trace,omitempty"`
		}{
			ConnectorName: connector.Name,
			Id:            connectorExpansion.Id.Id,
			Trace:         trace,
		})
	}
	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, connectorExpansion.Id.Id)
	return nil
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	userConfigs, err := getConfig(cmd)
	if err != nil {
		return err
	}
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}
	// Resolve Connector Name from ID
	connector, err := c.Client.Connect.GetExpansionById(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID, Id: args[0]})
	if err != nil {
		return err
	}
	_, err = c.Client.Connect.Update(context.Background(), &schedv1.ConnectorConfig{UserConfigs: *userConfigs, AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID, Name: connector.Info.Name, Plugin: (*userConfigs)["connector.class"]})
	if err != nil {
		return err
	}
	utils.Printf(cmd, errors.UpdatedConnectorMsg, args[0])
	return nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}
	connector, err := c.Client.Connect.GetExpansionById(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID, Id: args[0]})
	if err != nil {
		return err
	}
	err = c.Client.Connect.Delete(context.Background(), &schedv1.Connector{Name: connector.Info.Name, AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID})
	if err != nil {
		return err
	}
	utils.Printf(cmd, errors.DeletedConnectorMsg, args[0])
	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, connector.Id.Id)
	return nil
}

func (c *command) pause(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}
	connector, err := c.Client.Connect.GetExpansionById(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID, Id: args[0]})
	if err != nil {
		return err
	}
	err = c.Client.Connect.Pause(context.Background(), &schedv1.Connector{Name: connector.Info.Name, AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID})
	if err != nil {
		return err
	}
	utils.Printf(cmd, errors.PausedConnectorMsg, args[0])
	return nil
}

func (c *command) resume(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}
	connector, err := c.Client.Connect.GetExpansionById(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID, Id: args[0]})
	if err != nil {
		return err
	}
	err = c.Client.Connect.Resume(context.Background(), &schedv1.Connector{Name: connector.Info.Name, AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID})
	if err != nil {
		return err
	}
	utils.Printf(cmd, errors.ResumedConnectorMsg, args[0])
	return nil
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func printHumanDescribe(cmd *cobra.Command, connector *opv1.ConnectorExpansion) error {
	utils.Println(cmd, "Connector Details")
	data := &connectorDescribeDisplay{
		Name:   connector.Status.Name,
		ID:     connector.Id.Id,
		Status: connector.Status.Connector.State,
		Type:   connector.Info.Type,
		Trace:  connector.Status.Connector.Trace,
	}
	_ = printer.RenderTableOut(data, listFields, describeRenames, os.Stdout)
	utils.Println(cmd, "\n\nTask Level Details")
	var tasks [][]string
	titleRow := []string{"TaskId", "State"}
	for _, task := range connector.Status.Tasks {
		tasks = append(tasks, printer.ToRow(&taskDescribeDisplay{
			task.Id,
			task.State,
		}, titleRow))
	}
	printer.RenderCollectionTable(tasks, titleRow)
	utils.Println(cmd, "\n\nConfiguration Details")
	var configs [][]string
	titleRow = []string{"Config", "Value"}
	for name, value := range connector.Info.Config {
		configs = append(configs, printer.ToRow(&configDescribeDisplay{
			name,
			value,
		}, titleRow))
	}
	printer.RenderCollectionTable(configs, titleRow)
	return nil
}

func printStructuredDescribe(connector *opv1.ConnectorExpansion, format string) error {
	structuredDisplay := &structuredDescribeDisplay{
		Connector: &connectorDescribeDisplay{
			Name:   connector.Status.Name,
			ID:     connector.Id.Id,
			Status: connector.Status.Connector.State,
			Type:   connector.Info.Type,
			Trace:  connector.Status.Connector.Trace,
		},
		Tasks:   []taskDescribeDisplay{},
		Configs: []configDescribeDisplay{},
	}
	for _, task := range connector.Status.Tasks {
		structuredDisplay.Tasks = append(structuredDisplay.Tasks, taskDescribeDisplay{
			TaskId: task.Id,
			State:  task.State,
		})
	}
	for name, value := range connector.Info.Config {
		structuredDisplay.Configs = append(structuredDisplay.Configs, configDescribeDisplay{
			Config: name,
			Value:  value,
		})
	}
	return output.StructuredOutput(format, structuredDisplay)
}

func (c *command) Cmd() *cobra.Command {
	return c.Command
}

func (c *command) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}

func (c *command) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	if !pcmd.CanCompleteCommand(c.Command) {
		return suggestions
	}
	connectors, err := c.fetchConnectors()
	if err != nil {
		return suggestions
	}
	for _, conn := range connectors {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        conn.Id.Id,
			Description: conn.Info.Name,
		})
	}
	return suggestions
}

func (c *command) fetchConnectors() (map[string]*opv1.ConnectorExpansion, error) {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(c.Command)
	if err != nil {
		return nil, err
	}
	connectors, err := c.Client.Connect.ListWithExpansions(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.ID}, "status,info,id")
	if err != nil {
		return nil, err
	}
	return connectors, nil

}
