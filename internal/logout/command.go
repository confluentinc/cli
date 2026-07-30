package logout

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	pauth "github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/auth/sso"
	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type command struct {
	*pcmd.CLICommand
	cfg              *config.Config
	authTokenHandler pauth.AuthTokenHandler
}

func New(cfg *config.Config, prerunner pcmd.PreRunner, authTokenHandler pauth.AuthTokenHandler) *cobra.Command {
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

	c := &command{
		// Anonymous (not Authenticated): logout must not require being logged in, and must not
		// trigger an auto-login via env-var credentials only to immediately log back out.
		CLICommand:       pcmd.NewAnonymousCLICommand(cmd, prerunner),
		cfg:              cfg,
		authTokenHandler: authTokenHandler,
	}

	cmd.Short = fmt.Sprintf("Log out of %s.", context)

	cmd.RunE = c.logout

	return cmd
}

func (c *command) logout(_ *cobra.Command, _ []string) error {
	ctx := c.Config.Context()
	if ctx == nil {
		// Already logged out: do nothing.
		return nil
	}

	if ccloudv2.IsCCloudURL(ctx.Platform.Server, c.cfg.IsTest) {
		if _, err := c.revokeCCloudRefreshToken(ctx); err != nil {
			return err
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

	var userAgent string
	if c.Version != nil {
		userAgent = c.Version.UserAgent
	}
	client := ccloudv1.NewClientWithJWT(context.Background(), contextState.AuthToken, &ccloudv1.Params{
		BaseURL:   ctx.GetPlatformServer(),
		Logger:    log.CliLogger,
		UserAgent: userAgent,
	})

	req := &ccloudv1.AuthenticateRequest{IdToken: contextState.AuthToken}
	if sso.IsOkta(ctx.Platform.Server) {
		return client.Auth.OktaLogout(req)
	} else {
		return client.Auth.Logout(req)
	}
}
