package secret

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/secret"
)

type command struct {
	*pcmd.CLICommand
	flagResolver pcmd.FlagResolver
	plugin       secret.PasswordProtection
}

func New(prerunner pcmd.PreRunner, flagResolver pcmd.FlagResolver, plugin secret.PasswordProtection) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "secret",
		Short:       "Manage secrets for Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonCloudLogin},
	}

	c := &command{
		CLICommand:   pcmd.NewAnonymousCLICommand(cmd, prerunner),
		flagResolver: flagResolver,
		plugin:       plugin,
	}

	cmd.AddCommand(c.newMasterKeyCommand())
	cmd.AddCommand(c.newFileCommand())

	return cmd
}
