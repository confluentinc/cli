package connector

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"
	connectv1 "github.com/confluentinc/ccloudapis/connect/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/go-printer"
)

type command struct {
	*cobra.Command
	config     *config.Config
	client     ccloud.Connect
	userClient ccloud.User
	ch         *pcmd.ConfigHelper
}

type describeDisplay struct {
	Name     string
	ID       string
	Status   string
	MaxTasks string
	Type     string
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
	c.AddCommand(&cobra.Command{
		Use:   "describe",
		Short: "Describe connectors in current Kafka cluster context.",
		RunE:  c.describe,
		Args:  cobra.MaximumNArgs(1),
	})

	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List connectors in current Kafka cluster context.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})

	createCmd := &cobra.Command{
		Use:   "create --config <config>",
		Short: "Create connector in the current Kafka cluster context.",
		RunE:  c.create,
		Args:  cobra.NoArgs,
	}
	createCmd.Flags().String("config", "", "YAML connector config file")
	check(createCmd.MarkFlagRequired("config"))
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)
	deleteCmd := &cobra.Command{
		Use:   "delete <connector-id>",
		Short: "Delete connector in the current Kafka cluster context.",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	}
	c.AddCommand(deleteCmd)
	updateCmd := &cobra.Command{
		Use:   "update <connector-id> --config <config>",
		Short: "Update connector in the current Kafka cluster context.",
		RunE:  c.update,
		Args:  cobra.ExactArgs(1),
	}
	updateCmd.Flags().String("config", "", "YAML connector config file")
	check(updateCmd.MarkFlagRequired("config"))
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	pauseCmd := &cobra.Command{
		Use:   "pause <connector-id>",
		Short: "Pause a connector.",
		RunE:  c.pause,
		Args:  cobra.ExactArgs(1),
	}
	pauseCmd.Flags().StringP("output", "o", "", "Output format")
	c.AddCommand(pauseCmd)

	resumeCmd := &cobra.Command{
		Use:   "resume <connector-id>",
		Short: "Resume a connector.",
		RunE:  c.resume,
		Args:  cobra.ExactArgs(1),
	}
	resumeCmd.Flags().StringP("output", "o", "", "Output format")
	c.AddCommand(resumeCmd)

	restartCmd := &cobra.Command{
		Use:   "restart <connector-id>",
		Short: "Restart a connector.",
		RunE:  c.restart,
		Args:  cobra.ExactArgs(1),
	}
	restartCmd.Flags().StringP("output", "o", "", "Output format")
	c.AddCommand(restartCmd)
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
	if len(args) == 0 {
		return errors.New("Connector ID must be passed")
	}
	connector, err := c.client.GetByID(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Id: args[0]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "Connector Details\n")
	data := &describeDisplay{
		Name:   connector.Status.Name,
		ID:     connector.Id.Id,
		Status: connector.Status.Connector.State,
		Type:   connector.Info.Type,
	}
	_ = printer.RenderTableOut(data, listFields, describeRenames, os.Stdout)

	pcmd.Println(cmd, "\n\nTask Level Details\n")
	var tasks [][]string
	titleRow := []string{"TaskID", "State"}
	for _, task := range connector.Status.Tasks {

		record := &struct {
			TaskID int32
			State  string
		}{
			task.Id,
			task.State,
		}
		tasks = append(tasks, printer.ToRow(record, titleRow))
	}
	printer.RenderCollectionTable(tasks, titleRow)
	pcmd.Println(cmd, "\n\nConfiguration Details\n\n")
	var configs [][]string
	titleRow = []string{"ConfigName", "ConfigValue"}
	for name, value := range connector.Info.Config {

		record := &struct {
			ConfigName  string
			ConfigValue string
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
	connector, err := c.client.CreateOrUpdate(context.Background(), &connectv1.ConnectorConfig{UserConfigs: userConfigs, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "Created connector"+connector.Id+" "+connector.Name)
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
	connector, err := c.client.GetByID(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Id: args[0]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connectorUpdated, err := c.client.CreateOrUpdate(context.Background(), &connectv1.ConnectorConfig{UserConfigs: userConfigs, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Name: connector.Info.Name})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "Updated connector"+connectorUpdated.Id+" "+connectorUpdated.Name)
	return nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.GetKafkaCluster(cmd, c.ch)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if len(args) == 0 {
		return errors.New("Connector ID must be passed")
	}
	connector, err := c.client.GetByID(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Id: args[0]})
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
	if len(args) == 0 {
		return errors.New("Connector ID must be passed")
	}
	connector, err := c.client.GetByID(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Id: args[0]})
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
	if len(args) == 0 {
		return errors.New("Connector ID must be passed")
	}
	connector, err := c.client.GetByID(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Id: args[0]})
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

func (c *command) restart(cmd *cobra.Command, args []string) error {

	kafkaCluster, err := pcmd.GetKafkaCluster(cmd, c.ch)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if len(args) == 0 {
		return errors.New("Connector ID must be passed")
	}
	connector, err := c.client.GetByID(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id, Id: args[0]})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = c.client.Restart(context.Background(), &connectv1.Connector{Name: connector.Info.Name, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "Successfully restarted connector")
	return nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
