package user

import (
	"bufio"
	"context"
	"fmt"
	"github.com/codyaray/go-printer"
	"github.com/confluentinc/cli/shared/user"
	"strings"

	"os"

	"github.com/spf13/cobra"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

type command struct {
	*cobra.Command
	config *shared.Config
	client chttp.User
}

var (
	accountFields = []string{"Id", "ServiceName", "ServiceDescription", "OrganizationId"}
	displayFields = map[string]string{"ServiceName": "Name", "ServiceDescription": "Description"}
)

const nameLength = 32
const descriptionLength = 128

// grpcLoader is the default client loader for the CLI
func grpcLoader(i interface{}) error {
	return common.LoadPlugin(user.Name, i)
}

// New returns the Cobra command for Users.
func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "service-accounts",
			Short: "Manage service accounts.",
		},
		config: config,
	}
	err := cmd.init(grpcLoader)
	return cmd.Command, err
}

func (c *command) init(plugin common.Provider) error {
	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := c.config.CheckLogin(); err != nil {
			fmt.Printf("failed initial login check \n\n%+v\n", c.config)
			return err
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
	check(createCmd.MarkFlagRequired("name"))
	check(createCmd.MarkFlagRequired("description"))
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
	check(updateCmd.MarkFlagRequired("name"))
	check(updateCmd.MarkFlagRequired("description"))
	c.AddCommand(updateCmd)

	deactivateCmd := &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivate a Service Account.",
		RunE:  c.deactivate,
		Args:  cobra.NoArgs,
	}
	deactivateCmd.Flags().String("name", "", "service account name")
	check(deactivateCmd.MarkFlagRequired("name"))
	c.AddCommand(deactivateCmd)

	return nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return common.HandleError(err, cmd)
	}
	namelen := len(name)

	if namelen > nameLength {
		return fmt.Errorf("service name length should be less then 32 bytes.")
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	descriptionlen := len(description)

	if descriptionlen > descriptionLength {
		return fmt.Errorf("description length should be less then 128 bytes.")
	}

	user := &orgv1.User{
		Id:                 c.config.Auth.User.Id,
		ServiceName:        name,
		ServiceDescription: description,
		OrganizationId:     c.config.Auth.User.OrganizationId,
		ServiceAccount:     true,
	}

	user, errRet := c.client.CreateServiceAccount(context.Background(), user)

	if errRet != nil {
		return common.HandleError(errRet, cmd)
	}

	return printer.RenderTableOut(user, accountFields, displayFields, os.Stdout)
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

	user := &orgv1.User{
		Id:                 c.config.Auth.User.Id,
		ServiceName:        name,
		ServiceDescription: description,
		OrganizationId:     c.config.Auth.User.OrganizationId,
	}

	errRet := c.client.UpdateServiceAccount(context.Background(), user)

	if errRet != nil {
		return common.HandleError(errRet, cmd)
	}

	return nil

}

func (c *command) deactivate(cmd *cobra.Command, args []string) error {

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	user := &orgv1.User{
		Id:             c.config.Auth.User.Id,
		ServiceName:    name,
		OrganizationId: c.config.Auth.User.OrganizationId,
	}

	delAcl, err := deleteACL(name)

	if err != nil {
		return common.HandleError(err, cmd)
	}

	if delAcl {
		_ = &kafkav1.ACLFilter{
			EntryFilter: &kafkav1.AccessControlEntryConfig{Principal: user.ServiceName},
		}
		// TODO: get actual delete ACL logic in here
		// c.client.DeleteACL(context.Background, del)
	}

	errRet := c.client.DeactivateServiceAccount(context.Background(), user)

	if errRet != nil {
		return common.HandleError(errRet, cmd)
	}

	return nil

}

func (c *command) list(cmd *cobra.Command, args []string) error {
	user := &orgv1.User{
		Id:             c.config.Auth.User.Id,
		OrganizationId: c.config.Auth.User.OrganizationId,
	}

	users, errRet := c.client.GetServiceAccounts(context.Background(), user)

	if errRet != nil {
		return common.HandleError(errRet, cmd)
	}

	var data [][]string
	for _, user := range users {
		data = append(data, printer.ToRow(user, accountFields))
	}

	printer.RenderCollectionTable(data, accountFields)

	return nil

}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func deleteACL(serviceName string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Do you want to delete all the ACLs assoicated with the %s account? [N/Y] ", serviceName)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	r := strings.TrimSpace(response)
	return r == "" || r[0] == 'y' || r[0] == 'Y', nil
}
