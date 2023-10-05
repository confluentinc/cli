package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *serviceAccountCommand) newUnsetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "unset",
		Short:             "Unset the current service account.",
		Long:              "Unset the current service account that was set with `confluent iam service-account use`.",
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.unset,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *serviceAccountCommand) unset(_ *cobra.Command, args []string) error {
	serviceAccountToUnset := c.Context.GetCurrentServiceAccount()
	if serviceAccountToUnset == "" {
		return nil
	}

	if err := c.Context.SetCurrentServiceAccount(""); err != nil {
		return err
	}
	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UnsetResourceMsg, resource.ServiceAccount, serviceAccountToUnset)
	return nil
}
