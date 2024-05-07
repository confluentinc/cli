package logout

import (
	"fmt"

	"github.com/spf13/cobra"

	pauth "github.com/confluentinc/cli/v3/pkg/auth"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type command struct {
	*pcmd.CLICommand
	cfg *config.Config
}

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "logout",
		Args: cobra.NoArgs,
	}

	context := "Confluent Cloud or Confluent Platform"
	if cfg.IsCloudLogin() {
		context = "Confluent Cloud"
	} else if cfg.IsOnPremLogin() {
		context = "Confluent Platform"
	}

	cmd.Short = fmt.Sprintf("Log out of %s.", context)

	c := &command{
		CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner),
		cfg:        cfg,
	}
	cmd.RunE = c.logout

	return cmd
}

func (c *command) logout(_ *cobra.Command, _ []string) error {
	if err := pauth.PersistLogout(c.Config); err != nil {
		return err
	}

	output.Println(c.Config.EnableColor, "You are now logged out.")
	return nil
}
