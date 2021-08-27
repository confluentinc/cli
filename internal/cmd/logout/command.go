package logout

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type Command struct {
	*pcmd.CLICommand
	cfg             *v3.Config
	analyticsClient analytics.Client
	netrcHandler    netrc.NetrcHandler
}

func New(cfg *v3.Config, prerunner pcmd.PreRunner, analyticsClient analytics.Client, netrcHandler netrc.NetrcHandler) *Command {
	logoutCmd := &Command{
		cfg:             cfg,
		analyticsClient: analyticsClient,
		netrcHandler:    netrcHandler,
	}
	logoutCmd.init(prerunner)
	return logoutCmd
}

func (a *Command) init(prerunner pcmd.PreRunner) {
	context := "Confluent Cloud or Confluent Platform"
	if a.cfg.IsCloudLogin() {
		context = "Confluent Cloud"
	} else if a.cfg.IsOnPremLogin() {
		context = "Confluent Platform"
	}

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: fmt.Sprintf("Log out of %s.", context),
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(a.logout),
		PersistentPreRunE: pcmd.NewCLIPreRunnerE(func(cmd *cobra.Command, args []string) error {
			a.analyticsClient.SetCommandType(analytics.Logout)
			return a.CLICommand.PersistentPreRunE(cmd, args)
		}),
	}
	cliLogoutCmd := pcmd.NewAnonymousCLICommand(logoutCmd, prerunner)
	a.CLICommand = cliLogoutCmd
}

func (a *Command) logout(cmd *cobra.Command, _ []string) error {
	if a.Config.Config.Context() != nil {
		var level log.Level
		if a.Config.Logger != nil {
			level = a.Config.Logger.GetLevel()
		}

		username, err := a.netrcHandler.RemoveNetrcCredentials(a.cfg.IsCloudLogin(), a.Config.Config.Context().Name)
		if err == nil {
			if level >= log.WARN {
				utils.ErrPrintf(cmd, errors.RemoveNetrcCredentialsMsg, username, a.netrcHandler.GetFileName())
			}
		} else if !strings.Contains(err.Error(), "login credentials not found") && !strings.Contains(err.Error(), "keyword expected") {
			// return err when other than NetrcCredentialsNotFoundErrorMsg or parsing error
			return err
		}
	}
	err := pauth.PersistLogoutToConfig(a.Config.Config)
	if err != nil {
		return err
	}
	utils.Println(cmd, errors.LoggedOutMsg)
	return nil
}
