package logout

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	pauth "github.com/confluentinc/cli/v3/pkg/auth"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/netrc"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type Command struct {
	*pcmd.CLICommand
	cfg          *config.Config
	netrcHandler netrc.NetrcHandler
}

func New(cfg *config.Config, prerunner pcmd.PreRunner, netrcHandler netrc.NetrcHandler) *cobra.Command {
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

	c := &Command{
		CLICommand:   pcmd.NewAnonymousCLICommand(cmd, prerunner),
		cfg:          cfg,
		netrcHandler: netrcHandler,
	}
	cmd.RunE = c.logout

	return cmd
}

func (c *Command) logout(cmd *cobra.Command, _ []string) error {
	if c.Config.Config.Context() != nil {
		username, err := c.netrcHandler.RemoveNetrcCredentials(c.cfg.IsCloudLogin(), c.Config.Config.Context().GetNetrcMachineName())
		if err == nil {
			log.CliLogger.Warnf(errors.RemoveNetrcCredentialsMsg, username, c.netrcHandler.GetFileName())
		} else if !strings.Contains(err.Error(), "login credentials not found") && !strings.Contains(err.Error(), "keyword expected") {
			// return err when other than NetrcCredentialsNotFoundErrorMsg or parsing error
			return err
		}
	}

	if err := pauth.PersistLogout(c.Config.Config); err != nil {
		return err
	}

	output.Println(c.Config.EnableColor, errors.LoggedOutMsg)
	return nil
}
