package secret

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/secret"
)

type command struct {
	*pcmd.CLICommand
	resolv pcmd.FlagResolver
	plugin secret.PasswordProtection
}

// New returns the default command object for Password Protection
func New(prerunner pcmd.PreRunner, resolv pcmd.FlagResolver, plugin secret.PasswordProtection) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "secret",
		Short:       "Manage secrets for Confluent Platform.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &command{
		CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner),
		resolv:     resolv,
		plugin:     plugin,
	}
	c.init()
	return c.Command
}

func (c *command) init() {
	c.AddCommand(NewMasterKeyCommand(c.resolv, c.plugin))
	c.AddCommand(NewFileCommand(c.resolv, c.plugin))
}
