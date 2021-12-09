package iam

import (
	"context"
	"fmt"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type serviceAccountCommand struct {
	*pcmd.AuthenticatedCLICommand
	completableChildren []*cobra.Command
}

func NewServiceAccountCommand(prerunner pcmd.PreRunner) *serviceAccountCommand {
	cmd := &cobra.Command{
		Use:         "service-account",
		Aliases:     []string{"sa"},
		Short:       "Manage service accounts.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &serviceAccountCommand{AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	deleteCmd := c.newDeleteCommand()
	updateCmd := c.newUpdateCommand()

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(deleteCmd)
	c.AddCommand(c.newListCommand())
	c.AddCommand(updateCmd)

	c.completableChildren = []*cobra.Command{updateCmd, deleteCmd}

	return c
}

func (c *serviceAccountCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteServiceAccounts(c.Client)
}

func (c *serviceAccountCommand) Cmd() *cobra.Command {
	return c.Command
}

func (c *serviceAccountCommand) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	if c.Client == nil {
		return suggestions
	}
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

func requireLen(val string, maxLen int, field string) error {
	if len(val) > maxLen {
		return fmt.Errorf(field+" length should not exceed %d characters.", maxLen)
	}

	return nil
}
