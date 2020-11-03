package auth

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
)

// New returns a list of auth-related Cobra commands.
func New(cliName string, prerunner pcmd.PreRunner, logger *log.Logger, ccloudClientFactory pauth.CCloudClientFactory, mdsClientManager pauth.MDSClientManager,
	analyticsClient analytics.Client, netrcHandler netrc.NetrcHandler, loginTokenHandler pauth.LoginTokenHandler) []*cobra.Command {
	loginCmd := NewLoginCommand(cliName, prerunner, logger,
		ccloudClientFactory, mdsClientManager,
		analyticsClient, netrcHandler, loginTokenHandler,
	)
	logoutCmd := NewLogoutCmd(cliName, prerunner, analyticsClient)
	return []*cobra.Command{loginCmd.Command, logoutCmd.Command}
}
