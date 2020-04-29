package connector

import (
	"context"
	"os"

	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	opv1 "github.com/confluentinc/cc-structs/operator/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
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
func New(prerunner pcmd.PreRunner, config *v3.Config) *cobra.Command {
	cmd := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:   "connector",
				Short: "Manage Kafka Connect.",
			}, config, prerunner),
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a connector.",
		Example: FormatDescription(`
Describe connector and task level details of a connector in the current or specified Kafka cluster context.

::

        {{.CLIName}} connector describe <id>
        {{.CLIName}} connector describe <id> --cluster <cluster-id>		`, c.Config.CLIName),
		RunE: c.describe,
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "list",
		Short: "List connectors.",
		Example: FormatDescription(`
List connectors in the current or specified Kafka cluster context.

::

        {{.CLIName}} connector list
        {{.CLIName}} connector list --cluster <cluster-id>		`, c.Config.CLIName),
		RunE: c.list,
		Args: cobra.NoArgs,
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "create",
		Short: "Create a connector.",
		Example: FormatDescription(`
Create connector in the current or specified Kafka cluster context.

::

        {{.CLIName}} connector create --config <file>
        {{.CLIName}} connector create --cluster <cluster-id> --config <file>		`, c.Config.CLIName),
		RunE: c.create,
		Args: cobra.NoArgs,
	}
	cmd.Flags().String("config", "", "JSON connector config file.")
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	panicOnError(cmd.MarkFlagRequired("config"))
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a connector.",
		Example: FormatDescription(`
Delete connector in the current or specified Kafka cluster context.

::

        {{.CLIName}} connector delete <id>
        {{.CLIName}} connector delete <id> --cluster <cluster-id>	`, c.Config.CLIName),
		RunE: c.delete,
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "update <id>",
		Short: "Update connector configuration.",
		RunE:  c.update,
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().String("config", "", "JSON connector config file.")
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	panicOnError(cmd.MarkFlagRequired("config"))
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "pause <id>",
		Short: "Pause a connector.",
		Example: FormatDescription(`
Pause connector in the current or specified Kafka cluster context.

::

        {{.CLIName}} connector pause <connector-id>
        {{.CLIName}} connector pause <connector-id> --cluster <cluster-id>	`, c.Config.CLIName),
		RunE: c.pause,
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "resume <id>",
		Short: "Resume a connector.",
		Example: FormatDescription(`
Resume connector in the current or specified Kafka cluster context.

::

        {{.CLIName}} connector resume <id>
        {{.CLIName}} connector resume <id> --cluster <cluster-id>	`, c.Config.CLIName),
		RunE: c.resume,
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connectors, err := c.Client.Connect.ListWithExpansions(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.Id}, "status,info,id")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listFields, listStructuredLabels)
	if err != nil {
		return errors.HandleCommon(err, cmd)
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
	kafkaCluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connector, err := c.Client.Connect.GetExpansionById(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.Id, Id: args[0]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	outputOption, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if outputOption == output.Human.String() {
		return printHumanDescribe(cmd, connector)
	} else {
		return printStructuredDescribe(connector, outputOption)
	}
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	userConfigs, err := getConfig(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connector, err := c.Client.Connect.Create(context.Background(), &schedv1.ConnectorConfig{UserConfigs: *userConfigs, AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.Id, Name: (*userConfigs)["name"], Plugin: (*userConfigs)["connector.class"]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	// Resolve Connector ID from Name of created connector
	connectorExpansion, err := c.Client.Connect.GetExpansionByName(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.Id, Name: connector.Name})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	trace := connectorExpansion.Status.Connector.Trace
	if outputFormat == output.Human.String() {
		pcmd.Printf(cmd, "Created connector %s %s\n", connector.Name, connectorExpansion.Id.Id)
		if trace != "" {
			pcmd.Printf(cmd, "Error Trace: %s\n", trace)
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
	return nil
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	userConfigs, err := getConfig(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	kafkaCluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	// Resolve Connector Name from ID
	connector, err := c.Client.Connect.GetExpansionById(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.Id, Id: args[0]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	_, err = c.Client.Connect.Update(context.Background(), &schedv1.ConnectorConfig{UserConfigs: *userConfigs, AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.Id, Name: connector.Info.Name, Plugin: (*userConfigs)["connector.class"]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "Updated connector "+args[0])
	return nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connector, err := c.Client.Connect.GetExpansionById(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.Id, Id: args[0]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = c.Client.Connect.Delete(context.Background(), &schedv1.Connector{Name: connector.Info.Name, AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "Successfully deleted connector")
	return nil
}

func (c *command) pause(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connector, err := c.Client.Connect.GetExpansionById(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.Id, Id: args[0]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = c.Client.Connect.Pause(context.Background(), &schedv1.Connector{Name: connector.Info.Name, AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "Successfully paused connector")
	return nil
}

func (c *command) resume(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connector, err := c.Client.Connect.GetExpansionById(context.Background(), &schedv1.Connector{AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.Id, Id: args[0]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = c.Client.Connect.Resume(context.Background(), &schedv1.Connector{Name: connector.Info.Name, AccountId: c.EnvironmentId(), KafkaClusterId: kafkaCluster.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "Successfully resumed connector")
	return nil
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func printHumanDescribe(cmd *cobra.Command, connector *opv1.ConnectorExpansion) error {
	pcmd.Println(cmd, "Connector Details")
	data := &connectorDescribeDisplay{
		Name:   connector.Status.Name,
		ID:     connector.Id.Id,
		Status: connector.Status.Connector.State,
		Type:   connector.Info.Type,
		Trace:  connector.Status.Connector.Trace,
	}
	_ = printer.RenderTableOut(data, listFields, describeRenames, os.Stdout)
	pcmd.Println(cmd, "\n\nTask Level Details")
	var tasks [][]string
	titleRow := []string{"TaskId", "State"}
	for _, task := range connector.Status.Tasks {
		tasks = append(tasks, printer.ToRow(&taskDescribeDisplay{
			task.Id,
			task.State,
		}, titleRow))
	}
	printer.RenderCollectionTable(tasks, titleRow)
	pcmd.Println(cmd, "\n\nConfiguration Details")
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
