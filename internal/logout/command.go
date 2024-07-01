package logout

import (
	"fmt"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	pauth "github.com/confluentinc/cli/v3/pkg/auth"
	"github.com/confluentinc/cli/v3/pkg/auth/sso"
	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	cfg              *config.Config
	authTokenHandler pauth.AuthTokenHandler
}

func New(cfg *config.Config, prerunner pcmd.PreRunner, authTokenHandler pauth.AuthTokenHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "logout",
		Args: cobra.NoArgs,
	}

	context := "Confluent Cloud or Confluent Platform"
	c := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
		cfg:                     cfg,
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
	ctx := c.Config.Context()
	if ctx != nil {
		if ccloudv2.IsCCloudURL(ctx.Platform.Server, c.cfg.IsTest) {
			if _, err := c.revokeCCloudRefreshToken(ctx); err != nil {
				return err
			}
		}
	}

	if err := pauth.PersistLogout(c.Config); err != nil {
		return err
	}

	output.Println(c.Config.EnableColor, "You are now logged out.")
	return nil
}

func (c *command) revokeCCloudRefreshToken(ctx *config.Context) (*ccloudv1.AuthenticateReply, error) {
	contextState := c.Config.ContextStates[ctx.Name]
	if err := contextState.DecryptAuthToken(ctx.Name); err != nil {
		return nil, err
	}

	req := &ccloudv1.AuthenticateRequest{IdToken: contextState.AuthToken}
	if sso.IsOkta(ctx.Platform.Server) {
		return c.Client.Auth.OktaLogout(req)
	} else {
		return c.Client.Auth.Logout(req)
	}
}
