package logout

import (
	"fmt"
	"net/http"
	"net/url"

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
			if err := c.revokeCCloudSession(ctx); err != nil {
				// Local credentials are cleared regardless, so a failed revocation
				// cannot strand the user in a logged-in state.
				log.CliLogger.Warnf("Failed to revoke session: %v", err)
				output.ErrPrintln(c.Config.EnableColor, "Warning: your session could not be revoked and may still be active. Local credentials were removed.")
			}
		}
	}

	if err := pauth.PersistLogout(c.Config); err != nil {
		return err
	}

	output.Println(c.Config.EnableColor, "You are now logged out.")
	return nil
}

func (c *command) revokeCCloudSession(ctx *config.Context) error {
	if sso.IsOkta(ctx.Platform.Server) {
		contextState := c.Config.ContextStates[ctx.Name]
		if err := contextState.DecryptAuthToken(ctx.Name); err != nil {
			return err
		}

		_, err := c.Client.Auth.OktaLogout(&ccloudv1.AuthenticateRequest{IdToken: contextState.AuthToken})
		return err
	}

	return c.deleteSession()
}

// deleteSession scopes revocation to the CLI's own Auth0 client, leaving the user's
// browser and IDE sessions intact.
func (c *command) deleteSession() error {
	u, err := url.Parse(c.Client.BaseURL)
	if err != nil {
		return err
	}
	u = u.JoinPath("api", "iam", "v2", "sessions")
	u.RawQuery = url.Values{"client_id": []string{sso.GetAuth0CCloudClientIdFromBaseUrl(c.Client.BaseURL)}}.Encode()

	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.Client.UserAgent)

	res, err := c.Client.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("%s returned %s", req.URL.Path, res.Status)
	}

	return nil
}
