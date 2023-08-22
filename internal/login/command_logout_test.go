package login

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"
	mdsmock "github.com/confluentinc/mds-sdk-go-public/mdsv1/mock"

	climock "github.com/confluentinc/cli/v3/mock"
	pauth "github.com/confluentinc/cli/v3/pkg/auth"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/netrc"
)

const (
	testToken = "y0ur.jwt.T0kEn"
)

func TestLogout(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	cfg := config.AuthenticatedConfigMockWithContextName(config.MockContextName)
	contextName := cfg.Context().Name
	// run login command
	auth := &ccloudv1mock.Auth{
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
	userInterface := &ccloudv1mock.UserInterface{}
	logoutCmd, cfg := newLogoutCmd(auth, userInterface, true, req, mockNetrcHandler, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager, contextName)
	_, err := pcmd.ExecuteCommand(logoutCmd)
	req.NoError(err)
	exist, err := mockNetrcHandler.CheckCredentialExistFunc(true, contextName)
	req.NoError(err)
	req.Equal(exist, false)
	verifyLoggedOutState(t, cfg, contextName)
}

func TestRemoveNetrcCredentials(t *testing.T) {
	req := require.New(t)
	clearCCloudDeprecatedEnvVar(req)
	cfg := config.AuthenticatedCloudConfigMock()
	contextName := cfg.Context().GetNetrcMachineName()
	// run login command
	auth := &ccloudv1mock.Auth{
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
	userInterface := &ccloudv1mock.UserInterface{}
	loginCmd, _ := newLoginCmd(auth, userInterface, true, req, mockNetrcHandler, AuthTokenHandler, mockLoginCredentialsManager, LoginOrganizationManager)
	_, err := pcmd.ExecuteCommand(loginCmd)
	req.NoError(err)

	_, err = mockNetrcHandler.RemoveNetrcCredentials(true, contextName)
	req.NoError(err)
	exist, err := mockNetrcHandler.CheckCredentialExistFunc(true, contextName)
	req.NoError(err)
	req.Equal(exist, false)
}

func newLogoutCmd(auth *ccloudv1mock.Auth, userInterface *ccloudv1mock.UserInterface, isCloud bool, req *require.Assertions, netrcHandler netrc.NetrcHandler, authTokenHandler pauth.AuthTokenHandler, loginCredentialsManager pauth.LoginCredentialsManager, loginOrganizationManager pauth.LoginOrganizationManager, contextName string) (*cobra.Command, *config.Config) {
	config.SetTempHomeDir()
	cfg := config.AuthenticatedConfigMockWithContextName(contextName)
	var mdsClient *mdsv1.APIClient
	if !isCloud {
		mdsConfig := mdsv1.NewConfiguration()
		mdsClient = mdsv1.NewAPIClient(mdsConfig)
		mdsClient.TokensAndAuthenticationApi = &mdsmock.TokensAndAuthenticationApi{
			GetTokenFunc: func(ctx context.Context) (mdsv1.AuthenticationResponse, *http.Response, error) {
				return mdsv1.AuthenticationResponse{
					AuthToken: testToken1,
					TokenType: "JWT",
					ExpiresIn: 100,
				}, nil, nil
			},
		}
	}
	ccloudClientFactory := &climock.CCloudClientFactory{
		AnonHTTPClientFactoryFunc: func(baseURL string) *ccloudv1.Client {
			req.Equal("https://confluent.cloud", baseURL)
			return &ccloudv1.Client{Params: &ccloudv1.Params{HttpClient: new(http.Client)}, Auth: auth, User: userInterface}
		},
		JwtHTTPClientFactoryFunc: func(ctx context.Context, jwt, baseURL string) *ccloudv1.Client {
			return &ccloudv1.Client{Growth: &ccloudv1mock.Growth{
				GetFreeTrialInfoFunc: func(_ int32) ([]*ccloudv1.GrowthPromoCodeClaim, error) {
					return []*ccloudv1.GrowthPromoCodeClaim{}, nil
				},
			}, Auth: auth, User: userInterface}
		},
	}
	mdsClientManager := &climock.MDSClientManager{
		GetMDSClientFunc: func(_, _ string, _ bool) (*mdsv1.APIClient, error) {
			return mdsClient, nil
		},
	}
	prerunner := climock.NewPreRunnerMock(ccloudClientFactory.AnonHTTPClientFactory(ccloudURL), nil, mdsClient, nil, cfg)
	logoutCmd := NewLogoutCommand(cfg, prerunner, ccloudClientFactory, mdsClientManager, netrcHandler, loginCredentialsManager, loginOrganizationManager, authTokenHandler)
	return logoutCmd, cfg
}

func verifyLoggedOutState(t *testing.T, cfg *config.Config, loggedOutContext string) {
	req := require.New(t)
	fmt.Println(loggedOutContext)
	state := cfg.Contexts[loggedOutContext].State
	req.Empty(state.AuthToken)
	req.Empty(state.Auth)
}

func clearCCloudDeprecatedEnvVar(req *require.Assertions) {
	req.NoError(os.Unsetenv(pauth.DeprecatedConfluentCloudEmail))
}
