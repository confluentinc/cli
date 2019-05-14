package secret

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	secret "github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*cobra.Command
	config *config.Config
	prompt pcmd.Prompt
	plugin secret.PasswordProtection
}

// New returns the default command object for Password Protection
func New(prerunner pcmd.PreRunner, config *config.Config, prompt pcmd.Prompt, plugin secret.PasswordProtection) *cobra.Command {
	cmd := &command{
		Command: &cobra.Command{
			Use:               "secret",
			Short:             "Manage secrets for Confluent Platform",
			PersistentPreRunE: prerunner.Authenticated(),
		},
		config: config,
		prompt: prompt,
		plugin: plugin,

	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewMasterKeyCommand(c.config, c.prompt, c.plugin))
	c.AddCommand(NewFileCommand(c.config, c.plugin))
}

