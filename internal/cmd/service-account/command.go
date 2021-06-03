package service_account

import (
	"context"
	"fmt"
	"strconv"

	"github.com/c-bata/go-prompt"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	completableChildren []*cobra.Command
	analyticsClient     analytics.Client
}

var (
	listFields                = []string{"Id", "ResourceId", "ServiceName", "ServiceDescription"}
	listHumanLabels           = []string{"Id", "Resource ID", "Name", "Description"}
	listStructuredLabels      = []string{"id", "resource_id", "name", "description"}
	describeFields            = []string{"Id", "ResourceId", "ServiceName", "ServiceDescription"}
	describeHumanRenames      = map[string]string{"ServiceName": "Name", "ServiceDescription": "Description", "ResourceId": "Resource ID"}
	describeStructuredRenames = map[string]string{"ServiceName": "name", "ServiceDescription": "description", "ResourceId": "resource_id"}
)

const nameLength = 64
const descriptionLength = 128

// New returns the Cobra command for service accounts.
func New(prerunner pcmd.PreRunner, analyticsClient analytics.Client) *command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "service-account",
			Short: `Manage service accounts.`,
		}, prerunner)
	cmd := &command{
		AuthenticatedCLICommand: cliCmd,
		analyticsClient:         analyticsClient,
	}
	cmd.init()
	return cmd
}

func (c *command) Cmd() *cobra.Command {
	return c.Command
}

func (c *command) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return suggestions
	}

	for _, user := range users {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        fmt.Sprintf("%s", user.ResourceId),
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
				Text: "Update the description of a service account with resource ID ``sa-lqv3mm``",
				Code: `ccloud service-account update sa-lqv3mm --description "Update demo service account information."`,
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
				Text: "Delete a service account with resource ID ``sa-lqv3mm``",
				Code: "ccloud service-account delete sa-lqv3mm",
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
	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, user.Id)
	return output.DescribeObject(cmd, user, describeFields, describeHumanRenames, describeStructuredRenames)
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	if err := requireLen(description, descriptionLength, "description"); err != nil {
		return err
	}

	idp, err := strconv.Atoi(args[0])
	user := &orgv1.User{
		ServiceDescription: description,
	}
	if err == nil { // it's a numeric ID
		user.Id = int32(idp)
	} else { // it's a resource ID
		user.ResourceId = args[0]
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
	user := &orgv1.User{}
	if err == nil { // it's a numeric ID
		user.Id = int32(idp)
	} else { // it's a resource ID
		user.ResourceId = args[0]
	}
	err = c.Client.User.DeleteServiceAccount(context.Background(), user)
	if err != nil {
		return err
	}
	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, user.Id)
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
