package auth

import (
	"fmt"

	"github.com/confluentinc/cli/internal/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type logoutCommand struct {
	*pcmd.CLICommand
	analyticsClient analytics.Client
}

func NewLogoutCmd(cliName string, prerunner pcmd.PreRunner, analyticsClient analytics.Client) *logoutCommand {
	logoutCmd := &logoutCommand{
		analyticsClient: analyticsClient,
	}
	logoutCmd.init(cliName, prerunner)
	return logoutCmd
}

func (a *logoutCommand) init(cliName string, prerunner pcmd.PreRunner) {
	remoteAPIName := getRemoteAPIName(cliName)
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: fmt.Sprintf("Log out of %s.", remoteAPIName),
		RunE:  pcmd.NewCLIRunE(a.logout),
		Args:  cobra.NoArgs,
		PersistentPreRunE: pcmd.NewCLIPreRunnerE(func(cmd *cobra.Command, args []string) error {
			a.analyticsClient.SetCommandType(analytics.Logout)
			return a.CLICommand.PersistentPreRunE(cmd, args)
		}),
	}
	cliLogoutCmd := pcmd.NewAnonymousCLICommand(logoutCmd, prerunner)
	a.CLICommand = cliLogoutCmd
}

func (a *logoutCommand) logout(cmd *cobra.Command, _ []string) error {
	ctx := a.Config.Config.Context()
	if ctx == nil {
		return nil
	}
	err := ctx.DeleteUserAuth()
	if err != nil {
		return err
	}
	err = a.Config.Save()
	if err != nil {
		return err
	}
	pcmd.Println(cmd, errors.LoggedOutMsg)
	return nil
}
