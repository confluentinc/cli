package service_account

import (
	"context"
	"fmt"
	"strconv"

	"github.com/c-bata/go-prompt"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	completableChildren []*cobra.Command
}

var (
	listFields                = []string{"Id", "ServiceName", "ServiceDescription"}
	listHumanLabels           = []string{"Id", "Name", "Description"}
	listStructuredLabels      = []string{"id", "name", "description"}
	describeFields            = []string{"Id", "ServiceName", "ServiceDescription"}
	describeHumanRenames      = map[string]string{"ServiceName": "Name", "ServiceDescription": "Description"}
	describeStructuredRenames = map[string]string{"ServiceName": "name", "ServiceDescription": "description"}
)

const nameLength = 64
const descriptionLength = 128

// New returns the Cobra command for service accounts.
func New(prerunner pcmd.PreRunner) *command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "service-account",
			Short: `Manage service accounts.`,
		}, prerunner)
	cmd := &command{
		AuthenticatedCLICommand: cliCmd,
	}
	cmd.init()
	return cmd
}

func (c *command) Cmd() *cobra.Command {
	return c.Command
}

func (c *command) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	if !pcmd.CanCompleteCommand(c.Command) {
		return suggestions
	}

	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return suggestions
	}

	for _, user := range users {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        fmt.Sprintf("%d", user.Id),
			Description: fmt.Sprintf("%s: %s", user.ServiceName, user.ServiceDescription),
		})
	}

	return suggestions
}

func (c *command) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}

func (c *command) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List service accounts.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a service account.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a service account named ``DemoServiceAccount``.",
				Code: `ccloud service-account create DemoServiceAccount --description "This is a demo service account."`,
			},
		),
	}
	createCmd.Flags().String("description", "", "Description of the service account.")
	_ = createCmd.MarkFlagRequired("description")
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	updateCmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a service account.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.update),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update the description of a service account with the ID ``2786``",
				Code: `ccloud service-account update 2786 --description "Update demo service account information."`,
			},
		),
	}
	updateCmd.Flags().String("description", "", "Description of the service account.")
	_ = updateCmd.MarkFlagRequired("description")
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a service account.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.delete),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete a service account with the ID ``2786``",
				Code: "ccloud service-account delete 2786",
			},
		),
	}
	c.AddCommand(deleteCmd)
	c.completableChildren = []*cobra.Command{updateCmd, deleteCmd}
}

func requireLen(val string, maxLen int, field string) error {
	if len(val) > maxLen {
		return fmt.Errorf(field+" length should not exceed %d characters.", maxLen)
	}

	return nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := requireLen(name, nameLength, "service name"); err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	if err := requireLen(description, descriptionLength, "description"); err != nil {
		return err
	}

	user := &orgv1.User{
		ServiceName:        name,
		ServiceDescription: description,
		OrganizationId:     c.State.Auth.User.OrganizationId,
		ServiceAccount:     true,
	}
	user, err = c.Client.User.CreateServiceAccount(context.Background(), user)
	if err != nil {
		return err
	}
	return output.DescribeObject(cmd, user, describeFields, describeHumanRenames, describeStructuredRenames)
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	idp, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	id := int32(idp)

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	if err := requireLen(description, descriptionLength, "description"); err != nil {
		return err
	}

	user := &orgv1.User{
		Id:                 id,
		ServiceDescription: description,
	}
	err = c.Client.User.UpdateServiceAccount(context.Background(), user)
	if err != nil {
		return err
	}
	utils.ErrPrintf(cmd, errors.UpdateSuccessMsg, "description", "service account", args[0], description)
	return nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	idp, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}
	id := int32(idp)

	user := &orgv1.User{
		Id: id,
	}
	err = c.Client.User.DeleteServiceAccount(context.Background(), user)
	if err != nil {
		return err
	}
	return nil
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}
	for _, u := range users {
		outputWriter.AddElement(u)
	}
	return outputWriter.Out()
}
