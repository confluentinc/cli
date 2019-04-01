package environment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/go-printer"
)

type command struct {
	*cobra.Command
	config *config.Config
	client ccloud.Account
}

var (
	listFields = []string{"Id", "Name"}
	listLabels = []string{"Id", "Name"}
)

// New returns the Cobra command for `environment`.
func New(config *config.Config, client ccloud.Account) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "environment",
			Short: "Manage and select ccloud environments",
		},
		config: config,
		client: client,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := log.SetLoggingVerbosity(cmd, c.config.Logger); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		if err := c.config.CheckLogin(); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		return nil
	}

	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List environments",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})

	c.AddCommand(&cobra.Command{
		Use:   "use ID",
		Short: "Switch to the specified environment",
		RunE:  c.use,
		Args:  cobra.ExactArgs(1),
	})

	c.AddCommand(&cobra.Command{
		Use:   "create NAME",
		Short: "Create a new environment",
		RunE:  c.create,
		Args:  cobra.ExactArgs(1),
	})

	updateCmd := &cobra.Command{
		Use:   "update ID --name NEWNAME",
		Short: "Update the name of an environment",
		RunE:  c.update,
		Args:  cobra.ExactArgs(1),
	}
	updateCmd.Flags().String("name", "", "New name for environment")
	c.AddCommand(updateCmd)

	c.AddCommand(&cobra.Command{
		Use:   "delete ID",
		Short: "Delete an environment and ALL its resources",
		RunE:  c.delete,
		Args:  cobra.ExactArgs(1),
	})
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	environments, err := c.client.List(context.Background(), &orgv1.Account{})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	var data [][]string
	for _, environment := range environments {
		if environment.Id == c.config.Auth.Account.Id {
			environment.Id = fmt.Sprintf("* %s", environment.Id)
		} else {
			environment.Id = fmt.Sprintf("  %s", environment.Id)
		}
		data = append(data, printer.ToRow(environment, listFields))
	}
	printer.RenderCollectionTable(data, listLabels)
	return nil
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	id := args[0]

	for _, acc := range c.config.Auth.Accounts {
		if acc.Id == id {
			c.config.Auth.Account = acc
			err := c.config.Save()
			if err != nil {
				return errors.HandleCommon(errors.New("couldn't switch to new environment: couldn't save config."), cmd)
			}
			fmt.Println("Now using", id, "as the default (active) environment.")
			return nil
		}
	}

	return errors.HandleCommon(errors.New("specified environment ID not found.  Use `ccloud environment list` to see available environments."), cmd)
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	_, err := c.client.Create(context.Background(), &orgv1.Account{Name: name, OrganizationId: c.config.Auth.Account.OrganizationId})

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	id := args[0]
	newName := cmd.Flag("name").Value.String()

	err := c.client.Update(context.Background(), &orgv1.Account{Id: id, Name: newName, OrganizationId: c.config.Auth.Account.OrganizationId})

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	id := args[0]

	err := c.client.Delete(context.Background(), &orgv1.Account{Id: id, OrganizationId: c.config.Auth.Account.OrganizationId})

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}
