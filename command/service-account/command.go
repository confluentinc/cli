package service_account

import (
	"context"
	"fmt"
	"os"

	"github.com/codyaray/go-printer"
	"github.com/spf13/cobra"

	ccloud "github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	sharedUser "github.com/confluentinc/cli/shared/user"
)

type command struct {
	*cobra.Command
	config *shared.Config
	client ccloud.User
}

var (
	listFields      = []string{"Id", "ServiceName", "ServiceDescription"}
	listLabels      = []string{"Id", "Name", "Description"}
	describeFields  = []string{"Id", "ServiceName", "ServiceDescription"}
	describeRenames = map[string]string{"ServiceName": "Name", "ServiceDescription": "Description"}
)

const nameLength = 32
const descriptionLength = 128

// grpcLoader is the default client loader for the CLI
func grpcLoader(i interface{}) error {
	return common.LoadPlugin(sharedUser.Name, i)
}

// New returns the Cobra command for Users.
func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "service-account",
			Short: "Manage service accounts",
		},
		config: config,
	}
	err := cmd.init(grpcLoader)
	return cmd.Command, err
}

func (c *command) init(plugin common.Provider) error {
	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := c.config.CheckLogin(); err != nil {
			return common.HandleError(err, cmd)
		}
		// Lazy load plugin to avoid unnecessarily spawning child processes
		return plugin(&c.client)
	}

	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List Service Accounts.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})

	createCmd := &cobra.Command{
		Use:   "create service account",
		Short: "Create Service Account.",
		RunE:  c.create,
		Args:  cobra.NoArgs,
	}
	createCmd.Flags().String("name", "", "service account name")
	createCmd.Flags().String("description", "", "service account description")
	_ = createCmd.MarkFlagRequired("name")
	_ = createCmd.MarkFlagRequired("description")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	updateCmd := &cobra.Command{
		Use:   "update service account",
		Short: "Update Service Description of a Service Account.",
		RunE:  c.update,
		Args:  cobra.NoArgs,
	}
	updateCmd.Flags().String("name", "", "service account name")
	updateCmd.Flags().String("description", "", "service account description")
	_ = updateCmd.MarkFlagRequired("name")
	_ = updateCmd.MarkFlagRequired("description")
	c.AddCommand(updateCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Service Account.",
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	deleteCmd.Flags().String("name", "", "service account name")
	_ = deleteCmd.MarkFlagRequired("name")
	c.AddCommand(deleteCmd)

	return nil
}

func requireLen(val string, maxLen int, field string) error {
	if len(val) > maxLen {
		return fmt.Errorf(field+" length should be less then %d characters.", maxLen)
	}

	return nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	if err := requireLen(name, nameLength, "service name"); err != nil {
		return common.HandleError(err, cmd)
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	if err := requireLen(description, descriptionLength, "description"); err != nil {
		return common.HandleError(err, cmd)
	}

	user := &orgv1.User{
		ServiceName:        name,
		ServiceDescription: description,
		OrganizationId:     c.config.Auth.User.OrganizationId,
		ServiceAccount:     true,
	}

	user, err = c.client.CreateServiceAccount(context.Background(), user)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	return printer.RenderTableOut(user, describeFields, describeRenames, os.Stdout)
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return common.HandleError(err, cmd)
	}
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	if err := requireLen(description, descriptionLength, "description"); err != nil {
		return common.HandleError(err, cmd)
	}

	user := &orgv1.User{
		ServiceName:        name,
		ServiceDescription: description,
		OrganizationId:     c.config.Auth.User.OrganizationId,
	}

	err = c.client.UpdateServiceAccount(context.Background(), user)
	if err != nil {
		return common.HandleError(err, cmd)
	}
	return nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	user := &orgv1.User{
		ServiceName:    name,
		OrganizationId: c.config.Auth.User.OrganizationId,
	}

	err = c.client.DeleteServiceAccount(context.Background(), user)
	if err != nil {
		return common.HandleError(err, cmd)
	}
	return nil
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	users, err := c.client.GetServiceAccounts(context.Background())
	if err != nil {
		return common.HandleError(err, cmd)
	}

	var data [][]string
	for _, u := range users {
		data = append(data, printer.ToRow(u, listFields))
	}

	printer.RenderCollectionTable(data, listLabels)
	return nil
}
