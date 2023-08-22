package logout

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	pauth "github.com/confluentinc/cli/v3/pkg/auth"
	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/netrc"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type command struct {
	*pcmd.CLICommand
	cfg                      *config.Config
	ccloudClientFactory      pauth.CCloudClientFactory
	netrcHandler             netrc.NetrcHandler
	loginCredentialsManager  pauth.LoginCredentialsManager
	loginOrganizationManager pauth.LoginOrganizationManager
	authTokenHandler         pauth.AuthTokenHandler
}

func New(cfg *config.Config, prerunner pcmd.PreRunner, ccloudClientFactory pauth.CCloudClientFactory, netrcHandler netrc.NetrcHandler, loginCredentialsManager pauth.LoginCredentialsManager, loginOrganizationManager pauth.LoginOrganizationManager, authTokenHandler pauth.AuthTokenHandler) *cobra.Command {
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
		CLICommand:               pcmd.NewAnonymousCLICommand(cmd, prerunner),
		cfg:                      cfg,
		ccloudClientFactory:      ccloudClientFactory,
		netrcHandler:             netrcHandler,
		loginCredentialsManager:  loginCredentialsManager,
		loginOrganizationManager: loginOrganizationManager,
		authTokenHandler:         authTokenHandler,
	}
	cmd.RunE = c.logout

	return cmd
}

func (c *command) logout(cmd *cobra.Command, _ []string) error {
	ctx := c.Config.Config.Context()
	if ctx != nil {
		username, err := c.netrcHandler.RemoveNetrcCredentials(c.cfg.IsCloudLogin(), c.Config.Config.Context().GetNetrcMachineName())
		if err == nil {
			log.CliLogger.Warnf(errors.RemoveNetrcCredentialsMsg, username, c.netrcHandler.GetFileName())
		} else if !strings.Contains(err.Error(), "login credentials not found") && !strings.Contains(err.Error(), "keyword expected") {
			// return err when other than NetrcCredentialsNotFoundErrorMsg or parsing error
			return err
		}

		url := ctx.Platform.Server
		if isCCloud := ccloudv2.IsCCloudURL(url, c.cfg.IsTest); isCCloud {
			if err := c.revokeCCloudRefreshToken(cmd, url); err != nil {
				return err
			}
		}
	}

	if err := pauth.PersistLogout(c.Config.Config); err != nil {
		return err
	}

	output.Println(errors.LoggedOutMsg)
	return nil
}

func (c *command) revokeCCloudRefreshToken(cmd *cobra.Command, url string) error {
	ctx := c.Config.Config.Context()
	orgResourceId := pauth.GetOrgResourceId(cmd, c.loginOrganizationManager)
	credentials, err := pauth.GetCCloudCredentials(c.ccloudClientFactory, cmd, c.cfg, c.loginCredentialsManager, orgResourceId, url)
	if err != nil {
		return err
	}

	contextState := c.Config.Config.ContextStates[ctx.Name]
	if err := contextState.DecryptContextStateAuthToken(ctx.Name); err != nil {
		return err
	}
	credentials.AuthToken = contextState.AuthToken

	if err := c.authTokenHandler.RevokeRefreshToken(c.ccloudClientFactory, url, credentials); err != nil {
		return err
	}
	return nil
}
