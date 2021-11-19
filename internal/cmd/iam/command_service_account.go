package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type serviceAccountCommand struct {
	*pcmd.AuthenticatedCLICommand
	completableChildren []*cobra.Command
}

var (
	describeFields            = []string{"ResourceId", "ServiceName", "ServiceDescription"}
	describeHumanRenames      = map[string]string{"ServiceName": "Name", "ServiceDescription": "Description", "ResourceId": "ID"}
	describeStructuredRenames = map[string]string{"ServiceName": "name", "ServiceDescription": "description", "ResourceId": "id"}
)

const nameLength = 64
const descriptionLength = 128

func NewServiceAccountCommand(prerunner pcmd.PreRunner) *serviceAccountCommand {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:         "service-account",
			Short:       `Manage service accounts.`,
			Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		}, prerunner)
	cmd := &serviceAccountCommand{
		AuthenticatedCLICommand: cliCmd,
	}
	cmd.init()
	return cmd
}

func (c *serviceAccountCommand) Cmd() *cobra.Command {
	return c.Command
}

func (c *serviceAccountCommand) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return suggestions
	}

	for _, user := range users {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        user.ResourceId,
			Description: fmt.Sprintf("%s: %s", user.ServiceName, user.ServiceDescription),
		})
	}

	return suggestions
}

func (c *serviceAccountCommand) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}

func (c *serviceAccountCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List service accounts.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a service account.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a service account named `DemoServiceAccount`.",
				Code: `confluent service-account create DemoServiceAccount --description "This is a demo service account."`,
			},
		),
	}
	createCmd.Flags().String("description", "", "Description of the service account.")
	_ = createCmd.MarkFlagRequired("description")
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	c.AddCommand(createCmd)

	updateCmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a service account.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.update),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update the description of service account `sa-lqv3mm`",
				Code: `confluent service-account update sa-lqv3mm --description "Update demo service account information."`,
			},
		),
	}
	updateCmd.Flags().String("description", "", "Description of the service account.")
	_ = updateCmd.MarkFlagRequired("description")
	c.AddCommand(updateCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a service account.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.delete),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete service account `sa-lqv3mm`",
				Code: "confluent service-account delete sa-lqv3mm",
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

func (c *serviceAccountCommand) create(cmd *cobra.Command, args []string) error {
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
		ServiceAccount:     true,
	}
	user, err = c.Client.User.CreateServiceAccount(context.Background(), user)
	if err != nil {
		return err
	}
	return output.DescribeObject(cmd, user, describeFields, describeHumanRenames, describeStructuredRenames)
}

func (c *serviceAccountCommand) update(cmd *cobra.Command, args []string) error {
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	if err := requireLen(description, descriptionLength, "description"); err != nil {
		return err
	}

	if !strings.HasPrefix(args[0], "sa-") {
		return errors.New(errors.BadServiceAccountIDErrorMsg)
	}
	user := &orgv1.User{
		ResourceId:         args[0],
		ServiceDescription: description,
	}

	if err := c.Client.User.UpdateServiceAccount(context.Background(), user); err != nil {
		return err
	}

	utils.ErrPrintf(cmd, errors.UpdateSuccessMsg, "description", "service account", args[0], description)
	return nil
}

func (c *serviceAccountCommand) delete(cmd *cobra.Command, args []string) error {
	if !strings.HasPrefix(args[0], "sa-") {
		return errors.New(errors.BadServiceAccountIDErrorMsg)
	}
	user := &orgv1.User{ResourceId: args[0]}
	if err := c.Client.User.DeleteServiceAccount(context.Background(), user); err != nil {
		return err
	}
	utils.ErrPrintf(cmd, errors.DeletedServiceAccountMsg, args[0])
	return nil
}

func (c *serviceAccountCommand) list(cmd *cobra.Command, _ []string) error {
	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return err
	}

	var (
		listFields           = []string{"ResourceId", "ServiceName", "ServiceDescription"}
		listHumanLabels      = []string{"ID", "Name", "Description"}
		listStructuredLabels = []string{"id", "name", "description"}
	)

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}
	for _, u := range users {
		outputWriter.AddElement(u)
	}
	return outputWriter.Out()
}
