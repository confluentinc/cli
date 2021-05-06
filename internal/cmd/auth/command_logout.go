package auth

import (
	"fmt"

	"github.com/confluentinc/cli/internal/pkg/errors"

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
	err := a.netrcHandler.RemoveNetrcCredentials(a.Config.CLIName, a.Config.Config.Context().Name)
	if err != nil {
		return err
	}

	err = pauth.PersistLogoutToConfig(a.Config.Config)
	if err != nil {
		return err
	}
	utils.Println(cmd, errors.LoggedOutMsg)
	return nil
}
