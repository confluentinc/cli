package secret

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/secret"
)

type command struct {
	*cobra.Command
	resolv pcmd.FlagResolver
	plugin secret.PasswordProtection
}

// New returns the default command object for Password Protection
func New(resolv pcmd.FlagResolver, plugin secret.PasswordProtection) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:   "secret",
			Short: "Manage secrets for Confluent Platform.",
		},
		resolv: resolv,
		plugin: plugin,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewMasterKeyCommand(c.resolv, c.plugin))
	c.AddCommand(NewFileCommand(c.resolv, c.plugin))
}
