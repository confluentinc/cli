package logout

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"

	climock "github.com/confluentinc/cli/v4/mock"
	pauth "github.com/confluentinc/cli/v4/pkg/auth"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

const (
	testToken      = "y0ur.jwt.T0kEn"
	promptUser     = "prompt-user@confluent.io"
	promptPassword = " prompt-password "
	ccloudURL      = "https://confluent.cloud"
)

func newTestAuthTokenHandler(ctrl *gomock.Controller) *climock.MockAuthTokenHandler {
	authTokenHandler := climock.NewMockAuthTokenHandler(ctrl)
	authTokenHandler.EXPECT().GetCCloudTokens(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(testToken, "refreshToken", nil).AnyTimes()
	authTokenHandler.EXPECT().GetConfluentToken(gomock.Any(), gomock.Any(), gomock.Any()).Return(testToken, "", nil).AnyTimes()
	return authTokenHandler
}

func TestLogout(t *testing.T) {
	req := require.New(t)
	cfg := config.AuthenticatedConfigMockWithContextName(config.MockContextName)
	contextName := cfg.Context().Name
	ctrl := gomock.NewController(t)
	logoutCmd, cfg := newLogoutCmd(ctrl, getAuthMock(), nil, true, req, newTestAuthTokenHandler(ctrl), contextName)
	_, err := pcmd.ExecuteCommand(logoutCmd)
	req.NoError(err)
	verifyLoggedOutState(t, cfg, contextName)
}

func TestRevokeCCloudRefreshTokenNoopWhenContextStateMissing(t *testing.T) {
	req := require.New(t)
	cfg := config.AuthenticatedConfigMockWithContextName(config.MockContextName)
	ctx := cfg.Context()
	// Simulate a missing/corrupt context_state entry for an otherwise-valid context.
	delete(cfg.ContextStates, ctx.Name)

	c := &command{CLICommand: &pcmd.CLICommand{Config: cfg}, cfg: cfg}

	reply, err := c.revokeCCloudRefreshToken(ctx)
	req.NoError(err)
	req.Nil(reply)
}

func TestLogoutNoopWhenAlreadyLoggedOut(t *testing.T) {
	req := require.New(t)
	cfg := config.New()
	prerunner := climock.NewPreRunnerMock(nil, nil, nil, nil, cfg)
	// Logout returns before touching the auth token handler when already logged out.
	logoutCmd := New(cfg, prerunner, nil)

	_, err := pcmd.ExecuteCommand(logoutCmd)
	req.NoError(err)
}

func TestLogoutWithoutSessionCleansCredentials(t *testing.T) {
	for _, test := range []struct {
		name      string
		newConfig func() *config.Config
	}{
		{name: "cloud", newConfig: config.AuthenticatedCloudConfigMock},
		{name: "platform", newConfig: config.AuthenticatedOnPremConfigMock},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)
			cfg := test.newConfig()
			cfg.Filename = filepath.Join(t.TempDir(), "config.json")
			ctx := cfg.Context()
			ctx.State.AuthToken = ""
			ctx.State.AuthRefreshToken = "stale-refresh-token"
			cfg.SavedCredentials[ctx.Name] = &config.LoginCredential{Username: promptUser}
			prerunner := climock.NewPreRunnerMock(nil, nil, nil, nil, cfg)

			output, err := pcmd.ExecuteCommand(New(cfg, prerunner, nil))

			req.NoError(err)
			req.Empty(output)
			req.Empty(cfg.CurrentContext)
			req.NotContains(cfg.SavedCredentials, ctx.Name)
			req.Same(ctx, cfg.Contexts[ctx.Name])
			verifyLoggedOutState(t, cfg, ctx.Name)
			req.Empty(ctx.State.AuthRefreshToken)
		})
	}
}

func TestLogoutAPIKeyContext(t *testing.T) {
	req := require.New(t)
	cfg := config.New()
	cfg.IsTest = true
	cfg.Filename = filepath.Join(t.TempDir(), "config.json")
	req.NoError(cfg.CreateContext("api-key-context", "http://127.0.0.1:1024", "test-key", "test-secret"))
	req.NoError(cfg.UseContext("api-key-context"))
	ctx := cfg.Context()
	prerunner := climock.NewPreRunnerMock(nil, nil, nil, nil, cfg)

	output, err := pcmd.ExecuteCommand(New(cfg, prerunner, nil))

	req.NoError(err)
	req.Equal("You are now logged out.\n", output)
	req.Empty(cfg.CurrentContext)
	req.Same(ctx, cfg.Contexts[ctx.Name])
	req.Equal("test-key", ctx.Credential.APIKeyPair.Key)
	req.Equal("test-secret", ctx.Credential.APIKeyPair.Secret)
}

func newLogoutCmd(ctrl *gomock.Controller, auth *ccloudv1mock.Auth, userInterface *ccloudv1mock.UserInterface, isCloud bool, req *require.Assertions, authTokenHandler pauth.AuthTokenHandler, contextName string) (*cobra.Command, *config.Config) {
	config.SetTempHomeDir()
	cfg := config.AuthenticatedConfigMockWithContextName(contextName)
	var prerunner pcmd.PreRunner

	if !isCloud {
		mdsClient := climock.NewMdsClientMock(testToken)
		prerunner = climock.NewPreRunnerMock(nil, nil, mdsClient, nil, cfg)
	} else {
		ccloudClientFactory := climock.NewCCloudClientFactoryMock(ctrl, auth, userInterface, req)
		prerunner = climock.NewPreRunnerMock(ccloudClientFactory.AnonHTTPClientFactory(ccloudURL), nil, nil, nil, cfg)
	}
	logoutCmd := New(cfg, prerunner, authTokenHandler)
	return logoutCmd, cfg
}

func verifyLoggedOutState(t *testing.T, cfg *config.Config, loggedOutContext string) {
	req := require.New(t)
	state := cfg.Contexts[loggedOutContext].State
	req.Empty(state.AuthToken)
	req.Empty(state.Auth)
}

func getAuthMock() *ccloudv1mock.Auth {
	return &ccloudv1mock.Auth{
		LoginFunc: func(_ *ccloudv1.AuthenticateRequest) (*ccloudv1.AuthenticateReply, error) {
			return &ccloudv1.AuthenticateReply{Token: testToken}, nil
		},
		UserFunc: func() (*ccloudv1.GetMeReply, error) {
			return &ccloudv1.GetMeReply{
				User: &ccloudv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Organization: &ccloudv1.Organization{ResourceId: "o-123"},
				Accounts:     []*ccloudv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
}
