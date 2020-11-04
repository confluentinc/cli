package auth

import (
	"context"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
)

// New returns a list of auth-related Cobra commands.
func New(cliName string, prerunner pcmd.PreRunner, logger *log.Logger, userAgent string, analyticsClient analytics.Client, netrcHandler netrc.NetrcHandler,
	loginTokenHandler pauth.LoginTokenHandler) []*cobra.Command {
	var defaultAnonHTTPClientFactory = func(baseURL string, logger *log.Logger) *ccloud.Client {
		return ccloud.NewClient(&ccloud.Params{BaseURL: baseURL, HttpClient: ccloud.BaseClient, Logger: logger, UserAgent: userAgent})
	}
	var defaultJwtHTTPClientFactory = func(ctx context.Context, jwt string, baseURL string, logger *log.Logger) *ccloud.Client {
		return ccloud.NewClientWithJWT(ctx, jwt, &ccloud.Params{BaseURL: baseURL, Logger: logger, UserAgent: userAgent})
	}
	loginCmd := NewLoginCommand(cliName, prerunner, logger,
		defaultAnonHTTPClientFactory, defaultJwtHTTPClientFactory, &pauth.MDSClientManagerImpl{},
		analyticsClient, netrcHandler, loginTokenHandler,
	)
	logoutCmd := NewLogoutCmd(cliName, prerunner, analyticsClient)
	return []*cobra.Command{loginCmd.Command, logoutCmd.Command}
}
