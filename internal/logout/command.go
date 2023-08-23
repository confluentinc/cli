package logout

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

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
	*pcmd.AuthenticatedCLICommand
	cfg              *config.Config
	netrcHandler     netrc.NetrcHandler
	authTokenHandler pauth.AuthTokenHandler
}

func New(cfg *config.Config, prerunner pcmd.PreRunner, netrcHandler netrc.NetrcHandler, authTokenHandler pauth.AuthTokenHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "logout",
		Args: cobra.NoArgs,
	}

	context := "Confluent Cloud or Confluent Platform"
	c := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
		cfg:                     cfg,
		netrcHandler:            netrcHandler,
		authTokenHandler:        authTokenHandler,
	}
	if cfg.IsCloudLogin() {
		context = "Confluent Cloud"
	} else if cfg.IsOnPremLogin() {
		context = "Confluent Platform"
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
	}

	cmd.Short = fmt.Sprintf("Log out of %s.", context)

	cmd.RunE = c.logout

	return cmd
}

func (c *command) logout(_ *cobra.Command, _ []string) error {
	ctx := c.Config.Config.Context()
	if ctx != nil {
		username, err := c.netrcHandler.RemoveNetrcCredentials(c.cfg.IsCloudLogin(), c.Config.Config.Context().GetNetrcMachineName())
		if err == nil {
			log.CliLogger.Warnf(errors.RemoveNetrcCredentialsMsg, username, c.netrcHandler.GetFileName())
		} else if !strings.Contains(err.Error(), "login credentials not found") && !strings.Contains(err.Error(), "keyword expected") {
			// return err when other than NetrcCredentialsNotFoundErrorMsg or parsing error
			return err
		}

		url := c.Client.BaseURL
		if isCCloud := ccloudv2.IsCCloudURL(url, c.cfg.IsTest); isCCloud {
			if err := c.revokeCCloudRefreshToken(); err != nil {
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

func (c *command) revokeCCloudRefreshToken() error {
	ctx := c.Config.Config.Context()

	contextState := c.Config.Config.ContextStates[ctx.Name]
	if err := contextState.DecryptContextStateAuthToken(ctx.Name); err != nil {
		return err
	}
	req := &ccloudv1.AuthenticateRequest{IdToken: contextState.AuthToken}

	if _, err := c.authTokenHandler.RevokeRefreshToken(c.Client, req); err != nil {
		return err
	}
	return nil
}
