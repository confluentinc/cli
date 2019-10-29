package connector

import (
	"context"
	"fmt"
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
	Name      string
	ID        string
	Status    string
	Used      string
	Available string
	Tasks     string
}

var (
	describeLabels  = []string{"Name", "ID", "Status", "Tasks", "Available", "Used"}
	describeRenames = map[string]string{}
	listFields      = []string{"Id", "Name", "Status"}
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
		Use:   "delete --connector-id <connector-id>",
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

	//
	//getCmd := &cobra.Command{
	//	Use:   "status",
	//	Short: "Get status of a connector.",
	//	RunE:  c.status,
	//	Args:  cobra.ExactArgs(1),
	//}
	//getCmd.Flags().StringP("output", "o", "", "Output format")
	//c.AddCommand(getCmd)
	//
	pauseCmd := &cobra.Command{
		Use:   "pause",
		Short: "Pause a connector.",
		RunE:  c.pause,
		Args:  cobra.ExactArgs(1),
	}
	pauseCmd.Flags().StringP("output", "o", "", "Output format")
	c.AddCommand(pauseCmd)

	resumeCmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume a connector.",
		RunE:  c.resume,
		Args:  cobra.ExactArgs(1),
	}
	resumeCmd.Flags().StringP("output", "o", "", "Output format")
	c.AddCommand(resumeCmd)

	restartCmd := &cobra.Command{
		Use:   "restart",
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
	connectors, err := c.client.List(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	var data [][]string
	for _, connector := range connectors {
		if connector.KafkaClusterId == kafkaCluster.Id {
			data = append(data, printer.ToRow(connector, listFields))
		}
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

	kafkaCluster, err := pcmd.GetKafkaCluster(cmd, c.ch)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if len(args) == 0 {
		return errors.New("Connector ID must be passed")
	}
	connector, err := c.describeFromId(cmd, args[0])
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if connector.KafkaClusterId != kafkaCluster.Id {
		return errors.New("Not found in Kafka cluster context")
	}

	data := &describeDisplay{
		Name:   connector.Name,
		ID:     connector.Id,
		Status: connector.Status.String(),
	}
	_ = printer.RenderTableOut(data, describeLabels, describeRenames, os.Stdout)
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

	connector, err := c.client.Create(context.Background(), &connectv1.ConnectorConfig{UserConfigs: userConfigs, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id})

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	fmt.Print("Created connector" + connector.Id + " " + connector.Name)
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
	connect, err := c.describeFromId(cmd, args[0])
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if connect.KafkaClusterId != kafkaCluster.Id {
		return errors.New("Not found in Kafka cluster context")
	}
	connector, err := c.client.Update(context.Background(), &connectv1.Connector{UserConfigs: userConfigs, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id})

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	fmt.Print("Updated connector" + connector.Id + " " + connector.Name)
	return nil
}

//
//func (c *command) status(cmd *cobra.Command, args []string) error {
//	id := args[0]
//
//	err := c.client.Get(context.Background(), &orgv1.Account{Id: id, Name: newName, OrganizationId: c.config.Auth.Account.OrganizationId})
//
//	if err != nil {
//		return errors.HandleCommon(err, cmd)
//	}
//
//	return nil
//}
//
func (c *command) delete(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := pcmd.GetKafkaCluster(cmd, c.ch)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if len(args) == 0 {
		return errors.New("Connector ID must be passed")
	}
	connector, err := c.describeFromId(cmd, args[0])
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if connector.KafkaClusterId != kafkaCluster.Id {
		return errors.New("Not found in Kafka cluster context")
	}
	err = c.client.Delete(context.Background(), &connectv1.Connector{Name: connector.Name, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id})
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
	connector, err := c.client.Describe(context.Background(), &connectv1.Connector{Id: args[0], AccountId: c.config.Auth.Account.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if connector.KafkaClusterId != kafkaCluster.Id {
		return errors.New("Not found in Kafka cluster context")
	}
	err = c.client.Pause(context.Background(), &connectv1.Connector{Name: connector.Name, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id})
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
	connector, err := c.describeFromId(cmd, args[0])
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if connector.KafkaClusterId != kafkaCluster.Id {
		return errors.New("Not found in Kafka cluster context")
	}

	err = c.client.Resume(context.Background(), &connectv1.Connector{Name: connector.Name, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id})
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
	connector, err := c.describeFromId(cmd, args[0])
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if connector.KafkaClusterId != kafkaCluster.Id {
		return errors.New("Not found in Kafka cluster context")
	}
	// Pause and then Resume
	err = c.client.Pause(context.Background(), &connectv1.Connector{Name: connector.Name, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = c.client.Resume(context.Background(), &connectv1.Connector{Name: connector.Name, AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	pcmd.Println(cmd, "Successfully resumed connector")
	return nil
}

func (c *command) describeAll(cmd *cobra.Command, args []string) error {

	// Get the Kafka Cluster
	kafkaCluster, err := pcmd.GetKafkaCluster(cmd, c.ch)
	if err != nil {
		return err
	}
	connectors, _, err := c.client.ListByKafkaClusterId(context.Background(), &connectv1.Connector{AccountId: c.config.Auth.Account.Id, KafkaClusterId: kafkaCluster.Id}, "info")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	var data [][]string
	for _, connector := range connectors {
		data = append(data, printer.ToRow(&describeDisplay{
			ID: connector.Name,
		}, listFields))
	}
	fmt.Print(len(connectors))
	printer.RenderCollectionTable(data, listFields)
	return nil

}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
