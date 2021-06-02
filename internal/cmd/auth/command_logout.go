package auth

import (
	"fmt"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type logoutCommand struct {
	*pcmd.CLICommand
	cliName         string
	analyticsClient analytics.Client
	netrcHandler    netrc.NetrcHandler
}

func NewLogoutCmd(cliName string, prerunner pcmd.PreRunner, analyticsClient analytics.Client, netrcHandler netrc.NetrcHandler) *logoutCommand {
	logoutCmd := &logoutCommand{
		cliName:         cliName,
		analyticsClient: analyticsClient,
		netrcHandler:    netrcHandler,
	}
	logoutCmd.init(cliName, prerunner)
	return logoutCmd
}

func (a *logoutCommand) init(cliName string, prerunner pcmd.PreRunner) {
	remoteAPIName := getRemoteAPIName(cliName)
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: fmt.Sprintf("Log out of %s.", remoteAPIName),
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

func (a *logoutCommand) logout(cmd *cobra.Command, _ []string) error {
	if a.Config.Config.Context() != nil {
		var level log.Level
		if a.Config.Logger != nil {
			level = a.Config.Logger.GetLevel()
		}
		username, err := a.netrcHandler.RemoveNetrcCredentials(a.Config.CLIName, a.Config.Config.Context().Name)
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
