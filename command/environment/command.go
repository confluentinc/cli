package environment

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	ccloud "github.com/confluentinc/ccloud-sdk-go"
	orgv1 "github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/environment"
	"github.com/confluentinc/go-printer"
)

type command struct {
	*cobra.Command
	config *shared.Config
	client ccloud.Account
}

var (
	listFields = []string{"Id", "Name"}
	listLabels = []string{"Id", "Name"}
)

// New returns the Cobra command for `environment`.
func New(config *shared.Config, factory common.GRPCPluginFactory) (*cobra.Command, error) {
	return newCMD(config, factory.Create(environment.Name))
}

// newCMD returns a command for interacting with `environment`.
func newCMD(config *shared.Config, provider common.GRPCPlugin) (*cobra.Command, error) {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "environment",
			Short: "Manage and select ccloud environments",
		},
		config: config,
	}
	_, err := provider.LookupPath()
	if err != nil {
		return nil, err
	}
	err = cmd.init(provider)
	return cmd.Command, err
}

func (c *command) init(plugin common.GRPCPlugin) error {
	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := common.SetLoggingVerbosity(cmd, c.config.Logger); err != nil {
			return common.HandleError(err, cmd)
		}
		if err := c.config.CheckLogin(); err != nil {
			return common.HandleError(err, cmd)
		}
		// Lazy load plugin to avoid unnecessarily spawning child processes
		return plugin.Load(&c.client, c.config.Logger)
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

	return nil
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	environments, err := c.client.List(context.Background(), &orgv1.Account{})
	if err != nil {
		return common.HandleError(err, cmd)
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
				return common.HandleError(errors.New("Couldn't switch to new environment: couldn't save config."), cmd)
			}
			fmt.Println("Now using", id, "as the default (active) environment.")
			return nil
		}
	}

	return common.HandleError(errors.New("Specified environment ID not found.  Use `ccloud environment list` to see available environments."), cmd)
}
