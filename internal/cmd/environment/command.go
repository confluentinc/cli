package environment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

var (
	listFields           = []string{"Id", "Name"}
	listHumanLabels      = []string{"Id", "Name"}
	listStructuredLabels = []string{"id", "name"}
	createFields         = []string{"Name", "Id"}
	createHumanLabels    = map[string]string{"Name": "Environment Name"}
	createStructuredLabels = map[string]string{"Name": "name", "Id": "id"}
)

// New returns the Cobra command for `environment`.
func New(prerunner pcmd.PreRunner, config *v2.Config, cliName string) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "environment",
			Short: fmt.Sprintf("Manage and select %s environments.", cliName),
		},
		config, prerunner)
	cmd := &command{AuthenticatedCLICommand: cliCmd}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List Confluent Cloud environments.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	c.AddCommand(&cobra.Command{
		Use:   "use <environment-id>",
		Short: "Switch to the specified Confluent Cloud environment.",
		RunE:  c.use,
		Args:  cobra.ExactArgs(1),
	})

	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new Confluent Cloud environment.",
		RunE:  c.create,
		Args:  cobra.ExactArgs(1),
	}
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	updateCmd := &cobra.Command{
		Use:   "update <environment-id>",
		Short: "Update an existing Confluent Cloud environment.",
		RunE:  c.update,
		Args:  cobra.ExactArgs(1),
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
}

func (c *command) refreshEnvList(cmd *cobra.Command) error {
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
		return errors.Wrap(err, "unable to save user auth while refreshing environment list")
	}

	return nil
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	environments, err := c.Client.Account.List(context.Background(), &orgv1.Account{})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return errors.HandleCommon(err, cmd)
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
		return errors.HandleCommon(errors.New("The specified environment ID was not found.  To see available environments, use `ccloud environment list`."), cmd)
	}

	c.Context.State.Auth.Account = acc
	if err := c.Config.Save(); err != nil {
		return errors.HandleCommon(errors.New("couldn't switch to new environment: couldn't save config."), cmd)
	}
	pcmd.Println(cmd, "Now using", id, "as the default (active) environment.")
	return nil
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	environment, err := c.Client.Account.Create(context.Background(), &orgv1.Account{Name: name, OrganizationId: c.State.Auth.Account.OrganizationId})
	if err != nil {
		return errors.HandleCommon(err, cmd)
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
		return errors.HandleCommon(err, cmd)
	}
	return nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	id := args[0]

	err := c.Client.Account.Delete(context.Background(), &orgv1.Account{Id: id, OrganizationId: c.State.Auth.Account.OrganizationId})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
