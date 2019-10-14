package connect

import (
	"context"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"
	connectv1 "github.com/confluentinc/ccloudapis/connect/v1"
	"os"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*cobra.Command
	config      *config.Config
	client      ccloud.Connect
	kafkaClient ccloud.Kafka
	userClient  ccloud.User
	ch          *pcmd.ConfigHelper
}

type describeDisplay struct {
	Name            string
	ID              string
	Status          string
	Used            string
	Available       string
	Compatibility   string
	Mode            string
	ServiceProvider string
}

var (
	describeLabels  = []string{"Name", "ID", "Status", "Tasks", "Available", "Compatibility", "Mode", "ServiceProvider"}
	describeRenames = map[string]string{"ID": "Cluster ID", "URL": "Endpoint URL", "Used": "Used Schemas", "Available": "Available Schemas", "Compatibility": "Global Compatibility", "ServiceProvider": "Service Provider"}
	enableLabels    = []string{"Id", "Endpoint"}
	enableRenames   = map[string]string{"ID": "Cluster ID", "URL": "Endpoint URL"}
	listFields = []string{"ID", "Name", "Status"}
)

// New returns the default command object for interacting with KSQL.
func New(prerunner pcmd.PreRunner, config *config.Config, client ccloud.Connect,
	kafkaClient ccloud.Kafka, userClient ccloud.User, ch *pcmd.ConfigHelper) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "connector",
			Short:             "Manage Kafka Connect.",
			PersistentPreRunE: prerunner.Authenticated(),
		},
		config:      config,
		client:      client,
		kafkaClient: kafkaClient,
		userClient:  userClient,
		ch:          ch,
	}
	cmd.init()
	return cmd.Command
}


func (c *command) init() {
	c.AddCommand(&cobra.Command{
		Use:   "describe",
		Short: "Describe connectors in current Kafka cluster context.",
		RunE:  c.describe,
		Args:  cobra.NoArgs,
	})

	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List connectors in current Kafka cluster context.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})

	createCmd:= &cobra.Command{
		Use:   "create --config <config>",
		Short: "Create connector in the current Kafka cluster context.",
		RunE:  c.create,
		Args:  cobra.ExactArgs(1),
	}
	c.AddCommand(createCmd)
	deleteCmd := &cobra.Command{
		Use:   "delete --connector-id <connector-id>",
		Short: "Delete connector in the current Kafka cluster context.",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	}
	c.AddCommand(deleteCmd)
	updateCmd := &cobra.Command{
		Use:   "update --connector-id <connector-id> --config <config>",
		Short: "Update connector in the current Kafka cluster context.",
		RunE:  c.update,
		Args:  cobra.ExactArgs(2),
	}
	updateCmd.Flags().String("name", "", "New name for Confluent Cloud environment.")
	check(updateCmd.MarkFlagRequired("name"))
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	c.AddCommand(&cobra.Command{
		Use:   "delete <environment-id>",
		Short: "Delete a Confluent Cloud environment and all its resources.",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	})

	getCmd := &cobra.Command{
		Use:   "get ",
		Short: "Get a connector.",
		RunE:  c.get,
		Args:  cobra.ExactArgs(1),
	}
	getCmd.Flags().StringP("output", "o", "", "Output format")
	c.AddCommand(getCmd)

	pauseCmd := &cobra.Command{
		Use:   "get ID",
		Short: "Get a connector.",
		RunE:  c.pause,
		Args:  cobra.ExactArgs(1),
	}
	pauseCmd.Flags().StringP("output", "o", "", "Output format")
	c.AddCommand(pauseCmd)

	resumeCmd := &cobra.Command{
		Use:   "get ID",
		Short: "Get a connector.",
		RunE:  c.resume,
		Args:  cobra.ExactArgs(1),
	}
	resumeCmd.Flags().StringP("output", "o", "", "Output format")
	c.AddCommand(resumeCmd)

	restartCmd := &cobra.Command{
		Use:   "get ID",
		Short: "Get a connector.",
		RunE:  c.restart,
		Args:  cobra.ExactArgs(1),
	}
	restartCmd.Flags().StringP("output", "o", "", "Output format")
	c.AddCommand(restartCmd)
	c.AddCommand(NewTaskCommand(c.config, c.client, c.ch , c.logger))
}



func (c *command) list(cmd *cobra.Command, args []string) error {
	connectors, err := c.client.List(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	var data [][]string
	for _, connector := range connectors {
		data = append(data, printer.ToRow(connector, listFields))
	}
	printer.RenderCollectionTable(data, listFields)
	return nil
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return c.describeById(cmd, args)
	} else {
		return c.describeAll(cmd, args)
	}
}

func (c *command) describeById(cmd *cobra.Command, args []string) error {

	connectorId, err := cmd.Flags().GetString("connector-id")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	connector, err := c.client.Describe(context.Background(), &connectv1.Connector{Id: connectorId, AccountId: c.config.Auth.Account.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	data := &describeDisplay{
		Name:            connector.Name,
		ID:              connector.Id,
		Status:          connector.Status,
		ServiceProvider: serviceProvider,
		Used:            numSchemas,
		Available:       availableSchemas,
		Compatibility:   compatibility,
		Mode:            mode,
	}
	_ = printer.RenderTableOut(data, describeLabels, describeRenames, os.Stdout)
	return nil}

func (c *command) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	_, err := c.client.Create(context.Background(), &orgv1.Account{Name: name, OrganizationId: c.config.Auth.Account.OrganizationId})

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	id := args[0]
	newName := cmd.Flag("name").Value.String()

	err := c.client.Update(context.Background(), &orgv1.Account{Id: id, Name: newName, OrganizationId: c.config.Auth.Account.OrganizationId})

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	id := args[0]

	err := c.client.Delete(context.Background(), &orgv1.Account{Id: id, OrganizationId: c.config.Auth.Account.OrganizationId})

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (c* command) describeAll(cmd *cobra.Command, args []string) error {

	connectors, err := c.clientDescribe(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	var data [][]string
	for _, connector := range connectors {
		data = append(data, printer.ToRow(connector, listFields))
	}
	printer.RenderCollectionTable(data, listFields)
	return nil

}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
