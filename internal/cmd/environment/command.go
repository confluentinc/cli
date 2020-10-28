package environment

import (
	"context"
	"fmt"

	"github.com/c-bata/go-prompt"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	completableChildren []*cobra.Command
}

var (
	listFields             = []string{"Id", "Name"}
	listHumanLabels        = []string{"Id", "Name"}
	listStructuredLabels   = []string{"id", "name"}
	createFields           = []string{"Name", "Id"}
	createHumanLabels      = map[string]string{"Name": "Environment Name"}
	createStructuredLabels = map[string]string{"Name": "name", "Id": "id"}
)

// New returns the Cobra command for `environment`.
func New(cliName string, prerunner pcmd.PreRunner) *command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "environment",
			Short: fmt.Sprintf("Manage and select %s environments.", cliName),
		}, prerunner)
	cmd := &command{AuthenticatedCLICommand: cliCmd}
	cmd.init()
	return cmd
}

func (c *command) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List Confluent Cloud environments.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	useCmd := &cobra.Command{
		Use:   "use <environment-id>",
		Short: "Switch to the specified Confluent Cloud environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.use),
	}
	c.AddCommand(useCmd)

	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new Confluent Cloud environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.create),
	}
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	updateCmd := &cobra.Command{
		Use:   "update <environment-id>",
		Short: "Update an existing Confluent Cloud environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.update),
	}
	updateCmd.Flags().String("name", "", "New name for Confluent Cloud environment.")
	check(updateCmd.MarkFlagRequired("name"))
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <environment-id>",
		Short: "Delete a Confluent Cloud environment and all its resources.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.delete),
	}
	c.AddCommand(deleteCmd)
	c.completableChildren = []*cobra.Command{deleteCmd, updateCmd, useCmd}
}

func (c *command) refreshEnvList() error {
	environments, err := c.Client.Account.List(context.Background(), &orgv1.Account{})
	if err != nil {
		return err
	}
	c.State.Auth.Accounts = environments

	// If current env has gone away, reset active env to 0th env
	hasGoodEnv := false
	if c.State.Auth.Account != nil {
		for _, acc := range c.State.Auth.Accounts {
			if acc.Id == c.EnvironmentId() {
				hasGoodEnv = true
			}
		}
	}
	if !hasGoodEnv {
		c.State.Auth.Account = c.State.Auth.Accounts[0]
	}

	err = c.Config.Save()
	if err != nil {
		return errors.Wrap(err, errors.EnvRefreshErrorMsg)
	}

	return nil
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	environments, err := c.Client.Account.List(context.Background(), &orgv1.Account{})
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}
	for _, environment := range environments {
		// Add '*' only in the case where we are printing out tables
		if outputWriter.GetOutputFormat() == output.Human {
			if environment.Id == c.EnvironmentId() {
				environment.Id = fmt.Sprintf("* %s", environment.Id)
			} else {
				environment.Id = fmt.Sprintf("  %s", environment.Id)
			}
		}
		outputWriter.AddElement(environment)
	}
	return outputWriter.Out()
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	id := args[0]

	acc, err := c.Client.Account.Get(context.Background(), &orgv1.Account{Id: id})
	if err != nil {
		err = errors.NewErrorWithSuggestions(fmt.Sprintf(errors.EnvNotFoundErrorMsg, id), errors.EnvNotFoundSuggestions)
		return err
	}

	c.Context.State.Auth.Account = acc
	if err := c.Config.Save(); err != nil {
		return errors.Wrap(err, errors.EnvSwitchErrorMsg)
	}
	pcmd.Printf(cmd, errors.UsingEnvMsg, id)
	return nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	environment, err := c.Client.Account.Create(context.Background(), &orgv1.Account{Name: name, OrganizationId: c.State.Auth.Account.OrganizationId})
	if err != nil {
		return err
	}
	return output.DescribeObject(cmd, environment, createFields, createHumanLabels, createStructuredLabels)
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	id := args[0]

	newName, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	err = c.Client.Account.Update(context.Background(), &orgv1.Account{Id: id, Name: newName, OrganizationId: c.State.Auth.Account.OrganizationId})

	if err != nil {
		return err
	}
	pcmd.ErrPrintf(cmd, errors.UpdateSuccessMsg, "name", "environment", id, newName)
	return nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	id := args[0]

	err := c.Client.Account.Delete(context.Background(), &orgv1.Account{Id: id, OrganizationId: c.State.Auth.Account.OrganizationId})
	if err != nil {
		return err
	}
	pcmd.ErrPrintf(cmd, errors.DeletedEnvMsg, id)
	return nil
}

func (c *command) Cmd() *cobra.Command {
	return c.Command
}

func (c *command) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}

func (c *command) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	if !pcmd.CanCompleteCommand(c.Command) {
		return suggestions
	}
	environments, err := c.Client.Account.List(context.Background(), &orgv1.Account{})
	if err != nil {
		return suggestions
	}
	for _, env := range environments {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        env.Id,
			Description: env.Name,
		})
	}
	return suggestions
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
