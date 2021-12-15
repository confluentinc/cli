package logout

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type Command struct {
	*pcmd.CLICommand
	cfg             *v1.Config
	analyticsClient analytics.Client
	netrcHandler    netrc.NetrcHandler
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner, analyticsClient analytics.Client, netrcHandler netrc.NetrcHandler) *Command {
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
		CLICommand:      pcmd.NewAnonymousCLICommand(cmd, prerunner),
		cfg:             cfg,
		analyticsClient: analyticsClient,
		netrcHandler:    netrcHandler,
	}

	cmd.RunE = pcmd.NewCLIRunE(c.logout)

	return c
}

func (c *Command) logout(cmd *cobra.Command, _ []string) error {
	if c.Config.Config.Context() != nil {
		var level log.Level
		if c.Config.Logger != nil {
			level = c.Config.Logger.GetLevel()
		}

		username, err := c.netrcHandler.RemoveNetrcCredentials(c.cfg.IsCloudLogin(), c.Config.Config.Context().NetrcMachineName)
		if err == nil {
			if level >= log.WARN {
				utils.ErrPrintf(cmd, errors.RemoveNetrcCredentialsMsg, username, c.netrcHandler.GetFileName())
			}
		} else if !strings.Contains(err.Error(), "login credentials not found") && !strings.Contains(err.Error(), "keyword expected") {
			// return err when other than NetrcCredentialsNotFoundErrorMsg or parsing error
			return err
		}
	}

	if err := pauth.PersistLogoutToConfig(c.Config.Config); err != nil {
		return err
	}

	utils.Println(cmd, errors.LoggedOutMsg)
	return nil
}
