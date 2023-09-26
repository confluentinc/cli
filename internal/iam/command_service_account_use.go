package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *serviceAccountCommand) newUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <id>",
		Short:             "Choose a service account to be used in subsequent commands.",
		Long:              "Choose a service account to be used in subsequent commands which support passing a service account with the `--service-account` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.use,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *serviceAccountCommand) use(_ *cobra.Command, args []string) error {
	id := args[0]
	if _, _, err := c.V2Client.GetIamServiceAccount(id); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), "List available service accounts with `confluent iam service-account list`.")
	}

	if err := c.Context.SetCurrentServiceAccount(id); err != nil {
		return err
	}
	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(errors.UsingResourceMsg, resource.ServiceAccount, args[0])
	return nil
}
