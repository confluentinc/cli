package connector

import (
	"context"
	"os"

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

type describeDisplay struct {
	Name   string
	ID     string
	Status string
	Type   string
}

var (
	describeRenames = map[string]string{}
	listFields      = []string{"ID", "Name", "Status", "Type"}
)

// New returns the default command object for interacting with Connect.
func New(prerunner pcmd.PreRunner, config *config.Config, client ccloud.Connect, ch *pcmd.ConfigHelper) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "connector",
			Short:             "Manage Kafka Connect.",
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
		Use:   "describe <id>",
		Short: "Describe a connector.",
		Example: FormatDescription(`
Describe connector and task level details of a connector in the current or specified Kafka cluster context.

::

        {{.CLIName}} connector describe <id>
        {{.CLIName}} connector describe <id> --cluster <cluster-id>		`, c.config.CLIName),
		RunE: c.describe,
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "list",
		Short: "List connectors.",
		Example: FormatDescription(`
List connectors in the current or specified Kafka cluster context.

::

        {{.CLIName}} connector list
        {{.CLIName}} connector list --cluster <cluster-id>		`, c.config.CLIName),
		RunE: c.list,
		Args: cobra.NoArgs,
	}
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "create",
		Short: "Create a connector.",
		Example: FormatDescription(`
Create connector in the current or specified Kafka cluster context.

::

        {{.CLIName}} connector create --config <file>
        {{.CLIName}} connector create --cluster <cluster-id> --config <file>		`, c.config.CLIName),
		RunE: c.create,
		Args: cobra.NoArgs,
	}
	cmd.Flags().String("config", "", "JSON connector config file.")
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
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
        {{.CLIName}} connector delete <id> --cluster <cluster-id>	`, c.config.CLIName),
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
        {{.CLIName}} connector pause <connector-id> --cluster <cluster-id>	`, c.config.CLIName),
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
        {{.CLIName}} connector resume <id> --cluster <cluster-id>	`, c.config.CLIName),
		RunE: c.resume,
		Args: cobra.ExactArgs(1),
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
	connectors, err := c.client.ListWithExpansions(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id}, "status,info,id")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	var data [][]string
	for name, connector := range connectors {
		connector := &describeDisplay{
			Name:   name,
			ID:     connector.Id.Id,
			Status: connector.Status.Connector.State,
			Type:   connector.Info.Type,
		}
		data = append(data, printer.ToRow(connector, listFields))
	}
	printer.RenderCollectionTable(data, listFields)
	return nil
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.GetKafkaCluster(cmd, c.ch)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connector, err := c.client.GetExpansionById(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Id: args[0]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "Connector Details")
	data := &describeDisplay{
		Name:   connector.Status.Name,
		ID:     connector.Id.Id,
		Status: connector.Status.Connector.State,
		Type:   connector.Info.Type,
	}
	_ = printer.RenderTableOut(data, listFields, describeRenames, os.Stdout)

	pcmd.Println(cmd, "\n\nTask Level Details")
	var tasks [][]string
	titleRow := []string{"Task_ID", "State"}
	for _, task := range connector.Status.Tasks {
		record := &struct {
			Task_ID int32
			State   string
		}{
			task.Id,
			task.State,
		}
		tasks = append(tasks, printer.ToRow(record, titleRow))
	}
	printer.RenderCollectionTable(tasks, titleRow)
	pcmd.Println(cmd, "\n\nConfiguration Details")
	var configs [][]string
	titleRow = []string{"Configuration", "Value"}
	for name, value := range connector.Info.Config {
		record := &struct {
			Configuration string
			Value         string
		}{
			name,
			value,
		}
		configs = append(configs, printer.ToRow(record, titleRow))
	}
	printer.RenderCollectionTable(configs, titleRow)
	return nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.GetKafkaCluster(cmd, c.ch)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	userConfigs, err := getConfig(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connector, err := c.client.Create(context.Background(), &connectv1.ConnectorConfig{UserConfigs: *userConfigs, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Name: (*userConfigs)["name"], Plugin: (*userConfigs)["connector.class"]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Printf(cmd, "Created connector %s ", connector.Name)
	// Resolve Connector ID from Name of created connector
	connectorID, err := c.client.GetExpansionByName(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Name: connector.Name})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, connectorID.Id.Id)
	return nil
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	userConfigs, err := getConfig(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	kafkaCluster, err := pcmd.GetKafkaCluster(cmd, c.ch)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	// Resolve Connector Name from ID
	connector, err := c.client.GetExpansionById(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Id: args[0]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	_, err = c.client.Update(context.Background(), &connectv1.ConnectorConfig{UserConfigs: *userConfigs, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Name: connector.Info.Name, Plugin: (*userConfigs)["connector.class"]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "Updated connector "+args[0])
	return nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.GetKafkaCluster(cmd, c.ch)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connector, err := c.client.GetExpansionById(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Id: args[0]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = c.client.Delete(context.Background(), &connectv1.Connector{Name: connector.Info.Name, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "Successfully deleted connector")
	return nil
}

func (c *command) pause(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.GetKafkaCluster(cmd, c.ch)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connector, err := c.client.GetExpansionById(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Id: args[0]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = c.client.Pause(context.Background(), &connectv1.Connector{Name: connector.Info.Name, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "Successfully paused connector")
	return nil
}

func (c *command) resume(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.GetKafkaCluster(cmd, c.ch)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connector, err := c.client.GetExpansionById(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Id: args[0]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = c.client.Resume(context.Background(), &connectv1.Connector{Name: connector.Info.Name, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id})
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
