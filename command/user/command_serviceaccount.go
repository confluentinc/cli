package user

import (
	"bufio"
	"fmt"
	"context"
	"github.com/codyaray/go-printer"
	"strings"

	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
)

type command struct {
	*cobra.Command
	config *shared.Config
	user  User
}

var (
	accountFields  = []string{"Id", "ServiceName", "ServiceDescription", "OrganizationId"}
	displayFields = map[string]string{"ServiceName": "Name", "ServiceDescription": "Description"}
)

// New returns the Cobra command for Users.
func New(config *shared.Config) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "service-accounts",
			Short: "Manage service accounts.",
		},
		config: config,
	}
	err := cmd.init()
	return cmd.Command, err
}

func (c *command) init() error {
	path, err := exec.LookPath("confluent-user-plugin")
	if err != nil {
		return fmt.Errorf("skipping user: plugin isn't installed")
	}

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Plugins:          shared.PluginMap,
		Cmd:              exec.Command("sh", "-c", path), // nolint: gas
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Managed:          true,
		Logger: hclog.New(&hclog.LoggerOptions{
			Output: hclog.DefaultOutput,
			Level:  hclog.Info,
			Name:   "plugin",
		}),
	})

	// Connect via RPC.
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("user")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Got a client now communicating over RPC.
	c.user = raw.(User)

	// All commands require login first
	c.Command.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if err = c.config.CheckLogin(); err != nil {
			_ = common.HandleError(err, cmd)
			os.Exit(1)
		}
	}

	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List Service Accounts.",
		RunE:   c.list,
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
		Args: cobra.NoArgs,
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
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return common.HandleError(err, cmd)
	}

	user := &orgv1.User{
		Id:                 c.config.Auth.User.Id,
		ServiceName:        name,
		ServiceDescription: description,
		OrganizationId:     c.config.Auth.User.OrganizationId,
		ServiceAccount:     true,
	}

	user, errRet := c.user.CreateServiceAccount(context.Background(), user)

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

	errRet := c.user.UpdateServiceAccount(context.Background(), user)

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
		Id:                 c.config.Auth.User.Id,
		ServiceName:        name,
		OrganizationId:     c.config.Auth.User.OrganizationId,
	}

	delAcl, err := deleteACL(name)

	if err != nil {
		return common.HandleError(err, cmd)
	}

	if delAcl {
		// To-do Call DeleteACL API
	}

	errRet := c.user.DeactivateServiceAccount(context.Background(), user)

	if errRet != nil {
		return common.HandleError(errRet, cmd)
	}

	return nil

}

func (c *command) list(cmd *cobra.Command, args []string) error {
	user := &orgv1.User{
		Id:                 c.config.Auth.User.Id,
		OrganizationId:     c.config.Auth.User.OrganizationId,
	}

	users, errRet := c.user.GetServiceAccounts(context.Background(), user)

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
