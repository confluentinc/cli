package secret

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/secret"
)

type command struct {
	*pcmd.CLICommand
	plugin secret.PasswordProtection
}

func New(prerunner pcmd.PreRunner, plugin secret.PasswordProtection) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "secret",
		Short:       "Manage secrets for Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	c := &command{
		CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner),
		plugin:     plugin,
	}

	cmd.AddCommand(c.newMasterKeyCommand())
	cmd.AddCommand(c.newFileCommand())

	return cmd
}
